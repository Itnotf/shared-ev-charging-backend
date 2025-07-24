# 构建阶段
FROM golang:1.23-alpine AS builder

# 设置工作目录
WORKDIR /app

# 复制go.mod和go.sum以缓存依赖
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制项目文件
COPY . .

# 生成Swagger文档（如果需要）
RUN go install github.com/swaggo/swag/cmd/swag@latest && swag init

# 构建可执行文件
RUN go build -o app main.go

# 运行阶段
FROM alpine:latest

# 设置工作目录
WORKDIR /app

# 从构建阶段复制可执行文件
COPY --from=builder /app/app .

# 如果需要，复制配置文件
# COPY --from=builder /app/.env ./

# 暴露应用程序端口
EXPOSE 8080

# 启动应用程序
CMD ["./app"]