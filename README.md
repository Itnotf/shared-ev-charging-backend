# 充电位共享小程序后端

## 项目简介
本项目为微信群场景下的充电位共享小程序后端，基于 Go 语言开发，使用 Gin Web 框架，支持微信小程序登录、预约、充电记录、统计、文件上传等功能，适合家庭/社区/公司等场景。

## 主要功能
- 微信小程序一键登录（JWT 认证）
- 充电位预约（支持白班/夜班）
- 充电记录管理（上传用电量、图片、备注等）
- 统计报表（月度、每日、分时段）
- 用户专属电价管理
- 文件上传（图片，MinIO 对象存储）
- 管理员权限控制
- 健康检查接口
- Swagger API 文档

## 目录结构
```
shared_charge/
├── config/           # 配置加载
├── controllers/      # 路由控制器（业务接口）
├── middleware/       # Gin中间件
├── migrations/       # 数据库迁移SQL
├── models/           # 数据模型
├── scripts/          # 数据库迁移脚本（Win/Linux）
├── service/          # 业务逻辑层
├── utils/            # 工具类
├── uploads/          # 上传文件存储目录
├── logs/             # 日志目录
├── tmp/              # 临时文件目录
├── main.go           # 项目主入口
├── go.mod/go.sum     # Go依赖管理
├── env.example       # 环境变量示例
├── Dockerfile        # Docker镜像构建
└── README.md         # 项目说明
```

## 环境变量配置
请参考 `env.example` 文件，复制为 `.env` 并根据实际情况填写：
- 数据库连接（PostgreSQL）
- 服务端口/模式
- JWT 密钥与过期时间
- 微信小程序 AppID/Secret
- 默认电价、文件上传参数
- MinIO 对象存储配置
- Redis 配置

## 依赖安装与启动
1. 安装 Go 1.18 及以上版本
2. 安装依赖：
   ```bash
   go mod tidy
   ```
3. 启动服务：
   ```bash
   go run main.go
   ```
   或使用 Docker：
   ```bash
   docker build -t shared-charge .
   docker run -p 8080:8080 --env-file .env shared-charge
   ```

## 数据库迁移
- 需先安装 [golang-migrate](https://github.com/golang-migrate/migrate) 工具
- Windows:
  ```powershell
  ./scripts/migrate.ps1 up
  ```
- Linux/macOS:
  ```bash
  ./scripts/migrate.sh up
  ```

## API 文档（Swagger & swag 工具）
- 启动后访问 [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) 查看 Swagger API 文档
- 主要接口包括：
  - `/api/auth/login` 微信登录
  - `/api/auth/refresh` 刷新 token
  - `/api/users/profile` 用户信息
  - `/api/users/price` 用户电价
  - `/api/users/profile` 用户信息更新（POST）
  - `/api/reservations` 预约管理（GET/POST/DELETE）
  - `/api/reservations/current` 当前预约
  - `/api/reservations/current-status` 当前预约及充电状态
  - `/api/records` 充电记录（GET/POST）
  - `/api/records/unsubmitted` 未提交记录
  - `/api/upload/image` 文件上传
  - `/api/statistics/monthly` 月度统计
  - `/api/statistics/daily` 每日统计
  - `/api/statistics/monthly-shift` 分时段统计
  - `/health` 健康检查

### Swagger 文档生成与更新
本项目使用 [swag](https://github.com/swaggo/swag) 工具自动生成 API 文档。

1. 安装 swag 工具（仅需一次）：
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   # 或（老版本Go）
   # go get -u github.com/swaggo/swag/cmd/swag
   # GOPATH/bin 需加入环境变量PATH
   ```
2. 生成/更新文档：
   ```bash
   swag init
   ```
   该命令会在项目根目录生成 `docs/` 文件夹，包含自动生成的 Swagger 文档。

如有接口变更，请务必重新执行 `swag init` 保证文档同步。

## 技术栈
- Go 1.18+
- Gin Web 框架
- GORM（PostgreSQL）
- JWT 认证
- 微信小程序SDK
- MinIO 对象存储
- Redis 缓存
- Swagger API 文档
- Docker 容器化
- zap 日志

## 使用示例
### 用户微信登录
```bash
curl -X POST http://localhost:8080/api/auth/login -H "Content-Type: application/json" -d '{"code":"xxx"}'
```

### 获取用户信息
```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/users/profile
```

### 上传图片
```bash
curl -X POST http://localhost:8080/api/upload/image -H "Authorization: Bearer <token>" -F "file=@test.jpg"
```

## 贡献说明
欢迎提交 issue 和 PR 参与项目改进！

## License
Apache 2.0 
