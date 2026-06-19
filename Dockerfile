# syntax=docker/dockerfile:1

FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates tzdata
COPY go.mod ./
RUN go mod download
COPY . .
ARG VERSION=dev
ARG COMMIT=none
ARG DATE=unknown
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -trimpath \
    -ldflags "-s -w -X github.com/syu6noob/devsync/internal/core.Version=${VERSION} -X github.com/syu6noob/devsync/internal/core.Commit=${COMMIT} -X github.com/syu6noob/devsync/internal/core.Date=${DATE}" \
    -o /out/devsync-server ./cmd/devsync-server

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S devsync && adduser -S -G devsync devsync
WORKDIR /app
COPY --from=builder /out/devsync-server /usr/local/bin/devsync-server
RUN mkdir -p /data && chown -R devsync:devsync /data
USER devsync
EXPOSE 8080
ENV DEVSYNC_ADDR=:8080
ENV DEVSYNC_DATA_DIR=/data
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://127.0.0.1:8080/healthz || exit 1
ENTRYPOINT ["devsync-server"]
