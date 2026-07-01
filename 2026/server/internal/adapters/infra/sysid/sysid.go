// Package sysid は ID生成・時刻取得の adapter(usecase ポートを満たす)。
package sysid

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// UUIDGen は crypto/rand 由来の予測困難なIDを生成する。
type UUIDGen struct{}

func (UUIDGen) NewID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// rand.Read 失敗は極めて稀。フォールバックは時刻ベース。
		return hex.EncodeToString([]byte(time.Now().Format("20060102150405000000000")))
	}
	return hex.EncodeToString(b[:]) // 32文字
}

// SystemClock は実時刻を ISO8601(UTC) で返す。
type SystemClock struct{}

func (SystemClock) NowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}
