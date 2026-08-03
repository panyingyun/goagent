# 后端:Go 构建 -> 精简运行镜像
FROM golang:1.25-alpine AS build
WORKDIR /src
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
