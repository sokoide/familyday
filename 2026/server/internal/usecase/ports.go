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

// LLMJudgeGateway はステージ判定を行う外部 LLM の能力ポート。
// (vendor 名を含めず capability 指向 — Gemini 実装は infra 側)
type LLMJudgeGateway interface {
	Judge(ctx context.Context, stage domain.Stage, input string, lang domain.Lang) (JudgeResult, error)
}

// StoryGenerator はエンディング毎のストーリー文を生成する。
type StoryGenerator interface {
	Generate(ctx context.Context, t domain.EndingType, lives domain.Lives, route domain.DragonRoute, lang domain.Lang) (string, error)
}

// Image は生成された画像のバイト列と MIME。
type Image struct {
	Bytes []byte
	MIME  string // "image/png"
}

// ImageGenerator はエンディング画像を1枚生成する。
type ImageGenerator interface {
	Generate(ctx context.Context, t domain.EndingType, route domain.DragonRoute) (Image, error)
}

// EndingRepository はエンディング結果の永続化を行うポート。
// img が空(Bytes 空)の場合はメタデータのみ保存し、Presentation が fallback 画像を指す。
type EndingRepository interface {
	Save(ctx context.Context, e domain.Ending, img Image) error
	Load(ctx context.Context, id string) (domain.Ending, error)
}

// RateLimiter はキー単位の簡易レートリミット。
type RateLimiter interface {
	Allow(ctx context.Context, key string, limitPerMinute int) bool
}

// IDGenerator はエンディングID/画像ファイル名用の乱数由来文字列生成。
// (Math/rand を domain/usecase に持ち込まないためポート化)
type IDGenerator interface {
	NewID() string // 予測困難な16バイト級文字列
}

// Clock は現在時刻(ISO8601)を返す。
type Clock interface {
	NowISO() string
}

// Logger は usecase 層からの診断ログ用ポート(副作用)。infra が実装する。
// 画像生成失敗など、上位へ伝播させないが運用で見たい事象を記録する。
type Logger interface {
	Printf(format string, args ...any)
}

// NopLogger は何も出力しない Logger(テスト用既定値)。
type NopLogger struct{}

func (NopLogger) Printf(string, ...any) {}
