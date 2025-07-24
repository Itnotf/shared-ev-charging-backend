# 构建阶段
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o app main.go

# 运行阶段
FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/app ./app
COPY --from=builder /app/uploads ./uploads
COPY --from=builder /app/config ./config
COPY --from=builder /app/env.example ./env.example
EXPOSE 8080
CMD ["./app"] 