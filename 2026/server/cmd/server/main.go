// Command server は Composition Root(起動エントリ)。
// env を読み app.Options を組み立て、app.BuildMux で生成したハンドラを HTTP サーバとして仕上げる。
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/sokoide/familyday/server/internal/adapters/infra/gemini"
	"github.com/sokoide/familyday/server/internal/app"
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// loadEnvFirst は引数の候補パスを順に試し、最初に読めた .env で環境変数を設定する。
// 既にプロセス環境変数に設定されている値は上書きしない(godotenv の既定動作)。
// 1つでも読めれば true。
func loadEnvFirst(candidates ...string) bool {
	for _, p := range candidates {
		if err := godotenv.Load(p); err == nil {
			return true
		}
	}
	return false
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// .env があれば読み込む(本番は環境変数を直接設定してもよい。既存 env は上書きしない)。
	// サーバの CWD(2026/server) からでも 2026/.env を拾えるよう候補を複数試す。
	if !loadEnvFirst(".env", "../.env") {
		log.Println("no .env file loaded (using process env)")
	}

	gcfg := gemini.DefaultConfig()
	if v := os.Getenv("GEMINI_MODEL_JUDGE"); v != "" {
		gcfg.ModelJudge = v
	}
	if v := os.Getenv("GEMINI_MODEL_STORY"); v != "" {
		gcfg.ModelStory = v
	}
	if v := os.Getenv("GEMINI_MODEL_IMAGE"); v != "" {
		gcfg.ModelImage = v
	}
	// 画像サイズ(1辺px・正方形)。既定1024=ネイティブ出力(リサイズなし)。
	// 他モデル移行時や縮小用途に変更可。env: GEMINI_IMAGE_SIZE
	if v := os.Getenv("GEMINI_IMAGE_SIZE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			gcfg.ImageSize = n
		} else {
			log.Printf("invalid GEMINI_IMAGE_SIZE=%q, using default %d", v, gcfg.ImageSize)
		}
	}
	// 生成候補数(既定1=非バッチ)。増やすとコストN倍。env: GEMINI_IMAGE_COUNT
	if v := os.Getenv("GEMINI_IMAGE_COUNT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			gcfg.ImageCount = n
		} else {
			log.Printf("invalid GEMINI_IMAGE_COUNT=%q, using default %d", v, gcfg.ImageCount)
		}
	}

	opts := app.Options{
		APIKey:    os.Getenv("GEMINI_API_KEY"),
		BaseURL:   envOr("PUBLIC_BASE_URL", "http://localhost:8080"),
		DataDir:   envOr("DATA_DIR", "data"),
		StaticDir: envOr("STATIC_DIR", "static"),
		GeminiCfg: gcfg,
	}

	mux, err := app.BuildMux(ctx, opts)
	if err != nil {
		log.Fatalf("build server: %v", err)
	}

	port := envOr("PORT", "8080")
	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           logMiddleware(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s (baseURL=%s)", port, opts.BaseURL)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutCtx)
}

func logMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
