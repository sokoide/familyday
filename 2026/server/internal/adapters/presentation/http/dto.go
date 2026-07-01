// Package httpadapter は inbound HTTP の presentation adapter。
// リクエスト解析→UseCase呼出→レスポンスDTO生成・エラー→ステータス変換。
package httpadapter

// judge
type JudgeRequest struct {
	StageID   string `json:"stageId"`
	SessionID string `json:"sessionId"`
	Input     string `json:"input"`
	Lang      string `json:"lang"` // "ja"|"en"。応答メッセージの言語
}

type JudgeResponse struct {
	Verdict    string `json:"verdict"`     // Great|Good|Bad
	Route      string `json:"route"`       // stage3 のみ defeat|befriend、他は ""
	Message    string `json:"message"`
	LivesDelta int    `json:"livesDelta"`
	Advance    bool   `json:"advance"`
}

// ending
type EndingRequest struct {
	Lives       int    `json:"lives"`
	FinalAction string `json:"finalAction"` // defeat|befriend|gameover
	Cleared     bool   `json:"cleared"`
	SessionID   string `json:"sessionId"`
	Lang        string `json:"lang"` // "ja"|"en"。ストーリーの言語
}

type EndingResponse struct {
	EndingID   string `json:"endingId"`
	EndingType string `json:"endingType"`
	Story      string `json:"story"`
	ImageURL   string `json:"imageUrl"`
	ResultURL  string `json:"resultUrl"`
}

// result
type ResultResponse struct {
	EndingType string `json:"endingType"`
	Story      string `json:"story"`
	ImageURL   string `json:"imageUrl"`
	ResultURL  string `json:"resultUrl"`
	CreatedAt  string `json:"createdAt"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
