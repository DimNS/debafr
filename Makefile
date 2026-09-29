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
dev:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w -X 'main.appVersion=0.0.0'" -o .dev/debafr ./main.go
	@.dev/debafr

.PHONY: release
release:
	@./scripts/release.sh $(if $(DRY_RUN),--dry-run)

.PHONY: tools
tools: deps
	@go install -ldflags="-s -w" golang.org/x/vuln/cmd/govulncheck@latest
