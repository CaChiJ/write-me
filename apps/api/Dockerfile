FROM golang:1.24-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git build-base

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/write-me-api ./cmd/server

FROM alpine:3.20

WORKDIR /app
RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/write-me-api /usr/local/bin/write-me-api

EXPOSE 8080

CMD ["write-me-api"]
