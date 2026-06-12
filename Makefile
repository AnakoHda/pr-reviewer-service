
MIGRATION_DIR = ./migrations
DATABASE = postgres

 .PHONY: migrate-up migrate-down migrate-status clean lint test build

lint:
	golangci-lint run

test:
	go test -v -race ./...

bench:
	go test -bench=. -benchmem ./internal/storage/...

build:
	go build -o bin/service ./cmd/service/main.go

bin/goose:
	@ECHO "Installing goose migration tool..."
	@mkdir bin
	@GOBIN=$(PWD)/bin go install github.com/pressly/goose/v3/cmd/goose@v3.25.0

bin/codegen_install:
	@ECHO "Installing goose codegen tool..."
	@mkdir -p bin
	@GOBIN=$(PWD)/bin go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
codegen: bin/codegen_install
	bin/oapi-codegen --config=api/codegen.yaml api/openapi.yml

#migrate-create: bin/goose
#	@./bin/goose -dir $(MIGRATION_DIR) create additional_indexes sql
migrate-up: bin/goose
	@./bin/goose -dir $(MIGRATION_DIR) postgres "$(POSTGRES_URL)" up
migrate-down: bin/goose
	@./bin/goose -dir $(MIGRATION_DIR) postgres "$(POSTGRES_URL)" down
migrate-status: bin/goose
	@./bin/goose -dir $(MIGRATION_DIR) postgres "$(POSTGRES_URL)" status

clean:
	rm -rf bin