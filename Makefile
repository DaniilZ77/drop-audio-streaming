SERVICE_NAME=drop-audiostreaming

TEST_FLAGS=-count=1
BUILD_FLAGS=

.PHONY: run, build, lint, test, coverage, migrate-new, migrate-up, migrate-down

# TODO define your envs, switch log_level to `debug` during developing
PG_URL=postgres://postgres:postgres@localhost:5432/drop-audiostreaming

run: ### run app
	go run cmd/audiostreaming/main.go -db_url '$(PG_URL)' \
	-grpc_port localhost:50052 -grpc_user_client_addr localhost:50051 -http_port localhost:8081 -log_level debug -cert ./tls/cert.pem \
	-key ./tls/key.pem -minio_password minioadmin -minio_user minioadmin \
	-minio_endpoint 127.0.0.1:9000 -minio_bucket drop-audio \
	-minio_use_ssl false -minio_location us-east-1 \
	-chunk_size 1024 -grpc_client_retries 1 -grpc_client_timeout 2s

build: ### build app
	go build ${BUILD_FLAGS} -o ${SERVICE_NAME} cmd/audiostreaming/main.go

lint: ### run linter
	@golangci-lint --timeout=2m run

test: ### run test
	go test ${TEST_FLAGS} ./...

coverage: ### generate coverage report
	go test ${TEST_FLAGS} -coverprofile=coverage.out ./...
	go tool cover -html="coverage.out"

MIGRATION_NAME=composite_beats_view

migrate-new: ### create a new migration
	migrate create -ext sql -dir ./internal/data -seq ${MIGRATION_NAME}

migrate-up: ### apply all migrations
	migrate -path ./internal/data -database '$(PG_URL)?sslmode=disable' up

migrate-down: ### migration down
	migrate -path ./internal/data -database '$(PG_URL)?sslmode=disable' down

mock:
	mockery
