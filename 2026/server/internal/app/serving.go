package app

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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
// index.html には PUBLIC_BASE_URL 由来の image base を埋め込む。
func spaHandler(staticDir, baseURL string) http.HandlerFunc {
	fs := http.FileServer(http.Dir(staticDir))
	return func(w http.ResponseWriter, r *http.Request) {
		full := filepath.Join(staticDir, filepath.Clean(r.URL.Path))
		if r.URL.Path != "/" {
			if info, err := os.Stat(full); err == nil && !info.IsDir() {
				fs.ServeHTTP(w, r)
				return
			}
		}
		indexPath := filepath.Join(staticDir, "index.html")
		data, err := os.ReadFile(indexPath)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		imageBase := strings.TrimRight(baseURL, "/") + "/images"
		inject := fmt.Sprintf(`<script>window.__FD_IMAGE_BASE__=%q;</script>`, imageBase)
		html := strings.Replace(string(data), "</head>", inject+"</head>", 1)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	}
}
