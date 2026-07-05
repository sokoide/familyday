package persistence

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sokoide/familyday/server/internal/domain"
)

// newRepo は tmp 配下に endings を作った repo を返す(決定性のため t.TempDir 使用)。
func newRepo(t *testing.T) (*EndingRepo, string) {
	t.Helper()
	root := t.TempDir()
	endingsDir := filepath.Join(root, "endings")
	repo, err := NewEndingRepo(endingsDir)
	if err != nil {
		t.Fatalf("NewEndingRepo: %v", err)
	}
	return repo, endingsDir
}

func sampleEnding() domain.Ending {
	return domain.Ending{
		ID:         "abc123",
		EndingType: domain.EndingGreat,
		Lives:      domain.Lives(3),
		Route:      domain.RouteBefriend,
		Story:      "勇者はドラゴンと仲良くなった!",
		CreatedAt:  "2026-07-02T00:00:00Z",
	}
}

func TestEndingRepo_SaveLoad_RoundTrip(t *testing.T) {
	repo, endingsDir := newRepo(t)
	in := sampleEnding()

	if err := repo.Save(t.Context(), in); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// メタが正しいディレクトリ・ファイル名で書かれている
	if _, err := os.Stat(filepath.Join(endingsDir, "abc123.json")); err != nil {
		t.Fatalf("meta not written: %v", err)
	}

	out, err := repo.Load(t.Context(), "abc123")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if out.ID != in.ID || out.EndingType != in.EndingType || out.Story != in.Story ||
		out.CreatedAt != in.CreatedAt ||
		out.Lives != in.Lives || out.Route != in.Route {
		t.Errorf("round-trip mismatch:\n in =%+v\n out=%+v", in, out)
	}
}

func TestEndingRepo_Load_NotFound(t *testing.T) {
	repo, _ := newRepo(t)
	_, err := repo.Load(t.Context(), "nonexistent")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("want ErrNotFound, got %v", err)
	}
}

// Load は ID を filepath.Base で無害化する。パストラバーサルで endings 配下外へ抜けられないこと。
func TestEndingRepo_Load_PathTraversalSafe(t *testing.T) {
	repo, _ := newRepo(t)
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
