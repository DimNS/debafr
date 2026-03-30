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

.PHONY: run
run:
	@go run -race ./main.go

.PHONY: build
build:
	@CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w -X 'main.appVersion=0.0.0'" -o debafr ./main.go

.PHONY: tools
tools: deps
	@go install golang.org/x/vuln/cmd/govulncheck@latest
