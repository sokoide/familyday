package persistence

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sokoide/familyday/server/internal/domain"
	"github.com/sokoide/familyday/server/internal/usecase"
)

// newRepo は tmp 配下に endings/generated を作った repo を返す(決定性のため t.TempDir 使用)。
func newRepo(t *testing.T) (*EndingRepo, string, string) {
	t.Helper()
	root := t.TempDir()
	endingsDir := filepath.Join(root, "endings")
	imageDir := filepath.Join(root, "generated")
	repo, err := NewEndingRepo(endingsDir, imageDir)
	if err != nil {
		t.Fatalf("NewEndingRepo: %v", err)
	}
	return repo, endingsDir, imageDir
}

func sampleEnding() domain.Ending {
	return domain.Ending{
		ID:         "abc123",
		EndingType: domain.EndingGreat,
		Lives:      domain.Lives(3),
		Route:      domain.RouteBefriend,
		Story:      "勇者はドラゴンと仲良くなった!",
		ImageFile:  "abc123.png",
		CreatedAt:  "2026-07-02T00:00:00Z",
	}
}

func TestEndingRepo_SaveLoad_RoundTrip(t *testing.T) {
	repo, endingsDir, imageDir := newRepo(t)
	in := sampleEnding()
	img := usecase.Image{Bytes: []byte("\x89PNG\r\n\x1a\n FAKE"), MIME: "image/png"}

	if err := repo.Save(t.Context(), in, img); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// 画像とメタがそれぞれ正しいディレクトリ・ファイル名で書かれている
	got, err := os.ReadFile(filepath.Join(imageDir, "abc123.png"))
	if err != nil {
		t.Fatalf("image not written: %v", err)
	}
	if !bytes.Equal(got, img.Bytes) {
		t.Errorf("image bytes mismatch")
	}
	if _, err := os.Stat(filepath.Join(endingsDir, "abc123.json")); err != nil {
		t.Fatalf("meta not written: %v", err)
	}

	out, err := repo.Load(t.Context(), "abc123")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if out.ID != in.ID || out.EndingType != in.EndingType || out.Story != in.Story ||
		out.ImageFile != in.ImageFile || out.CreatedAt != in.CreatedAt ||
		out.Lives != in.Lives || out.Route != in.Route {
		t.Errorf("round-trip mismatch:\n in =%+v\n out=%+v", in, out)
	}
}

// 画像生成失敗(空 Image)時は画像ファイルを書かずメタのみ保存するフォールバック設計。
// Presentation が fallback 画像を指す前提だが、ここでは「保存は成功しメタは読める」ことだけ検証する。
func TestEndingRepo_Save_EmptyImage_WritesMetaOnly(t *testing.T) {
	repo, endingsDir, imageDir := newRepo(t)
	in := sampleEnding()

	if err := repo.Save(t.Context(), in, usecase.Image{}); err != nil {
		t.Fatalf("Save with empty image: %v", err)
	}

	// 画像は存在しない
	if _, err := os.Stat(filepath.Join(imageDir, "abc123.png")); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("image file should NOT exist on empty image, got err=%v", err)
	}
	// メタは存在し読める
	if _, err := os.Stat(filepath.Join(endingsDir, "abc123.json")); err != nil {
		t.Fatalf("meta should exist: %v", err)
	}
	if _, err := repo.Load(t.Context(), "abc123"); err != nil {
		t.Fatalf("Load after empty-image save: %v", err)
	}
}

func TestEndingRepo_Load_NotFound(t *testing.T) {
	repo, _, _ := newRepo(t)
	_, err := repo.Load(t.Context(), "nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

// Load は ID を filepath.Base で無害化する。パストラバーサルで endings 配下外へ抜けられないこと。
func TestEndingRepo_Load_PathTraversalSafe(t *testing.T) {
	repo, _, _ := newRepo(t)
	// 通り抜け先を tmp 直下に仕込み、ハズレなら NotFound になるはず。
	root := filepath.Dir(filepath.Dir(repo.dir)) // .../endings の2階層上=tmp
	leak := filepath.Join(root, "secret.json")
	if err := os.WriteFile(leak, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}

	for _, id := range []string{"../../secret", "..%2f..%2fsecret", "/etc/passwd"} {
		if _, err := repo.Load(t.Context(), id); !errors.Is(err, domain.ErrNotFound) {
			t.Errorf("id=%q should resolve to ErrNotFound (sandboxed), got %v", id, err)
		}
	}
}
