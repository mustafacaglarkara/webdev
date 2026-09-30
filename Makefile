GO      ?= go
PKGS    := ./...
BIN_DIR := bin
PORT    ?= 8080

.DEFAULT_GOAL := help

.PHONY: help build vet test race cover lint tidy fmt check integration vuln run-crm run-demo run-routerdemo i18ncheck stop-port clean

help: ## Hedefleri listeler
	@awk 'BEGIN {FS = ":.*## "} /^[a-zA-Z0-9_-]+:.*## / {printf "  %-16s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Tüm paketleri derler, komutları bin/ altına çıkarır
	$(GO) build $(PKGS)
	@mkdir -p $(BIN_DIR)
	$(GO) build -o $(BIN_DIR)/ ./cmd/...

vet: ## go vet
	$(GO) vet $(PKGS)

test: ## Testleri çalıştırır
	$(GO) test -count=1 $(PKGS)

race: ## Testleri yarış dedektörüyle çalıştırır
	$(GO) test -race -count=1 $(PKGS)

cover: ## Kapsam raporu üretir (coverage.out)
	$(GO) test -count=1 -coverprofile=coverage.out $(PKGS)
	$(GO) tool cover -func=coverage.out | tail -n 1

lint: ## golangci-lint (kuruluysa)
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint kurulu değil"; exit 1; }
	golangci-lint run $(PKGS)

tidy: ## go mod tidy
	$(GO) mod tidy

fmt: ## gofmt
	gofmt -s -w cmd pkg

check: build vet race ## CI ile aynı kontroller

integration: ## Gerçek sunuculara karşı testler (WEBDEV_IT_POSTGRES_DSN / _MYSQL_DSN / _SQLSERVER_DSN)
	$(GO) test -tags integration -race -count=1 -run TestIntegration -v ./pkg/db/

vuln: ## govulncheck
	$(GO) run golang.org/x/vuln/cmd/govulncheck@latest $(PKGS)

run-crm: ## Örnek CRM uygulamasını çalıştırır
	$(GO) run ./cmd/crm

run-demo: ## Migration/token demosunu çalıştırır
	$(GO) run ./cmd/demo

run-routerdemo: ## Router demosunu çalıştırır
	$(GO) run ./cmd/routerdemo

i18ncheck: ## Şablon ve locale anahtarlarını karşılaştırır
	$(GO) run ./cmd/i18ncheck

stop-port: ## PORT değişkenindeki portu dinleyen süreci durdurur
	./scripts/stop-port.sh $(PORT)

clean: ## Derleme çıktılarını siler
	rm -rf $(BIN_DIR) coverage.out
