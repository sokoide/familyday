package httpadapter

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"strings"
	"unicode/utf8"

	"github.com/sokoide/familyday/server/internal/domain"
	"github.com/sokoide/familyday/server/internal/usecase"
)

// Handler は UseCase と設定を保持する presentation adapter。
type Handler struct {
	judge   *usecase.JudgeUseCase
	ending  *usecase.EndingUseCase
	baseURL string // 絶対URL構築用(末尾スラッシュなし)
}

func NewHandler(j *usecase.JudgeUseCase, e *usecase.EndingUseCase, baseURL string) *Handler {
	baseURL = strings.TrimRight(baseURL, "/")
	return &Handler{judge: j, ending: e, baseURL: baseURL}
}

func (h *Handler) resultURL(id string) string {
	return h.baseURL + "/r/" + path.Base(id)
}

func (h *Handler) endingImageURL(t domain.EndingType) string {
	isClear := t == domain.EndingGreat || t == domain.EndingSuccess
	return h.baseURL + "/images/" + map[bool]string{true: "successful.jpg", false: "failed.jpg"}[isClear]
}

func (h *Handler) Judge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}
	var req JudgeRequest
	if err := decodeJSON(w, r, &req); err != nil {
		if errors.Is(err, errRequestTooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, "request too large", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}
	out, err := h.judge.Judge(r.Context(), usecase.JudgeInput{
		StageID:   req.StageID,
		SessionID: req.SessionID,
		Input:     req.Input,
		Lang:      domain.NormalizeLang(req.Lang),
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, JudgeResponse{
		Verdict:    string(out.Verdict),
		Route:      string(out.Route),
		Message:    out.Message,
		Reason:     out.Reason,
		LivesDelta: out.LivesDelta,
		Advance:    out.Advance,
	})
}

func (h *Handler) Ending(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}
	var req EndingRequest
	if err := decodeJSON(w, r, &req); err != nil {
		if errors.Is(err, errRequestTooLarge) {
			writeError(w, http.StatusRequestEntityTooLarge, "request too large", err.Error())
			return
		}
		writeError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}
	out, err := h.ending.Resolve(r.Context(), usecase.EndingInput{
		Lives:       req.Lives,
		FinalAction: req.FinalAction,
		Cleared:     req.Cleared,
		Lang:        domain.NormalizeLang(req.Lang),
		History: func() []usecase.AdventureEvent {
			// History の要素数・各フィールド長に上限を設け、悪意クライアントによる
			// プロンプト膨張・コスト増大を防ぐ。4ステージ分をカバーする上限。
			const maxHistory = 16
			const maxFieldLen = 300
			if len(req.History) == 0 {
				return nil
			}
			if len(req.History) > maxHistory {
				req.History = req.History[:maxHistory]
			}
			history := make([]usecase.AdventureEvent, 0, len(req.History))
			for _, item := range req.History {
				history = append(history, usecase.AdventureEvent{
					StageIndex: item.StageIndex,
					Spoken:     truncateRunes(item.Spoken, maxFieldLen),
					Verdict:    item.Verdict,
					Reason:     truncateRunes(item.Reason, maxFieldLen),
				})
			}
			return history
		}(),
	}, req.SessionID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, EndingResponse{
		EndingID:   out.EndingID,
		EndingType: string(out.EndingType),
		Story:      out.Story,
		ImageURL:   h.endingImageURL(out.EndingType),
		ResultURL:  h.resultURL(out.EndingID),
	})
}

func (h *Handler) Result(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/result/")
	id = strings.TrimSpace(id)
	if id == "" {
		writeError(w, http.StatusBadRequest, "missing id", "")
		return
	}
	e, err := h.ending.Load(r.Context(), id)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, ResultResponse{
		EndingType: string(e.EndingType),
		Story:      e.Story,
		ImageURL:   h.endingImageURL(e.EndingType),
		ResultURL:  h.resultURL(e.ID),
		CreatedAt:  e.CreatedAt,
	})
}

// --- helpers ---

var errRequestTooLarge = errors.New("request body too large")

// truncateRunes は s を最大 n ルーンに切り詰める。履歴フィールドの長さ制限用。
func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	runes := []rune(s)
	return string(runes[:n])
}

func decodeJSON(w http.ResponseWriter, r *http.Request, v any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 64<<10)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		if strings.Contains(err.Error(), "http: request body too large") {
			return errRequestTooLarge
		}
		return err
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// writeDomainError はドメインエラーを transport ステータスに変換(意味の境界変換)。
func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "INVALID_INPUT", err.Error())
	case errors.Is(err, domain.ErrRateLimited):
		writeError(w, http.StatusTooManyRequests, "RATE_LIMITED", err.Error())
	case errors.Is(err, domain.ErrNotFound):
		writeError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	case errors.Is(err, domain.ErrUpstream):
		writeError(w, http.StatusBadGateway, "UPSTREAM", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL", err.Error())
	}
}

func writeError(w http.ResponseWriter, status int, code, msg string) {
	writeJSON(w, status, ErrorResponse{Error: code, Message: msg})
}
