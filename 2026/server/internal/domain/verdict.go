package domain

import "fmt"

// Verdict は LLM 判定の3値。子供向けゲームブックの審判結果。
type Verdict string

const (
	VerdictGreat Verdict = "Great" // 大成功: ノーダメージで次のステージへ
	VerdictGood Verdict = "Good"  // 成功: 次へ進むがライフ -1
	VerdictBad  Verdict = "Bad"   // 失敗: ライフ -1 で同ステージリトライ
)

func ParseVerdict(s string) (Verdict, error) {
	switch Verdict(s) {
	case VerdictGreat, VerdictGood, VerdictBad:
		return Verdict(s), nil
	}
	return "", fmt.Errorf("%w: unknown verdict %q", ErrInvalidInput, s)
}

// LivesDelta はこの判定による残ライフ変化量。
// Great=0, Good=-1, Bad=-1
func (v Verdict) LivesDelta() int {
	switch v {
	case VerdictGreat:
		return 0
	case VerdictGood, VerdictBad:
		return -1
	}
	return 0
}

// Advances は次ステージへ進めるか(リトライでないか)。
func (v Verdict) Advances() bool {
	return v == VerdictGreat || v == VerdictGood
}
