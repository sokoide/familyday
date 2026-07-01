package gemini

import (
	"context"
	"fmt"
	"strings"

	"github.com/sokoide/familyday/server/internal/domain"
	"google.golang.org/genai"
)

// StoryGenerator はエンディング毎のストーリー文を生成する。
type StoryGenerator struct {
	c *Client
}

func NewStoryGenerator(c *Client) *StoryGenerator { return &StoryGenerator{c: c} }

func (s *StoryGenerator) Generate(ctx context.Context, t domain.EndingType, _ domain.Lives, _ domain.DragonRoute, lang domain.Lang) (string, error) {
	cfg := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: "あなたは子供向け絵本の文を書く作家です。"}}},
		SafetySettings:    safetySettings(),
		ThinkingConfig:    &genai.ThinkingConfig{ThinkingBudget: thinkingBudgetTokens()},
	}
	contents := []*genai.Content{
		{Role: "user", Parts: []*genai.Part{{Text: storyPrompt(t, lang)}}},
	}
	resp, err := s.c.client.Models.GenerateContent(ctx, s.c.cfg.ModelStory, contents, cfg)
	if err != nil {
		return "", fmt.Errorf("gemini story: %w", err)
	}
	text, err := firstText(resp)
	if err != nil {
		return "", fmt.Errorf("gemini story empty: %w", err)
	}
	return strings.TrimSpace(text), nil
}
