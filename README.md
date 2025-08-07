# 充电位共享小程序后端

## 项目简介
本项目为微信群场景下的充电位共享小程序后端，基于 Go 语言开发，使用 Gin Web 框架，支持微信小程序登录、预约、充电记录、统计、文件上传等功能，适合家庭/社区/公司等场景。

## 主要功能
- 微信小程序一键登录（JWT 认证）
- 充电位预约（支持白班/夜班）
- 充电记录管理（上传用电量、图片、备注等）
- 充电记录查询与更新（按月筛选、详情查看、记录编辑）
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

## Docker 部署
我们已经通过GitHub Actions自动构建并发布Docker镜像。你可以使用以下步骤通过Docker部署应用程序。

### 准备 .env 文件
在项目根目录下，复制`env.example`文件并重命名为`.env`。根据实际环境需求，填写`.env`文件中的变量值。

### 拉取 Docker 镜像
从Docker Hub或GitHub Packages拉取最新的Docker镜像：

#### 从Docker Hub拉取镜像
```bash
docker pull itnotf/shared-ev-charging-backend:latest
```

### 运行 Docker 容器
使用以下命令运行Docker容器，并加载`.env`文件中的环境变量：

```bash
docker run -p 8080:8080 --env-file .env itnotf/shared-ev-charging-backend:latest
```

## API 文档（Swagger & swag 工具）
- 启动后访问 [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) 查看 Swagger API 文档

### 主要接口列表

#### 认证相关
- `POST /api/auth/login` 微信登录
- `POST /api/auth/refresh` 刷新 token

#### 用户相关
- `GET /api/users/profile` 获取用户信息
- `POST /api/users/profile` 更新用户信息
- `GET /api/users/price` 获取用户电价

#### 预约相关
- `GET /api/reservations` 获取预约列表
- `POST /api/reservations` 创建预约
- `DELETE /api/reservations/:id` 删除预约
- `GET /api/reservations/current` 获取当前预约
- `GET /api/reservations/current-status` 获取当前预约及充电状态

#### 充电记录相关
- `GET /api/records` 获取充电记录列表
- `POST /api/records` 创建充电记录
- `GET /api/records/unsubmitted` 获取未提交记录
- `GET /api/records/list` 获取指定月份充电记录列表
- `GET /api/records/:id` 获取充电记录详情
- `PUT /api/records/:id` 更新充电记录

#### 文件上传
- `POST /api/upload/image` 上传图片
- `GET /api/image/:filename` 获取图片

#### 统计相关
- `GET /api/statistics/monthly` 月度统计
- `GET /api/statistics/daily` 每日统计
- `GET /api/statistics/monthly-shift` 分时段统计

#### 系统相关
- `GET /health` 健康检查

#### 管理员相关（仅管理员可访问）
- `GET /api/admin/users` 获取所有用户列表
- `POST /api/admin/user/can_reserve` 修改用户预约权限
- `POST /api/admin/user/unit_price` 修改用户电价
- `GET /api/admin/monthly_report` 获取月度对账数据

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

### 管理员接口示例

#### 获取所有用户列表
```bash
curl -H "Authorization: Bearer <admin_token>" http://localhost:8080/api/admin/users
```

#### 修改用户预约权限
```bash
curl -X POST http://localhost:8080/api/admin/user/can_reserve \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "can_reserve": false}'
```

#### 修改用户电价
```bash
curl -X POST http://localhost:8080/api/admin/user/unit_price \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "unit_price": 0.65}'
```

#### 获取月度对账数据
```bash
curl -H "Authorization: Bearer <admin_token>" \
  "http://localhost:8080/api/admin/monthly_report?month=2024-08"
```

## 贡献说明
欢迎提交 issue 和 PR 参与项目改进！

## License
Apache 2.0 
