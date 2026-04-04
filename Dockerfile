# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /retro-treasure-api ./cmd/server

FROM alpine:3.20
RUN adduser -D -H -u 10001 appuser
WORKDIR /app
COPY --from=builder /retro-treasure-api /app/retro-treasure-api

ENV APP_NAME=retro-treasure-api
ENV APP_PORT=8080
EXPOSE 8080

USER appuser
CMD ["/app/retro-treasure-api"]
