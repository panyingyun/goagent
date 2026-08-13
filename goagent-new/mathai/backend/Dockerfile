FROM golang:1.25-alpine AS builder
WORKDIR /src
# 国内构建默认走 goproxy.cn,避免 proxy.golang.org 超时
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod ./
COPY pkg/ ./pkg/
COPY *.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/mathai .

FROM alpine:3.22
RUN addgroup -S mathai && adduser -S -G mathai mathai
WORKDIR /app
COPY --from=builder /out/mathai /app/mathai
RUN mkdir -p /app/data && chown -R mathai:mathai /app
USER mathai
ENV PORT=8080 MATHAI_DATA_FILE=/app/data/sessions.json
EXPOSE 8080
VOLUME ["/app/data"]
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD wget -q -O - http://127.0.0.1:8080/livez || exit 1
ENTRYPOINT ["/app/mathai"]
