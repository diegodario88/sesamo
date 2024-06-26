FROM golang:1.22-alpine

WORKDIR /app

RUN apk add --no-cache make

RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["air"]
