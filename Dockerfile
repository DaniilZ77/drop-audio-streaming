FROM golang:1.23.1-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o bin/drop-audio-streaming ./cmd/audiostreaming
RUN go build -o bin/migrator ./cmd/migrator

FROM busybox:1.37.0
WORKDIR /app
COPY --from=builder ./app/bin ./bin
COPY --from=builder ./app/internal/db/migrations ./migrations
COPY --from=builder ./app/tls ./tls