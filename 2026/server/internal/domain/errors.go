// Package domain はビジネスルールの中核。
// 技術(net/http, SDK, ORM 等)に依存しない純粋なドメイン。
package domain

import "errors"

// ドメインエラー。意味だけを横断し、Presentation 層が transport エラーに変換する。
var (
	ErrInvalidInput = errors.New("invalid input")
	ErrRateLimited  = errors.New("rate limited")
	ErrUpstream     = errors.New("upstream error")
	ErrNotFound     = errors.New("not found")
)
