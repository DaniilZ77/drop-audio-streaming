ARG GO_VERSION=1.23.1
ARG GO_BASE_IMAGE=alpine

FROM --platform=linux/amd64 golang:${GO_VERSION}-${GO_BASE_IMAGE} AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o bin/drop-audio-streaming ./cmd/audiostreaming
RUN go build -o bin/migrator ./cmd/migrator

FROM --platform=linux/amd64 alpine:latest
WORKDIR /app
COPY --from=builder ./app/bin ./bin
COPY --from=builder ./app/internal/data ./data
COPY --from=builder ./app/tls ./tls