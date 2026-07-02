package app

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// setupDir は tmp 配下に画像ファイルを1つ置き、そのディレクトリパスを返す。
func setupDir(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	if name != "" {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func do(t *testing.T, h http.Handler, target string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	h.ServeHTTP(rec, req)
	return rec
}

// 当該ディレクトリ配下のファイルは配信される。
func TestSafeFileServer_ServesFile(t *testing.T) {
	dir := setupDir(t, "abc.png", "PNGDATA")
	h := safeFileServer(dir)

	rec := do(t, h, "/img/abc.png")
	if rec.Code != http.StatusOK {
		t.Fatalf("want 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if rec.Body.String() != "PNGDATA" {
		t.Errorf("body mismatch: %q", rec.Body.String())
	}
}

// 存在しないファイルは 404。
func TestSafeFileServer_NotFound(t *testing.T) {
	dir := setupDir(t, "", "")
	h := safeFileServer(dir)

	if rec := do(t, h, "/img/nope.png"); rec.Code != http.StatusNotFound {
		t.Errorf("want 404, got %d", rec.Code)
	}
}

// パストラバーサル: ../ や絶対パスでディレクトリ外へ抜けられない(filepath.Base で無害化)。
func TestSafeFileServer_PathTraversalSafe(t *testing.T) {
	dir := setupDir(t, "real.png", "ok")
	// tmp 直下(配信ディレクトリの1つ上)に secret を置き、リークしないか検証
	leak := filepath.Join(filepath.Dir(dir), "secret.txt")
	if err := os.WriteFile(leak, []byte("LEAK"), 0o644); err != nil {
		t.Fatal(err)
	}

	h := safeFileServer(dir)
	for _, target := range []string{
		"/img/../secret.txt",
		"/img/..%2fsecret.txt",
		"/img/%2e%2e%2fsecret.txt",
	} {
		rec := do(t, h, target)
		// いずれも配信ディレクトリ内と解釈され 404、あるいは本体は配信されるが LEAK は返らない
		if rec.Body.String() == "LEAK" {
			t.Errorf("%q leaked secret content (status=%d)", target, rec.Code)
		}
	}

	// 参考: cleanup(他テストに影響しないよう削除)
	_ = os.Remove(leak)
}

// 空/ルート/カレントは 404(クラッシュしない)。
func TestSafeFileServer_EmptyPath(t *testing.T) {
	dir := setupDir(t, "x.png", "x")
	h := safeFileServer(dir)

	for _, target := range []string{"/img/", "/img/.", "/img/"} {
		if rec := do(t, h, target); rec.Code != http.StatusNotFound {
			t.Errorf("target=%q want 404, got %d", target, rec.Code)
		}
	}
}
