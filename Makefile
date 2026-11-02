.DEFAULT_GOAL := help
.PHONY: help build run validate test test-race test-cover cover-html test-integration \
        compose-up compose-down fmt fmt-check vet lint tidy check ci \
        docker-build docker-run clean

BINARY  := aurora-exporter
PKG     := github.com/Jeffersonmf/aurora-exporter
CONFIG  ?= config/aurora-exporter.example.yaml
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -s -w \
	-X $(PKG)/internal/version.Version=$(VERSION) \
	-X $(PKG)/internal/version.Commit=$(COMMIT) \
	-X $(PKG)/internal/version.Date=$(DATE)

help: ## 📖 Lista os comandos disponíveis
	@echo "🚚 aurora-exporter: comandos"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'

build: ## 🔨 Compila o binário (bin/aurora-exporter, versão via ldflags)
	@echo "🔨 Compilando $(VERSION)..."
	@CGO_ENABLED=0 go build -trimpath -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./cmd/aurora-exporter

run: build ## ▶️  Sobe o serve com CONFIG (Ctrl+C = shutdown gracioso)
	@echo "▶️  Rodando com $(CONFIG)..."
	@./bin/$(BINARY) serve -config $(CONFIG)

validate: build ## ✅ Valida o arquivo de configuração CONFIG
	@./bin/$(BINARY) validate -config $(CONFIG)

test: ## 🧪 Roda os testes unitários
	@echo "🧪 Rodando testes..."
	@go test ./...

test-race: ## 🧪 Testes unitários com o detector de race + goleak
	@echo "🧪 Rodando testes com -race..."
	@go test ./... -race

test-cover: ## 📊 Testes com relatório de cobertura por pacote
	@go test ./... -race -coverprofile=coverage.out -covermode=atomic
	@go tool cover -func=coverage.out | tail -1

cover-html: test-cover ## 📊 Abre o relatório de cobertura no navegador
	@go tool cover -html=coverage.out

compose-up: ## 🐳 Sobe Postgres (source), Postgres e MySQL (targets) locais
	@docker compose up -d --wait

compose-down: ## 🐳 Derruba os bancos locais
	@docker compose down -v

test-integration: compose-up ## 🔗 Suíte de integração contra bancos reais (docker compose)
	@echo "🔗 Rodando testes de integração..."
	@go test -tags integration -race -count=1 ./test/integration/...

fmt: ## 🎨 Formata o código (gofmt)
	@echo "🎨 Formatando..."
	@gofmt -w .

fmt-check: ## 🎨 Verifica formatação sem alterar nada
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

vet: ## 🔍 go vet (inclui o código com build tag de integração)
	@go vet ./... && go vet -tags integration ./test/...

lint: ## 🔍 golangci-lint (errcheck, govet, staticcheck, gosec, revive, ...)
	@echo "🔍 Rodando golangci-lint..."
	@golangci-lint run ./...

tidy: ## 🧹 go mod tidy e verifica que nada mudou
	@go mod tidy && git diff --exit-code go.mod go.sum

check: fmt-check vet lint test-race ## ✅ Roda tudo que o CI roda, localmente (sem integração)

ci: check ## 🤖 Alias de check, usado pelo pipeline de CI

docker-build: ## 🐳 Build multi-stage da imagem (binário estático, FROM scratch)
	@echo "🐳 Buildando imagem $(BINARY):$(VERSION)..."
	@docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --build-arg DATE=$(DATE) \
		-t $(BINARY):$(VERSION) -t $(BINARY):latest .

docker-run: docker-build ## 🐳 Roda a imagem com CONFIG montado em /etc/aurora-exporter
	@docker run --rm -p 8080:8080 --env-file .env \
		-v $(abspath $(CONFIG)):/etc/aurora-exporter/config.yaml:ro $(BINARY):latest

clean: ## 🧹 Remove artefatos de build
	@echo "🧹 Limpando..."
	@rm -rf bin/ coverage.out
