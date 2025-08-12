#!/bin/bash

# 数据库迁移脚本
# 使用方法: ./scripts/migrate.sh [up|down|version|force]

set -e

# 读取 .env 文件
if [ -f ".env" ]; then
    echo "读取 .env 文件..."
    # 只读取数据库相关的环境变量
    while IFS= read -r line; do
        # 跳过注释行和空行
        if [[ ! "$line" =~ ^[[:space:]]*# ]] && [[ -n "$line" ]]; then
            # 只处理数据库相关的变量
            if [[ "$line" =~ ^DB_ ]]; then
                export "$line"
            fi
        fi
    done < .env
fi

# 数据库连接配置
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
DB_NAME=${DB_NAME:-"shared_charge"}
DB_USER=${DB_USER:-"postgres"}
DB_PASSWORD=${DB_PASSWORD:-"password"}

# 构建数据库URL
DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

# 迁移文件路径
MIGRATIONS_PATH="migrations"

echo "数据库迁移工具"
echo "数据库: ${DB_HOST}:${DB_PORT}/${DB_NAME}"
echo "迁移路径: ${MIGRATIONS_PATH}"
echo ""

case "$1" in
    "up")
        echo "执行迁移..."
        migrate -path ${MIGRATIONS_PATH} -database "${DATABASE_URL}" up
        ;;
    "down")
        echo "回滚迁移..."
        migrate -path ${MIGRATIONS_PATH} -database "${DATABASE_URL}" down
        ;;
    "version")
        echo "当前迁移版本..."
        migrate -path ${MIGRATIONS_PATH} -database "${DATABASE_URL}" version
        ;;
    "force")
        if [ -z "$2" ]; then
            echo "错误: 请指定版本号"
            echo "用法: ./scripts/migrate.sh force <version>"
            exit 1
        fi
        echo "强制设置迁移版本为 $2..."
        migrate -path ${MIGRATIONS_PATH} -database "${DATABASE_URL}" force $2
        ;;
    *)
        echo "用法: $0 {up|down|version|force <version>}"
        echo ""
        echo "命令说明:"
        echo "  up      - 执行所有未应用的迁移"
        echo "  down    - 回滚最后一个迁移"
        echo "  version - 显示当前迁移版本"
        echo "  force   - 强制设置迁移版本"
        exit 1
        ;;
esac 