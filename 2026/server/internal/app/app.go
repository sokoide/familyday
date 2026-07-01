// Package app は Composition Root のワイヤリングを提供する。
// cmd/server(本番起動)と統合テストの双方から再利用する。
// このパッケージは全レイヤを知ってよい「最外周」の組立係である。
package app

import (
	"context"
	"net/http"
	"path/filepath"

	"github.com/sokoide/familyday/server/internal/adapters/infra/gemini"
	"github.com/sokoide/familyday/server/internal/adapters/infra/persistence"
	"github.com/sokoide/familyday/server/internal/adapters/infra/ratelimit"
	"github.com/sokoide/familyday/server/internal/adapters/infra/sysid"
	httpadapter "github.com/sokoide/familyday/server/internal/adapters/presentation/http"
	"github.com/sokoide/familyday/server/internal/usecase"
)

// Options はサーバ構成。cmd/server が env から組み立てる。
type Options struct {
	APIKey    string
	BaseURL   string
	DataDir   string // エンディングJSON・生成画像のルート
	StaticDir string // フロントビルド成果物
	GeminiCfg gemini.Config
}

// BuildMux は具象 adapter を生成・注入し、ルーティング済みの http.Handler を返す。
func BuildMux(ctx context.Context, opts Options) (http.Handler, error) {
	gclient, err := gemini.NewClient(ctx, opts.APIKey, opts.GeminiCfg)
	if err != nil {
		return nil, err
	}

	repo, err := persistence.NewEndingRepo(
		filepath.Join(opts.DataDir, "endings"),
		filepath.Join(opts.DataDir, "generated"),
	)
	if err != nil {
		return nil, err
	}

	limiter := ratelimit.New()

	judgeUC := usecase.NewJudgeUseCase(gemini.NewJudgeGateway(gclient), limiter)
	endingUC := usecase.NewEndingUseCase(
		gemini.NewStoryGenerator(gclient),
		gemini.NewImagenGenerator(gclient),
		repo,
		limiter,
		sysid.UUIDGen{},
		sysid.SystemClock{},
	)

	h := httpadapter.NewHandler(judgeUC, endingUC, opts.BaseURL)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/judge", h.Judge)
	mux.HandleFunc("/api/ending", h.Ending)
	mux.HandleFunc("/api/result/", h.Result)
	mux.Handle("/img/", http.StripPrefix("/img/", safeFileServer(filepath.Join(opts.DataDir, "generated"))))
	mux.HandleFunc("/", spaHandler(opts.StaticDir))
	return mux, nil
}
