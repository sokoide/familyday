package domain

import "fmt"

// StageID は4ステージの識別子。
type StageID string

const (
	StageRiver  StageID = "stage1" // 川(渡河)
	StageGate   StageID = "stage2" // 城門(ゴーレム)
	StageFire   StageID = "stage3" // 炎のトラップ
	StageDragon StageID = "stage4" // ドラゴンとの最終決戦
)

func ParseStageID(s string) (StageID, error) {
	switch StageID(s) {
	case StageRiver, StageGate, StageFire, StageDragon:
		return StageID(s), nil
	}
	return "", fmt.Errorf("%w: unknown stage %q", ErrInvalidInput, s)
}

// Next は次ステージ。最終ステージは空文字(=エンディングへ)。
func (s StageID) Next() StageID {
	switch s {
	case StageRiver:
		return StageGate
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
	NeedsDragonRoute bool   // stage4(ドラゴン)は route 判定が必要
}
