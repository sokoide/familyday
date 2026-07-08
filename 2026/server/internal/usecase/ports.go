// Package usecase はアプリケーションのワークフローとポートを定義する。
// ポート(インタフェース)はこの層が所有し、Infrastructure Adapter が実装する。
package usecase

import (
	"context"

	"github.com/sokoide/familyday/server/internal/domain"
)

// JudgeResult は LLM が返す生の判定。ドメイン値に変換済み。
type JudgeResult struct {
	Verdict domain.Verdict
	Route   domain.DragonRoute
	Message string
	Reason  string // 判定理由(子供向け)。ライフ変化の根拠を示す
}

// AdventureEvent はエンディング要約の材料になる1手の記録。
type AdventureEvent struct {
	StageIndex int
	Spoken     string
	Verdict    string
	Reason     string
}

// LLMJudgeGateway はステージ判定を行う外部 LLM の能力ポート。
// (vendor 名を含めず capability 指向 — Gemini 実装は infra 側)
type LLMJudgeGateway interface {
	Judge(ctx context.Context, stage domain.Stage, input string, lang domain.Lang) (JudgeResult, error)
}

// StoryInput はエンディング要約生成の入力。
type StoryInput struct {
	EndingType domain.EndingType
	Lives      domain.Lives
	Route      domain.DragonRoute
	Lang       domain.Lang
	History    []AdventureEvent
}

// StoryGenerator は冒険履歴からエンディング要約を生成する外部 LLM の能力ポート。
type StoryGenerator interface {
	Generate(ctx context.Context, input StoryInput) (string, error)
}

// EndingRepository はエンディング結果の永続化を行うポート。
type EndingRepository interface {
	Save(ctx context.Context, e domain.Ending) error
	Load(ctx context.Context, id string) (domain.Ending, error)
}

// RateLimiter はキー単位の簡易レートリミット。
type RateLimiter interface {
	Allow(ctx context.Context, key string, limitPerMinute int) bool
}

// IDGenerator はエンディングID/画像ファイル名用の乱数由来文字列生成。
// (Math/rand を domain/usecase に持ち込まないためポート化)
type IDGenerator interface {
	NewID() (string, error) // 予測困難な16バイト級文字列
}

// Clock は現在時刻(ISO8601)を返す。
type Clock interface {
	NowISO() string
}

// Logger は usecase 層からの診断ログ用ポート(副作用)。infra が実装する。
type Logger interface {
	Printf(format string, args ...any)
}

// NopLogger は何も出力しない Logger(テスト用既定値)。
type NopLogger struct{}

func (NopLogger) Printf(string, ...any) {}
