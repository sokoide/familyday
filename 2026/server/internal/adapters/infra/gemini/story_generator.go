package gemini

import (
	"context"
	"fmt"
	"strings"

	"github.com/sokoide/familyday/server/internal/domain"
	"github.com/sokoide/familyday/server/internal/usecase"
	"google.golang.org/genai"
)

// StoryGenerator はエンディング毎のストーリー文を生成する。
type StoryGenerator struct {
	c *Client
}

func NewStoryGenerator(c *Client) *StoryGenerator { return &StoryGenerator{c: c} }

func (s *StoryGenerator) Generate(ctx context.Context, input usecase.StoryInput) (string, error) {
	cfg := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{Parts: []*genai.Part{{Text: storySystemPrompt(input.Lang)}}},
		SafetySettings:    safetySettings(),
		ThinkingConfig:    &genai.ThinkingConfig{ThinkingBudget: thinkingBudgetTokens()},
	}
	contents := []*genai.Content{
		{Role: "user", Parts: []*genai.Part{{Text: storyPrompt(input)}}},
	}
	resp, err := s.c.client.Models.GenerateContent(ctx, s.c.cfg.ModelStory, contents, cfg)
	if err != nil {
		return "", fmt.Errorf("%w: gemini story: %v", domain.ErrUpstream, err)
	}
	text, err := firstText(resp)
	if err != nil {
		return "", fmt.Errorf("%w: gemini story empty: %v", domain.ErrUpstream, err)
	}
	return strings.TrimSpace(text), nil
}
