package gemini

import (
	"context"
	"fmt"

	"github.com/sokoide/familyday/server/internal/domain"
	"github.com/sokoide/familyday/server/internal/usecase"
	"google.golang.org/genai"
)

// ImagenGenerator はエンディング画像を1枚生成する ImageGenerator 実装。
type ImagenGenerator struct {
	c *Client
}

func NewImagenGenerator(c *Client) *ImagenGenerator { return &ImagenGenerator{c: c} }

func (g *ImagenGenerator) Generate(ctx context.Context, t domain.EndingType, route domain.DragonRoute) (usecase.Image, error) {
	cfg := &genai.GenerateImagesConfig{
		NumberOfImages:   1,
		AspectRatio:      "1:1",
		SafetyFilterLevel: genai.SafetyFilterLevelBlockOnlyHigh,
		OutputMIMEType:   "image/png",
	}
	prompt := imagenPrompt(t, route)
	resp, err := g.c.client.Models.GenerateImages(ctx, g.c.cfg.ModelImagen, prompt, cfg)
	if err != nil {
		return usecase.Image{}, fmt.Errorf("imagen generate: %w", err)
	}
	if resp == nil || len(resp.GeneratedImages) == 0 || resp.GeneratedImages[0].Image == nil {
		return usecase.Image{}, fmt.Errorf("imagen: no image returned")
	}
	img := resp.GeneratedImages[0].Image
	return usecase.Image{Bytes: img.ImageBytes, MIME: "image/png"}, nil
}
