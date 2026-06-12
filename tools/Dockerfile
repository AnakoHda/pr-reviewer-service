FROM golang:1.25-alpine AS builder


#RUN apk add --no-cache bash git postgresql-client
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.25.0

COPY entrypoint.sh /goose/entrypoint.sh
RUN chmod +x /goose/entrypoint.sh