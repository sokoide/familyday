package domain

import "fmt"

// EndingType はエンディングの3分岐。
type EndingType string

const (
	EndingGreat    EndingType = "great"    // 🏆 伝説の勇者エンド
	EndingSuccess  EndingType = "success"  // ✨ がんばった勇者エンド
	EndingGameOver EndingType = "gameover" // 😢 また挑戦してねエンド
)

// DragonRoute はステージ3のクリア経路。エンディング分岐に使用。
type DragonRoute string

const (
	RouteDefeat   DragonRoute = "defeat"   // ドラゴンを打ち倒した
	RouteBefriend DragonRoute = "befriend" // ドラゴンと友だちになった
	RouteNone     DragonRoute = ""         // stage1/2 または未クリア
)

func ParseDragonRoute(s string) (DragonRoute, error) {
	switch DragonRoute(s) {
	case RouteDefeat, RouteBefriend, RouteNone:
		return DragonRoute(s), nil
	}
	return "", fmt.Errorf("%w: unknown dragon route %q", ErrInvalidInput, s)
}

// DecideEnding は「残ライフ・クリア有無・ドラゴン経路」からエンディング種を決定する純粋ルール。
//   - ライフ0、またはクリアできていない → gameover
//   - 満ライフ、またはドラゴンと友好 → great
//   - それ以外(1〜2 でクリア) → success
func DecideEnding(lives Lives, cleared bool, route DragonRoute) EndingType {
	if !cleared || lives.Dead() {
		return EndingGameOver
	}
	if lives.Full() || route == RouteBefriend {
		return EndingGreat
	}
	return EndingSuccess
}
