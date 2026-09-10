FROM golang:1.25-alpine AS builder

WORKDIR /app

# Устанавливаем сертификаты явно, если образ минимальный
RUN apk --no-cache add ca-certificates

COPY go.mod go.sum ./

# Используем кеш-маунт для папки модулей Go
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# Используем кеш-маунты для компилятора и модулей
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /main ./cmd/api

FROM scratch

COPY --from=builder /main /main
COPY .env .env
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 8080

ENTRYPOINT ["/main"]