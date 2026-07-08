// Package sysid は ID生成・時刻取得・ログ の adapter(usecase ポートを満たす)。
package sysid

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"time"
)

// UUIDGen は crypto/rand 由来の予測困難なIDを生成する。
type UUIDGen struct{}

func (UUIDGen) NewID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("crypto/rand unavailable: %w", err)
	}
	return hex.EncodeToString(b[:]), nil // 32文字
}

// SystemClock は実時刻を ISO8601(UTC) で返す。
type SystemClock struct{}

func (SystemClock) NowISO() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// StdLogger は標準 log パッケージで出力する Logger 実装。
type StdLogger struct{}

func (StdLogger) Printf(format string, args ...any) {
	log.Printf(format, args...)
}
