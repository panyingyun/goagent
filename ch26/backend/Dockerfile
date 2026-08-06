# 后端:Go 构建 -> 精简运行镜像
FROM golang:1.25-alpine AS build
WORKDIR /src
# 国内构建默认走 goproxy.cn,避免 proxy.golang.org 超时
ENV GOPROXY=https://goproxy.cn,direct
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/backend .

FROM alpine:3.20
WORKDIR /app
COPY --from=build /app/backend .
EXPOSE 8080
ENV DATABASE_PATH=/app/data/math_graph_ai.db
ENTRYPOINT ["./backend"]
