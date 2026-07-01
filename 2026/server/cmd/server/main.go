// Command server は Composition Root。具象 adapter を組み立て、
// 内側(usecase/domain)にはインタフェースを注入する。
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/sokoide/familyday/server/internal/adapters/infra/gemini"
	"github.com/sokoide/familyday/server/internal/adapters/infra/persistence"
	"github.com/sokoide/familyday/server/internal/adapters/infra/ratelimit"
	"github.com/sokoide/familyday/server/internal/adapters/infra/sysid"
	httpadapter "github.com/sokoide/familyday/server/internal/adapters/presentation/http"
	"github.com/sokoide/familyday/server/internal/usecase"
)

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// .env があれば読み込む(本番は環境変数を直接設定してもよい)
	if err := godotenv.Load(); err != nil {
		log.Println("no .env file loaded (using process env)")
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	baseURL := envOr("PUBLIC_BASE_URL", "http://localhost:8080")
	port := envOr("PORT", "8080")
	dataDir := envOr("DATA_DIR", "data")

	gcfg := gemini.DefaultConfig()
	if v := os.Getenv("GEMINI_MODEL_JUDGE"); v != "" {
		gcfg.ModelJudge = v
	}
	if v := os.Getenv("GEMINI_MODEL_STORY"); v != "" {
		gcfg.ModelStory = v
	}
	if v := os.Getenv("GEMINI_MODEL_IMAGEN"); v != "" {
		gcfg.ModelImagen = v
	}

	gclient, err := gemini.NewClient(ctx, apiKey, gcfg)
	if err != nil {
		log.Fatalf("gemini client: %v", err)
	}

	repo, err := persistence.NewEndingRepo(
		filepath.Join(dataDir, "endings"),
		filepath.Join(dataDir, "generated"),
	)
	if err != nil {
		log.Fatalf("ending repo: %v", err)
	}

	limiter := ratelimit.New()
	idgen := sysid.UUIDGen{}
	clock := sysid.SystemClock{}

	judgeUC := usecase.NewJudgeUseCase(gemini.NewJudgeGateway(gclient), limiter)
	endingUC := usecase.NewEndingUseCase(
		gemini.NewStoryGenerator(gclient),
		gemini.NewImagenGenerator(gclient),
		repo,
		limiter,
		idgen,
		clock,
	)

	h := httpadapter.NewHandler(judgeUC, endingUC, baseURL)

	mux := http.NewServeMux()

	// API
	mux.HandleFunc("/api/judge", h.Judge)
	mux.HandleFunc("/api/ending", h.Ending)
	mux.HandleFunc("/api/result/", h.Result) // /api/result/{id}

	// 生成画像の静的配信(トラバーサルは ServeFile が安全処理)
	imgDir := filepath.Join(dataDir, "generated")
	mux.Handle("/img/", http.StripPrefix("/img/", safeFileServer(imgDir)))

	// フロント静的配信 + SPA フォールバック(/r/{id} 等)
	staticDir := envOr("STATIC_DIR", "static")
	mux.HandleFunc("/", spaHandler(staticDir))

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           logMiddleware(mux),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("listening on :%s (baseURL=%s)", port, baseURL)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutCtx)
}

// safeFileServer は指定ディレクトリ配下のみ配信する。
func safeFileServer(dir string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := filepath.Base(r.URL.Path)
		if name == "" || name == "/" || name == "." {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(dir, name))
	})
}

// spaHandler は静的ファイルを配信し、未存在パスは index.html にフォールバックする。
func spaHandler(staticDir string) http.HandlerFunc {
	fs := http.FileServer(http.Dir(staticDir))
	return func(w http.ResponseWriter, r *http.Request) {
		// /api/, /img/ は上書き済み。ここはフロント。
		full := filepath.Join(staticDir, filepath.Clean(r.URL.Path))
		if _, err := os.Stat(full); err != nil {
			// 存在しない → index.html(SPA ルーティング /r/{id} 用)
			indexPath := filepath.Join(staticDir, "index.html")
			if _, e2 := os.Stat(indexPath); e2 == nil {
				r2 := r.Clone(r.Context())
				r2.URL.Path = "/"
				fs.ServeHTTP(w, r2)
				return
			}
			http.NotFound(w, r)
			return
		}
		fs.ServeHTTP(w, r)
	}
}

func logMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
