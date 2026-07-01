package app

import (
	"net/http"
	"os"
	"path/filepath"
)

// safeFileServer は指定ディレクトリ配下のみ配信する(パストラバーサル対策)。
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

// spaHandler は静的ファイルを配信し、未存在パスは index.html にフォールバックする(/r/{id} 用)。
func spaHandler(staticDir string) http.HandlerFunc {
	fs := http.FileServer(http.Dir(staticDir))
	return func(w http.ResponseWriter, r *http.Request) {
		full := filepath.Join(staticDir, filepath.Clean(r.URL.Path))
		if _, err := os.Stat(full); err != nil {
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
