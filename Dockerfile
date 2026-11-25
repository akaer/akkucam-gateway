FROM golang:1.23-alpine AS build
WORKDIR /src
ARG COMMIT=
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
COPY web ./web
RUN CGO_ENABLED=0 GOOS=linux go test ./... && \
    CGO_ENABLED=0 GOOS=linux go build -trimpath \
      -ldflags="-s -w -X akkucam-gateway/internal/version.commit=${COMMIT}" \
      -o /out/akkucam-gateway ./cmd/gateway

FROM alpine:3.21
RUN apk add --no-cache ffmpeg && addgroup -S gateway && adduser -S -G gateway gateway
WORKDIR /app
COPY --from=build /out/akkucam-gateway /usr/local/bin/akkucam-gateway
COPY web /app/web
USER gateway
EXPOSE 8080/tcp 7200/udp 32108/udp
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -q -O /dev/null http://127.0.0.1:8080/api/health || exit 1
ENTRYPOINT ["/usr/local/bin/akkucam-gateway"]
