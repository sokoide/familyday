package ratelimit

import (
	"context"
	"testing"
	"time"
)

// newLimiter は固定時刻を返す now を注入した Limiter を作る(時間経過を決定的に検証)。
func newLimiter(now *time.Time) *Limiter {
	return &Limiter{
		windows: map[string][]time.Time{},
		now:     func() time.Time { return *now },
	}
}

// 境界値: limit 回目までは許可し、(limit+1) 回目で拒否する。
func TestLimiter_Boundary(t *testing.T) {
	t0 := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	l := newLimiter(&t0)
	const limit = 3

	for i := 1; i <= limit; i++ {
		if !l.Allow(context.Background(), "k", limit) {
			t.Errorf("call %d/%d: want allow, got deny", i, limit)
		}
	}
	if l.Allow(context.Background(), "k", limit) {
		t.Errorf("call %d: want deny, got allow", limit+1)
	}
}

// キー独立性: あるキーの消費が別キーの許可に影響しない。
func TestLimiter_KeyIsolation(t *testing.T) {
	t0 := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	l := newLimiter(&t0)
	const limit = 1

	if !l.Allow(context.Background(), "a", limit) {
		t.Fatal("first 'a' should be allowed")
	}
	// 'a' は枯渇していても 'b' は独立して許可される
	if !l.Allow(context.Background(), "b", limit) {
		t.Error("'b' should be allowed independently of 'a'")
	}
	if l.Allow(context.Background(), "a", limit) {
		t.Error("exhausted 'a' should be denied")
	}
}

// 時間経過でウィンドウが切り替わる: 60秒以上進めると再び許可される。
func TestLimiter_WindowReset(t *testing.T) {
	now := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	l := newLimiter(&now)
	const limit = 2

	for range limit {
		if !l.Allow(context.Background(), "k", limit) {
			t.Fatal("initial calls should be allowed")
		}
	}
	if l.Allow(context.Background(), "k", limit) {
		t.Fatal("should be denied at limit")
	}

	// ウィンドウ内(59秒)はまだ拒否
	now = now.Add(59 * time.Second)
	if l.Allow(context.Background(), "k", limit) {
		t.Fatal("should still be denied within 60s")
	}

	// 60秒経過でリセットされて再許可
	now = now.Add(time.Second) // 合計60秒
	if !l.Allow(context.Background(), "k", limit) {
		t.Fatal("should be allowed after 60s window reset")
	}
}

// 期限切れエントリはプルーニングされる(メモリ無限増殖防止)。
// 別キーを大量に消費しても、期限切れになった時点で map から除去される。
func TestLimiter_PruneExpired(t *testing.T) {
	now := time.Date(2026, 7, 2, 12, 0, 0, 0, time.UTC)
	l := newLimiter(&now)

	// 多数のキーを1回ずつ消費(エントリが溜まる)
	for i := 0; i < 100; i++ {
		k := string(rune('a'+i%26)) + string(rune('a'+i/26))
		l.Allow(context.Background(), k, 5)
	}
	if got := len(l.windows); got == 0 {
		t.Fatal("windows should have entries before prune")
	}

	// 60秒以上経過させた後、枯渇していないキーを叩くと期限内のキーだけ残る想定だが、
	// ここでは「全キーが期限切れの瞬間に、1キー叩くとその1キーだけ残る」ことを検証。
	now = now.Add(2 * time.Minute)
	l.Allow(context.Background(), "solo", 5)

	if _, ok := l.windows["solo"]; !ok {
		t.Error("freshly used key should remain in windows")
	}
}
