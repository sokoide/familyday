//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"

	"github.com/sokoide/familyday/server/internal/adapters/infra/gemini"
	"github.com/sokoide/familyday/server/internal/domain"
)

// TestIntegration_ImageDirect は ImageGenerator を直接呼ぶ。
// usecase 層が画像エラーを握り潰すため、infra 単体で原因を特定するのが目的。
// 無料ティア等でクォータ不足(429)の場合は skip とする(コードの問題ではない)。
func TestIntegration_ImageDirect(t *testing.T) {
	apiKey := requireAPIKey(t)
	ctx := context.Background()
	c, err := gemini.NewClient(ctx, apiKey, gemini.DefaultConfig())
	if err != nil {
		t.Fatal(err)
	}
	g := gemini.NewImageGenerator(c)
	img, err := g.Generate(ctx, domain.EndingGreat, domain.RouteBefriend)
	if err != nil {
		msg := err.Error()
		// 429 / quota / RESOURCE_EXHAUSTED は課金プランの問題なのでコード側の不具合ではない
		if strings.Contains(msg, "429") || strings.Contains(msg, "RESOURCE_EXHAUSTED") || strings.Contains(msg, "quota") {
			t.Skipf("image model quota exceeded (likely free-tier limit=0): %v", err)
		}
		t.Fatalf("image generate failed: %v", err)
	}
	if len(img.Bytes) == 0 {
		t.Fatal("image bytes empty")
	}
	t.Logf("OK: %d bytes, mime=%s", len(img.Bytes), img.MIME)
}
