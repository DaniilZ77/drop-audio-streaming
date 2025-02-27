FROM golang:1.23.1-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o bin/audiostreaming ./cmd/audiostreaming

FROM alpine:3.21
RUN apk --no-cache add curl
WORKDIR /app
COPY --from=builder ./app/bin ./bin
COPY --from=builder ./app/config ./config
COPY --from=builder ./app/tls ./tls

HEALTHCHECK --interval=30s --timeout=1m --start-period=30s --start-interval=10s --retries=2 CMD curl -f http://localhost:8081/health

ENTRYPOINT ["./bin/audiostreaming"]