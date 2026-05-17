.PHONY: help all proto ui-deps ui-build build bundle coordinator worker ui website test ui-test go-test lint clean

BINARY   := dist/qaynaq
UI_DIR   := ui
WEB_DIR  := website
PROTO_DIR := proto
PROTO_OUT := internal/protogen
WEB_DIST := internal/web/dist

LOAD_ENV := set -a && . ./.env && set +a

help:
	@echo "Targets:"
	@echo "  proto           Generate Go code from .proto files"
	@echo "  ui-deps         Install UI dependencies"
	@echo "  ui-build        Build UI and stage assets for embedding"
	@echo "  build           Build the Go binary"
	@echo "  bundle          Build UI + embed + Go binary"
	@echo "  coordinator     Run coordinator from ./.env"
	@echo "  worker          Run worker from ./.env"
	@echo "  ui              Start UI dev server"
	@echo "  website         Start website dev server"
	@echo "  test            Run Go and UI tests"
	@echo "  go-test         Run Go tests only"
	@echo "  ui-test         Run UI tests only"
	@echo "  lint            Run golangci-lint"
	@echo "  clean           Remove build artifacts"

all: help

proto:
	protoc --proto_path=$(PROTO_DIR) \
	       --go_out=$(PROTO_OUT) --go_opt=paths=source_relative \
	       --go-grpc_out=$(PROTO_OUT) --go-grpc_opt=paths=source_relative \
	       --grpc-gateway_out=$(PROTO_OUT) --grpc-gateway_opt=paths=source_relative \
	       --validate_out=lang=go:$(PROTO_OUT) --validate_opt=paths=source_relative \
	       $(PROTO_DIR)/*.proto

ui-deps:
	pnpm --prefix $(UI_DIR) install

ui-build: ui-deps
	pnpm --prefix $(UI_DIR) build
	rm -rf $(WEB_DIST)
	mkdir -p $(WEB_DIST)
	cp -r $(UI_DIR)/dist/. $(WEB_DIST)/

build:
	go build -o $(BINARY) ./cmd/qaynaq

bundle: ui-build build

coordinator:
	@test -f .env || { echo "Error: .env not found. Run: cp .env.example .env"; exit 1; }
	$(LOAD_ENV) && $(BINARY) -role coordinator -grpc-port $${GRPC_PORT} -http-port $${HTTP_PORT}

worker:
	@test -f .env || { echo "Error: .env not found. Run: cp .env.example .env"; exit 1; }
	$(LOAD_ENV) && $(BINARY) -role worker -grpc-port $${WORKER_GRPC_PORT} -discovery-uri $${DISCOVERY_URI}

ui:
	pnpm --prefix $(UI_DIR) run dev

website:
	pnpm --prefix $(WEB_DIR) start

test: go-test ui-test

go-test:
	go test ./...

ui-test: ui-deps
	pnpm --prefix $(UI_DIR) test

lint:
	golangci-lint run ./...

clean:
	rm -rf dist $(UI_DIR)/dist $(WEB_DIST)
