package gemini

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

// makePNG は size×size の合成 PNG を返す(ネットワーク不要)。
func makePNG(t *testing.T, size int) []byte {
	t.Helper()
	m := image.NewRGBA(image.Rect(0, 0, size, size))
	for y := range size {
		for x := range size {
			m.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 128, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, m); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestResizeSquare_Downscale(t *testing.T) {
	src := makePNG(t, 1024)
	out, err := resizeSquare(src, 512)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	b := img.Bounds()
	if b.Dx() != 512 || b.Dy() != 512 {
		t.Errorf("want 512x512, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestResizeSquare_Upscale(t *testing.T) {
	src := makePNG(t, 256)
	out, err := resizeSquare(src, 512)
	if err != nil {
		t.Fatal(err)
	}
	img, _ := png.Decode(bytes.NewReader(out))
	b := img.Bounds()
	if b.Dx() != 512 || b.Dy() != 512 {
		t.Errorf("want 512x512, got %dx%d", b.Dx(), b.Dy())
	}
}

func TestResizeSquare_InvalidPNG(t *testing.T) {
	if _, err := resizeSquare([]byte("not a png"), 512); err == nil {
		t.Error("want error for invalid png")
	}
}

func TestDefaultConfig(t *testing.T) {
	c := DefaultConfig()
	if c.ImageSize != 1024 {
		t.Errorf("ImageSize=%d want 1024", c.ImageSize)
	}
	if c.ImageCount != 1 {
		t.Errorf("ImageCount=%d want 1 (non-batch)", c.ImageCount)
	}
	if c.ModelJudge != "gemini-3.1-flash-lite" || c.ModelStory != "gemini-3.1-flash-lite" {
		t.Errorf("judge/story model wrong: %s/%s", c.ModelJudge, c.ModelStory)
	}
	if c.ModelImage != "gemini-3.1-flash-lite-image" {
		t.Errorf("ModelImage=%s want gemini-3.1-flash-lite-image", c.ModelImage)
	}
}

func TestMax1(t *testing.T) {
	if max1(0) != 1 || max1(-3) != 1 || max1(5) != 5 {
		t.Error("max1 normalization wrong")
	}
}
