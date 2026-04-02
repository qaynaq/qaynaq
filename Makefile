.PHONY: check-env protoc ui-deps statik build run-coordinator run-worker run-ui

check-env:
	@if [ ! -f .env ]; then \
		echo "Error: .env file not found!"; \
		echo "Please copy .env.example to .env and configure your settings:"; \
		echo "  cp .env.example .env"; \
		exit 1; \
	fi

protoc:
	protoc --proto_path=proto \
       --go_out=internal/protogen --go_opt=paths=source_relative \
       --go-grpc_out=internal/protogen --go-grpc_opt=paths=source_relative \
       --grpc-gateway_out=internal/protogen --grpc-gateway_opt=paths=source_relative \
       --validate_out=lang=go:internal/protogen --validate_opt=paths=source_relative \
       proto/*.proto

ui-deps:
	pnpm --prefix ui install

statik:
	pnpm --prefix ui build
	statik -src=ui/dist -dest=internal/ -f -m

build:
	make statik
	go build -o dist/qaynaq ./cmd/qaynaq

run-coordinator: check-env
	set -a && . ./.env && set +a && \
	./dist/qaynaq -role coordinator -grpc-port $${GRPC_PORT} -http-port $${HTTP_PORT}

run-worker: check-env
	set -a && . ./.env && set +a && \
	./dist/qaynaq -role worker -grpc-port $${WORKER_GRPC_PORT} -discovery-uri $${DISCOVERY_URI}

run-ui:
	pnpm --prefix ui run dev

run-website:
	pnpm --prefix website start

