package domain

// Ending は1プレイの確定結果。ストーリーと共に永続化される。
// 永続の詳細(ファイル/DB)は知らない純粋なドメインオブジェクト。
type Ending struct {
	ID         string
	EndingType EndingType
	Lives      Lives
	Route      DragonRoute
	Story      string // LLM 生成のストーリー文
	CreatedAt  string // ISO8601(タイムスタンプは外側から注入)
}
