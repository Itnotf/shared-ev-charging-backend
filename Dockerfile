# 构建阶段
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .

# 安装swag工具
RUN go install github.com/swaggo/swag/cmd/swag@latest

# 复制所有文件到工作目录
COPY . .

# 生成Swagger文档
RUN swag init

# 运行go mod tidy和go build命令
RUN go mod tidy && go build -o app main.go

# 运行阶段
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app ./app
EXPOSE 8080
CMD ["./app"] 