package domain

import "fmt"

// Lives は残ライフ(0..3)。不変の値オブジェクト。
const MaxLives = 3

type Lives int

// NewLives は範囲チェック付きコンストラクタ。
func NewLives(n int) (Lives, error) {
	if n < 0 || n > MaxLives {
		return 0, fmt.Errorf("%w: lives out of range %d", ErrInvalidInput, n)
	}
	return Lives(n), nil
}

// Apply は判定による変化を適用し、0..MaxLives に収まる新しい Lives を返す。
func (l Lives) Apply(delta int) Lives {
	v := int(l) + delta
	return Lives(max(0, min(MaxLives, v)))
}

// Dead はライフ切れ(ゲームオーバー)か。
func (l Lives) Dead() bool { return l <= 0 }

// Full は満ライフか。
func (l Lives) Full() bool { return l >= MaxLives }
