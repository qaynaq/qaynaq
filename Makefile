.PHONY: help all proto ui-deps ui-build build bundle coordinator worker ui website clean

BINARY   := dist/qaynaq
UI_DIR   := ui
WEB_DIR  := website
PROTO_DIR := proto
PROTO_OUT := internal/protogen
STATIK_OUT := internal/

LOAD_ENV := set -a && . ./.env && set +a

help:
	@echo "Targets:"
	@echo "  proto           Generate Go code from .proto files"
	@echo "  ui-deps         Install UI dependencies"
	@echo "  ui-build        Build UI and embed into the binary via statik"
	@echo "  build           Build the Go binary"
	@echo "  bundle          Build UI + embed + Go binary"
	@echo "  coordinator     Run coordinator from ./.env"
	@echo "  worker          Run worker from ./.env"
	@echo "  ui              Start UI dev server"
	@echo "  website         Start website dev server"
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

ui-build:
	pnpm --prefix $(UI_DIR) build
	statik -src=$(UI_DIR)/dist -dest=$(STATIK_OUT) -f -m

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

clean:
	rm -rf dist $(UI_DIR)/dist
