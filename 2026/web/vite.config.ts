import { defineConfig } from "vite";
import { resolve } from "path";

// ビルド成果物をGoサーバの静的配信ディレクトリへ直接出力
export default defineConfig({
  build: {
    outDir: resolve(__dirname, "../server/static"),
    emptyOutDir: true,
  },
  server: {
    // 開発時にGoバックエンド(8080)へプロキシ
    proxy: {
      "/api": "http://localhost:8080",
      "/img": "http://localhost:8080",
    },
  },
});
