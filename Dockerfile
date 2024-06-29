FROM golang:1.22-alpine

WORKDIR /app

RUN apk add --no-cache make

RUN go install github.com/air-verse/air@latest \
 && go install github.com/pressly/goose/v3/cmd/goose@latest

COPY go.mod go.sum ./

RUN go mod download

COPY . .

CMD ["air"]
