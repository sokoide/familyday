package domain

// Lang は応答(メッセージ・ストーリー)の言語。
type Lang string

const (
	LangJA Lang = "ja"
	LangEN Lang = "en"
)

// NormalizeLang は不正値をデフォルト(ja)に正規化。
func NormalizeLang(s string) Lang {
	if Lang(s) == LangEN {
		return LangEN
	}
	return LangJA
}

// Name はプロンプトで指示する自然言語名。
func (l Lang) Name() string {
	if l == LangEN {
		return "English"
	}
	return "日本語"
}
