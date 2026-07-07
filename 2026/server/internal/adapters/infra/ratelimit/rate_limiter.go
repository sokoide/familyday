// Package ratelimit はメモリベースの簡易レートリミット adapter。
package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter は sliding-window 風のメモリカウンタ(分単位)。
type Limiter struct {
	mu        sync.Mutex
	windows   map[string][]time.Time
	now       func() time.Time
	lastSweep time.Time // 前回の全キー掃除時刻
}

func New() *Limiter {
	n := time.Now
	return &Limiter{
		windows:   map[string][]time.Time{},
		now:       n,
		lastSweep: n(),
	}
}

// Allow は直近60秒のカウントが limit 未満なら許可し、カウントを進める。
// 期限内ヒットが無くなったキーは map から破棄し、メモリ無限増殖を防ぐ。
func (l *Limiter) Allow(ctx context.Context, key string, limitPerMinute int) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 周期的に全キーを掃除する(自キー掃除だけでは未参照の dead key が残り、
	// sessionId 単位で map が無限増殖するのを防ぐ)。Allow 呼び出しに償却され、
	// 60s につき最大1回しか走らないのでレイテンシ影響も小さい。
	if l.now().Sub(l.lastSweep) >= 60*time.Second {
		l.sweepLocked()
	}

	cutoff := l.now().Add(-60 * time.Second)
	hits := l.windows[key]
	// 期限切れを破棄(インプレースフィルタ: 書込位置 ≦ 読取位置なので安全)
	pruned := hits[:0]
	for _, t := range hits {
		if t.After(cutoff) {
			pruned = append(pruned, t)
		}
	}

	// 期限内ヒットが無くなったら即座に破棄(以降の枝で使い回さない)
	if len(pruned) == 0 {
		delete(l.windows, key)
		if limitPerMinute <= 0 {
			// limit 0 = 常に拒否。カウントは進めない。
			return false
		}
		// 新規キー相当として 1 件目を記録して許可
		l.windows[key] = []time.Time{l.now()}
		return true
	}

	if len(pruned) >= limitPerMinute {
		// 枯渇: 期限内の有効ヒットのみ保存して拒否。
		// 次回 Allow で期限内が 0 になれば delete され、リークしない。
		l.windows[key] = pruned
		return false
	}
	pruned = append(pruned, l.now())
	l.windows[key] = pruned
	return true
}

// sweepLocked は全キーを走査し、期限内ヒットが無くなったキーを map から破棄する。
// l.mu の保護下で呼ぶこと。
func (l *Limiter) sweepLocked() {
	cutoff := l.now().Add(-60 * time.Second)
	for k, hits := range l.windows {
		pruned := hits[:0]
		for _, t := range hits {
			if t.After(cutoff) {
				pruned = append(pruned, t)
			}
		}
		if len(pruned) == 0 {
			delete(l.windows, k)
		} else {
			l.windows[k] = pruned
		}
	}
	l.lastSweep = l.now()
}
