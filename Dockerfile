FROM golang:1.24-alpine AS base
WORKDIR /app
RUN apk add --no-cache make
COPY go.mod go.sum ./
RUN go mod download

FROM base AS development
RUN go install github.com/air-verse/air@latest \
 && go install github.com/pressly/goose/v3/cmd/goose@latest
COPY . .
CMD ["air", "-c", ".air.toml"]

FROM base AS builder
COPY . .
RUN CGO_ENABLED=0 go build -o /app/sesamo -ldflags="-s -w" ./cmd/main.go

FROM alpine:latest AS production
WORKDIR /app
COPY --from=builder /app/sesamo /usr/local/bin/sesamo
RUN chmod +x /usr/local/bin/sesamo
CMD ["/usr/local/bin/sesamo"]
