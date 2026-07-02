package gemini

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"

	"github.com/sokoide/familyday/server/internal/domain"
	"github.com/sokoide/familyday/server/internal/usecase"
	"golang.org/x/image/draw"
	"google.golang.org/genai"
)

// ImageGenerator はエンディング画像を1枚生成する ImageGenerator 実装。
// Nano Banana Lite(gemini-3.1-flash-lite-image)を generate_content で呼ぶ。
// (Imagen 系は 2026-08-17 に廃止のため generate_images から移行)
type ImageGenerator struct {
	c *Client
}

func NewImageGenerator(c *Client) *ImageGenerator { return &ImageGenerator{c: c} }

func (g *ImageGenerator) Generate(ctx context.Context, t domain.EndingType, route domain.DragonRoute) (usecase.Image, error) {
	prompt := imagenPrompt(t, route)
	cfg := &genai.GenerateContentConfig{
		// 画像生成モデルには IMAGE を要求(テキストも併せて返しうる)
		ResponseModalities: []string{"IMAGE", "TEXT"},
		SafetySettings:     safetySettings(),
	}
	contents := []*genai.Content{
		{Role: "user", Parts: []*genai.Part{{Text: prompt}}},
	}

	// ImageCount 回生成を試行し、最初の成功画像を採用(非バッチ既定=1)。
	// Nano Banana は generate_content 1回で1枚を返すため、N>1 は複数回呼んで最初の成功を採用する。
	// (枚数を増やすと生成コストが N 倍になる点に注意。env GEMINI_IMAGE_COUNT で制御。)
	var lastErr error
	for i := 0; i < max1(g.c.cfg.ImageCount); i++ {
		resp, err := g.c.client.Models.GenerateContent(ctx, g.c.cfg.ModelImage, contents, cfg)
		if err != nil {
			lastErr = fmt.Errorf("image generate: %w", err)
			continue
		}
		blob, err := firstInlineImage(resp)
		if err != nil {
			lastErr = fmt.Errorf("image generate: %w", err)
			continue
		}
		data := blob.Data
		// ImageSize が 0/未設定でなく、1024(既定)以外ならサーバ側でリサイズ。
		if size := g.c.cfg.ImageSize; size > 0 {
			if resized, err := resizeSquare(data, size); err == nil {
				data = resized
			}
			// リサイズ失敗時は元画像をそのまま返す(体験優先)
		}
		return usecase.Image{Bytes: data, MIME: "image/png"}, nil
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("image generate: no image returned")
	}
	return usecase.Image{}, lastErr
}

// resizeSquare は PNG バイト列を size×size の正方形にリサイズして PNG で返す。
// 高品質ダウンスケールに CatmullRom を使用。アップスケールも可能だが推奨しない。
func resizeSquare(pngBytes []byte, size int) ([]byte, error) {
	src, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		return nil, fmt.Errorf("png decode: %w", err)
	}
	dst := image.NewRGBA(image.Rect(0, 0, size, size))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	var buf bytes.Buffer
	if err := png.Encode(&buf, dst); err != nil {
		return nil, fmt.Errorf("png encode: %w", err)
	}
	return buf.Bytes(), nil
}

// max1 は n を 1以上に正規化(0や負の誤設定を防ぐ)。
func max1(n int) int {
	if n < 1 {
		return 1
	}
	return n
}
