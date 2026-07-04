package httpadapter

import (
	"encoding/json"
	"errors"
	"net/http"
	"path"
	"strings"

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

// imageUrl / resultUrl を絶対URLで構築。
func (h *Handler) imageURL(file string) string {
	if file == "" {
		return ""
	}
	return h.baseURL + "/img/" + path.Base(file)
}
func (h *Handler) resultURL(id string) string {
	return h.baseURL + "/r/" + path.Base(id)
}

func (h *Handler) Judge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed", "")
		return
	}
	var req JudgeRequest
	if err := decodeJSON(r, &req); err != nil {
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
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid json", err.Error())
		return
	}
	out, err := h.ending.Resolve(r.Context(), usecase.EndingInput{
		Lives:       req.Lives,
		FinalAction: req.FinalAction,
		Cleared:     req.Cleared,
		Lang:        domain.NormalizeLang(req.Lang),
	}, req.SessionID)
	if err != nil {
		writeDomainError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, EndingResponse{
		EndingID:   out.EndingID,
		EndingType: string(out.EndingType),
		Story:      out.Story,
		ImageURL:   h.imageURL(out.ImageFile),
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
		ImageURL:   h.imageURL(e.ImageFile),
		ResultURL:  h.resultURL(e.ID),
		CreatedAt:  e.CreatedAt,
	})
}

// --- helpers ---

func decodeJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
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
