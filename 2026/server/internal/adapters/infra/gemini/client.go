package gemini

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sokoide/familyday/server/internal/domain"
	"github.com/sokoide/familyday/server/internal/usecase"
	"google.golang.org/genai"
)

// Client は Gemini への共通クライアントと設定を持つ。
type Client struct {
	client *genai.Client
	cfg    Config
}

func NewClient(ctx context.Context, apiKey string, cfg Config) (*Client, error) {
	if apiKey == "" {
		return nil, errors.New("gemini: empty API key")
	}
	c, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, fmt.Errorf("gemini: new client: %w", err)
	}
	return &Client{client: c, cfg: cfg}, nil
}

// --- LLMJudgeGateway 実装 ---

type JudgeGateway struct {
	c *Client
}

func NewJudgeGateway(c *Client) *JudgeGateway { return &JudgeGateway{c: c} }

func (g *JudgeGateway) Judge(ctx context.Context, stage domain.Stage, input string, lang domain.Lang) (usecase.JudgeResult, error) {
	cfg := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: judgeSystemPrompt(stage, lang)}}},
		ResponseMIMEType:  "application/json",
		ResponseSchema:    judgeSchema(stage.NeedsDragonRoute),
		SafetySettings:    safetySettings(),
		ThinkingConfig:    &genai.ThinkingConfig{ThinkingBudget: thinkingBudgetTokens()},
	}
	contents := []*genai.Content{
		{Role: "user", Parts: []*genai.Part{{Text: input}}},
	}
	resp, err := g.c.client.Models.GenerateContent(ctx, g.c.cfg.ModelJudge, contents, cfg)
	if err != nil {
		// 通信・認証エラーは本物の upstream 障害
		return usecase.JudgeResult{}, fmt.Errorf("%w: gemini judge: %v", domain.ErrUpstream, err)
	}
	text, err := firstText(resp)
	if err != nil {
		// セーフティブロックや空応答は「Bad(リトライ)」に倒す。
		// 子供が不適切語を言った時などにエラー画面ではなく「もう一度」で体験を継続。
		return usecase.JudgeResult{
			Verdict: domain.VerdictBad,
			Route:   domain.RouteNone,
			Message: blockedMessage,
		}, nil
	}

	var j judgeJSON
	if err := json.Unmarshal([]byte(text), &j); err != nil {
		return usecase.JudgeResult{}, fmt.Errorf("%w: parse judge json: %v", domain.ErrUpstream, err)
	}
	v, err := domain.ParseVerdict(j.Verdict)
	if err != nil {
		return usecase.JudgeResult{}, err
	}
	route := domain.RouteNone
	if stage.NeedsDragonRoute && j.Route != "" {
		route, err = domain.ParseDragonRoute(j.Route)
		if err != nil {
			route = domain.RouteDefeat // 不正時は安全側に倒す
		}
	}
	return usecase.JudgeResult{Verdict: v, Route: route, Message: j.Message}, nil
}

// firstText はレスポンスの最初のテキストを取り出す。空なら error。
func firstText(resp *genai.GenerateContentResponse) (string, error) {
	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return "", errors.New("no candidates")
	}
	for _, p := range resp.Candidates[0].Content.Parts {
		if p.Text != "" {
			return p.Text, nil
		}
	}
	return "", errors.New("no text part")
}

// firstInlineImage はレスポンスの最初のインライン画像(Blob)を取り出す。無ければ error。
// Nano Banana 系画像生成モデルは画像を content parts の InlineData で返す。
func firstInlineImage(resp *genai.GenerateContentResponse) (*genai.Blob, error) {
	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return nil, errors.New("no candidates")
	}
	for _, p := range resp.Candidates[0].Content.Parts {
		if p.InlineData != nil && len(p.InlineData.Data) > 0 {
			return p.InlineData, nil
		}
	}
	return nil, errors.New("no image part")
}
