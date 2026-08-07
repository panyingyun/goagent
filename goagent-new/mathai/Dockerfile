# Build context must be the goagent-new directory.
FROM node:22-alpine AS frontend-builder
WORKDIR /src/frontend
# 国内构建默认走 npmmirror,避免 registry.npmjs.org 超时
RUN npm config set registry https://registry.npmmirror.com
COPY mathai/frontend/package.json ./
RUN npm install --ignore-scripts --no-audit --no-fund
COPY mathai/frontend/ ./
RUN npm run build

FROM golang:1.25-alpine AS backend-builder
WORKDIR /src
# 国内构建默认走 goproxy.cn,避免 proxy.golang.org 超时
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod ./
COPY pkg/ ./pkg/
COPY mathai/backend/ ./mathai/backend/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/mathai ./mathai/backend


FROM alpine:3.22
RUN addgroup -S mathai && adduser -S -G mathai mathai
WORKDIR /app
COPY --from=backend-builder /out/mathai /app/mathai
COPY --from=frontend-builder /src/frontend/dist/ /app/frontend/dist/
RUN mkdir -p /app/data && chown -R mathai:mathai /app
USER mathai
ENV PORT=8080 MATHAI_DATA_FILE=/app/data/sessions.json MATHAI_STATIC_DIR=/app/frontend/dist
EXPOSE 8080
VOLUME ["/app/data"]
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 CMD wget -q -O - http://127.0.0.1:8080/livez || exit 1
ENTRYPOINT ["/app/mathai"]
