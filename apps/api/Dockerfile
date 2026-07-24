# Builds both the api and worker binaries from one image; docker-compose's
# `command:` picks which one actually runs in a given container.
FROM golang:1.25-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -o /out/worker ./cmd/worker

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /out/api /app/api
COPY --from=builder /out/worker /app/worker
WORKDIR /app
