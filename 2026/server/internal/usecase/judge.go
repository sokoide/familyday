package usecase

import (
	"context"
	"strings"
	"unicode/utf8"

	"github.com/sokoide/familyday/server/internal/domain"
)

// JudgeInput は /api/judge の入力 DTO。
type JudgeInput struct {
	StageID   string
	SessionID string // レートリミットキー(空なら未制限扱いにはせず "anon" に正規化)
	Input     string
	Lang      domain.Lang
}

// JudgeOutput は判定結果。lives 更新はフロント権威だが、整合用に delta/advance も返す。
type JudgeOutput struct {
	Verdict    domain.Verdict
	Route      domain.DragonRoute // stage3 以外は空文字
	Message    string
	Reason     string // 判定理由(子供向け)
	LivesDelta int    // -1 / 0
	Advance    bool   // 次ステージへ進めるか
}

const (
	maxInputRunes = 200
	judgeLimit    = 30 // 1キー 30回/分
)

// JudgeUseCase は1ターンの判定を指揮する。
type JudgeUseCase struct {
	judge LLMJudgeGateway
	lim   RateLimiter
}

func NewJudgeUseCase(j LLMJudgeGateway, l RateLimiter) *JudgeUseCase {
	return &JudgeUseCase{judge: j, lim: l}
}

func (u *JudgeUseCase) Judge(ctx context.Context, in JudgeInput) (JudgeOutput, error) {
	input := strings.TrimSpace(in.Input)
	if input == "" {
		return JudgeOutput{}, domain.ErrInvalidInput
	}
	if utf8.RuneCountInString(input) > maxInputRunes {
		return JudgeOutput{}, domain.ErrInvalidInput
	}

	stageID, err := domain.ParseStageID(in.StageID)
	if err != nil {
		return JudgeOutput{}, err
	}
	stage, ok := domain.LookupStage(stageID)
	if !ok {
		return JudgeOutput{}, domain.ErrInvalidInput
	}

	key := rateKey("judge", in.SessionID)
	if !u.lim.Allow(ctx, key, judgeLimit) {
		return JudgeOutput{}, domain.ErrRateLimited
	}

	res, err := u.judge.Judge(ctx, stage, input, in.Lang)
	if err != nil {
		return JudgeOutput{}, err
	}

	return JudgeOutput{
		Verdict:    res.Verdict,
		Route:      res.Route,
		Message:    res.Message,
		Reason:     res.Reason,
		LivesDelta: res.Verdict.LivesDelta(),
		Advance:    res.Verdict.Advances(),
	}, nil
}

func rateKey(prefix, sessionID string) string {
	s := strings.TrimSpace(sessionID)
	if s == "" {
		s = "anon"
	}
	return prefix + ":" + s
}
