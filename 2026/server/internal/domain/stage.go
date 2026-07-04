package domain

import "fmt"

// StageID は3ステージの識別子。
type StageID string

const (
	StageGate   StageID = "stage1" // 城門(ゴーレム)
	StageFire   StageID = "stage2" // 炎のトラップ
	StageDragon StageID = "stage3" // ドラゴンとの最終決戦
)

func ParseStageID(s string) (StageID, error) {
	switch StageID(s) {
	case StageGate, StageFire, StageDragon:
		return StageID(s), nil
	}
	return "", fmt.Errorf("%w: unknown stage %q", ErrInvalidInput, s)
}

// Next は次ステージ。最終ステージは空文字(=エンディングへ)。
func (s StageID) Next() StageID {
	switch s {
	case StageGate:
		return StageFire
	case StageFire:
		return StageDragon
	}
	return ""
}

// IsLast は最終ステージか。
func (s StageID) IsLast() bool { return s == StageDragon }

// Stage は1ステージの描写メタ。成功条件は LLM 審判プロンプトに渡す基準。
// 画面表示用の文言は UI 側(messages.ts)が唯一の正(二重管理を避けるため Goal フィールドは持たない)。
type Stage struct {
	ID               StageID
	Title            string // 表示名(プロンプト埋め込み用。UI 表示は messages.ts 側)
	Situation        string // 状況描写(プロンプト埋め込み用)
	SuccessSpec      string // 審判基準(LLM へ渡す成功条件の自然言語仕様)
	NeedsDragonRoute bool   // stage3 は route 判定が必要
}
