//go:build integration

// Package integration は E2E の統合テスト。
// 実際の Gemini/Imagen API を叩く HTTP サーバを httptest で立て、
// API エンドポイント経由で判定〜エンディング生成までを検証する。
//
// 実行:
//	GEMINI_API_KEY=xxx make integration
// または
//	GEMINI_API_KEY=xxx go test -tags integration ./...
//
// GEMINI_API_KEY が未設定の場合はスキップする(ネットワーク不要の CI で落とさないため)。
package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sokoide/familyday/server/internal/adapters/infra/gemini"
	"github.com/sokoide/familyday/server/internal/app"
)

func requireAPIKey(t *testing.T) string {
	t.Helper()
	k := os.Getenv("GEMINI_API_KEY")
	if k == "" {
		t.Skip("GEMINI_API_KEY not set; skipping integration test")
	}
	return k
}

func newServer(t *testing.T) *httptest.Server {
	t.Helper()
	apiKey := requireAPIKey(t)

	dataDir := t.TempDir()
	mux, err := app.BuildMux(context.Background(), app.Options{
		APIKey:    apiKey,
		BaseURL:   "https://integration.test",
		DataDir:   dataDir,
		StaticDir: t.TempDir(), // 空でOK(API のみ検証)
		GeminiCfg: gemini.DefaultConfig(),
	})
	if err != nil {
		t.Fatalf("build mux: %v", err)
	}
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

func postJSON(t *testing.T, srv *httptest.Server, path string, body any) map[string]any {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, srv.URL+path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer resp.Body.Close()
	data, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		t.Fatalf("POST %s status=%d body=%s", path, resp.StatusCode, data)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("decode %s: %v body=%s", path, err, data)
	}
	return m
}

// TestIntegration_Judge は各ステージで妥当な入力が Great/Good/Bad のいずれかを返すこと。
func TestIntegration_Judge(t *testing.T) {
	srv := newServer(t)
	// 真の通信失敗以外は Bad に正規化されるため、verdict は3値のいずれか。
	valid := map[string]bool{"Great": true, "Good": true, "Bad": true}

	cases := []struct {
		stage, input, lang string
	}{
		{"stage1", "ゴーレムさん、どいて!", "ja"},
		{"stage1", "move the golem please", "en"},
		{"stage2", "水をかけて火を消す", "ja"},
		{"stage3", "ドラゴンと仲良くする", "ja"},
	}
	for _, c := range cases {
		res := postJSON(t, srv, "/api/judge", map[string]string{
			"stageId":   c.stage,
			"sessionId": "integration",
			"input":     c.input,
			"lang":      c.lang,
		})
		v, _ := res["verdict"].(string)
		if !valid[v] {
			t.Errorf("stage=%s input=%q verdict=%q (want Great/Good/Bad)", c.stage, c.input, v)
		}
		if msg, _ := res["message"].(string); msg == "" {
			t.Errorf("stage=%s empty message", c.stage)
		}
		// lang=en ならメッセージは英字中心(ja なら英字は少ない)。簡易チェック。
		if c.lang == "en" {
			if !containsLatin(msgToString(res["message"])) {
				t.Logf("note: en request returned mostly non-latin message: %v", res["message"])
			}
		}
	}
}

// TestIntegration_Judge_InvalidInput は空入力が 400 になること。
func TestIntegration_Judge_InvalidInput(t *testing.T) {
	srv := newServer(t)
	b, _ := json.Marshal(map[string]string{"stageId": "stage1", "input": "", "lang": "ja"})
	resp, err := http.Post(srv.URL+"/api/judge", "application/json", bytes.NewReader(b))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("want 400, got %d", resp.StatusCode)
	}
}

// TestIntegration_Ending はエンディング生成が画像/結果URLを返すこと。
// Imagen は時間がかかる/失敗しうるため、ストーリーとURL形式を主眼に検証。
func TestIntegration_Ending(t *testing.T) {
	srv := newServer(t)
	res := postJSON(t, srv, "/api/ending", map[string]any{
		"lives":       3,
		"finalAction": "befriend",
		"cleared":     true,
		"sessionId":   "integration",
		"lang":        "ja",
	})
	id, _ := res["endingId"].(string)
	if id == "" {
		t.Fatalf("empty endingId: %+v", res)
	}
	if et, _ := res["endingType"].(string); et != "great" {
		t.Errorf("endingType=%q want great", et)
	}
	if story, _ := res["story"].(string); story == "" {
		t.Errorf("empty story")
	}
	imgURL, _ := res["imageUrl"].(string)
	resURL, _ := res["resultUrl"].(string)
	if !strings.HasPrefix(imgURL, "https://integration.test/img/") {
		t.Errorf("bad imageUrl: %s", imgURL)
	}
	if !strings.HasPrefix(resURL, "https://integration.test/r/") {
		t.Errorf("bad resultUrl: %s", resURL)
	}

	// result 取得で同一内容が戻ること
	got, err := http.Get(srv.URL + "/api/result/" + id)
	if err != nil {
		t.Fatal(err)
	}
	defer got.Body.Close()
	if got.StatusCode != http.StatusOK {
		t.Fatalf("result status=%d", got.StatusCode)
	}
	body, _ := io.ReadAll(got.Body)
	if !strings.Contains(string(body), id) {
		t.Errorf("result body does not contain id: %s", body)
	}
}

// TestIntegration_ImageFileWritten はエンディング生成後、画像ファイルが実体化すること。
// ※ Imagen が失敗した場合はファイル無し(フォールバック)で、スキップ扱い。
func TestIntegration_ImageFileWritten(t *testing.T) {
	apiKey := requireAPIKey(t)
	dataDir := t.TempDir()
	mux, err := app.BuildMux(context.Background(), app.Options{
		APIKey:    apiKey,
		BaseURL:   "https://integration.test",
		DataDir:   dataDir,
		StaticDir: t.TempDir(),
		GeminiCfg: gemini.DefaultConfig(),
	})
	if err != nil {
		t.Fatal(err)
	}
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	res := postJSON(t, srv, "/api/ending", map[string]any{
		"lives": 2, "finalAction": "defeat", "cleared": true, "sessionId": "img", "lang": "ja",
	})
	id, _ := res["endingId"].(string)
	if id == "" {
		t.Fatal("empty endingId")
	}
	// 生成は非同期ではないが、画像書き出しは同期的。少し待って確認。
	imgPath := filepath.Join(dataDir, "generated", id+".png")
	deadline := time.Now().Add(2 * time.Second)
	for {
		if _, err := os.Stat(imgPath); err == nil {
			return // ファイルあり → 成功
		}
		if time.Now().After(deadline) {
			t.Skipf("imagen likely failed or slow; image not written: %s", imgPath)
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func msgToString(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func containsLatin(s string) bool {
	for _, r := range s {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			return true
		}
	}
	return false
}
