.EXPORT_ALL_VARIABLES:
GOBIN = $(shell pwd)/bin

.PHONY: init
init: tools

.PHONY: deps
deps:
	@go mod tidy

.PHONY: audit
audit: tools
	@export PATH="$(shell pwd)/bin:$(PATH)"; govulncheck ./...
	@trivy fs ./

.PHONY: lint
lint:
	@golangci-lint config verify
	@golangci-lint run

.PHONY: test
test:
	@go test -race -failfast -count=1 ./...

.PHONY: dev
dev: dev-build
	@cd .dev && DEBAFR_DEV_MODE=true ./debafr

.PHONY: dev-build
# Собираем под машину разработчика, без GOOS/GOARCH: dev-стенд запускается
# локально, кросс-компиляция тут только мешает.
dev-build:
	@CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X 'main.appVersion=0.0.0'" -o .dev/debafr ./main.go

.PHONY: dev-status
dev-status:
	@cd .dev && DEBAFR_DEV_MODE=true ./debafr status

# Поднять стенд руками (compose), минуя TUI.
.PHONY: dev-up
dev-up:
	@cd .dev && APP_VERSION=$${APP_VERSION:-1.28.1-alpine} docker compose -f compose.blue.yaml up -d

.PHONY: dev-down
dev-down:
	@cd .dev && docker compose -f compose.blue.yaml down --remove-orphans
	@cd .dev && docker compose -f compose.green.yaml down --remove-orphans
	@docker rm -f debafr_test_nginx 2>/dev/null || true

# Убрать всё, что стенд оставил: контейнеры, сети, образы, json-лог.
.PHONY: dev-clean
dev-clean: dev-down
	@docker image ls -q 'nginx' | xargs -r docker image rm -f 2>/dev/null || true
	@rm -f .dev/debafr.json

.PHONY: release
release:
	@./scripts/release.sh $(if $(DRY_RUN),--dry-run)

.PHONY: tools
tools: deps
	@go install -ldflags="-s -w" golang.org/x/vuln/cmd/govulncheck@latest
