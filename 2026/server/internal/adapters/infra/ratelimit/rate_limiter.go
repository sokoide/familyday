// Package ratelimit はメモリベースの簡易レートリミット adapter。
package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter は sliding-window 風のメモリカウンタ(分単位)。
type Limiter struct {
	mu       sync.Mutex
	windows  map[string][]time.Time
	now      func() time.Time
}

func New() *Limiter {
	return &Limiter{
		windows: map[string][]time.Time{},
		now:     time.Now,
	}
}

// Allow は直近60秒のカウントが limit 未満なら許可し、カウントを進める。
// ウィンドウが空になったキーは map から破棄し、メモリ無限増殖を防ぐ。
func (l *Limiter) Allow(ctx context.Context, key string, limitPerMinute int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	cutoff := l.now().Add(-60 * time.Second)
	hits := l.windows[key]
	// 期限切れを破棄(インプレースフィルタ: 書込位置 ≦ 読取位置なので安全)
	pruned := hits[:0]
	for _, t := range hits {
		if t.After(cutoff) {
			pruned = append(pruned, t)
		}
	}

	if len(pruned) >= limitPerMinute {
		l.windows[key] = pruned
		return false
	}
	pruned = append(pruned, l.now())
	if len(pruned) == 0 {
		delete(l.windows, key)
	} else {
		l.windows[key] = pruned
	}
	return true
}
