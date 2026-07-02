# ドラゴン城の秘宝(2026)— ビルド/テスト/実行
#
# よく使う:
#   make            # build + unit test
#   make integration GEMINI_API_KEY=xxx   # 実APIのE2Eテスト
#   make run        # サーバ起動(2026/.env を読む)
#   make dev        # フロント開発サーバ(HMR)
#   make check      # Clean Architecture 依存方向チェック

SERVER_DIR := 2026/server
WEB_DIR    := 2026/web
BIN        := $(SERVER_DIR)/bin/familyday
PORT       ?= 8080

# Clean Architecture チェックスクリプト(スキル同梱)。環境変数で上書き可。
CLEANARCH_SH ?= $(HOME)/.claude-glm/skills/cleanarch-master/scripts/check.sh

GOFLAGS :=
NPM ?= npm

.DEFAULT_GOAL := all
.PHONY: all build build-web build-server build-bin test unit integration vet fmt check clean run dev help

all: build unit

# --- build ---

build: build-web build-server

# node_modules が無いときだけ install する(冪等・高速)
build-web:
	@if [ ! -d $(WEB_DIR)/node_modules ]; then \
		echo ">> installing web deps"; \
		cd $(WEB_DIR) && $(NPM) install; \
	fi
	@echo ">> building web (-> $(SERVER_DIR)/static)"
	cd $(WEB_DIR) && $(NPM) run build

build-server:
	@echo ">> building server"
	cd $(SERVER_DIR) && go build $(GOFLAGS) ./...

build-bin: build-server
	@mkdir -p $(SERVER_DIR)/bin
	cd $(SERVER_DIR) && go build $(GOFLAGS) -o bin/familyday ./cmd/server
	@echo ">> binary: $(BIN)"

# --- test ---

test: unit

# ユニットテスト(ネットワーク不要・CI安全)
unit:
	cd $(SERVER_DIR) && go test $(GOFLAGS) ./...

# 統合テスト(実Gemini/Imagen使用)。GEMINI_API_KEY が無いと各テストは skip する。
integration:
	cd $(SERVER_DIR) && go test $(GOFLAGS) -tags integration ./...

# まとめて実行
test-all: unit integration

# --- quality ---

vet:
	cd $(SERVER_DIR) && go vet ./...

fmt:
	cd $(SERVER_DIR) && gofmt -s -w .

# Clean Architecture の依存方向・技術リーク検査
check:
	@if [ ! -f "$(CLEANARCH_SH)" ]; then \
		echo "[skip] cleanarch script not found: $(CLEANARCH_SH)"; \
	else \
		cd $(SERVER_DIR) && bash $(CLEANARCH_SH) ./internal/...; \
	fi

# --- run / dev ---

# バックグラウンドのフロント成果物は事前に make build-web 済みであること。
run:
	cd $(SERVER_DIR) && PORT=$(PORT) go run ./cmd/server

dev:
	cd $(WEB_DIR) && $(NPM) run dev

# --- misc ---

clean:
	rm -rf $(SERVER_DIR)/bin $(SERVER_DIR)/static $(SERVER_DIR)/data $(WEB_DIR)/dist
	cd $(SERVER_DIR) && go clean -testcache

help:
	@echo "Targets:"
	@echo "  make                         build + unit test (default)"
	@echo "  make build                   build web + server"
	@echo "  make build-bin               build server binary to $(BIN)"
	@echo "  make unit                    run Go unit tests (no network)"
	@echo "  make integration             run E2E tests (needs GEMINI_API_KEY)"
	@echo "  make test-all                unit + integration"
	@echo "  make vet / fmt               go vet / gofmt"
	@echo "  make check                   Clean Architecture dependency check"
	@echo "  make run [PORT=8080]         run server (reads 2026/.env)"
	@echo "  make dev                     vite dev server with API proxy"
	@echo "  make clean                   remove build/test artifacts"
