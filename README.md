# Shared EV Charging Backend

> [中文说明 | Chinese README](./README.zh-CN.md)

## Project Overview
This project is a backend for a WeChat Mini Program for shared EV charging spots, developed in Go using the Gin web framework. It supports WeChat login, reservation, charging records, statistics, file upload, and admin management, suitable for family, community, and company scenarios.

## Features
- One-click WeChat Mini Program login (JWT authentication)
- Charging spot reservation (day/night shift)
- Charging record management (upload kWh, image, remarks, etc.)
- Charging record query and update (monthly filter, detail view, edit)
- Statistical reports (monthly, daily, by timeslot)
- User-specific electricity price management
- File upload (image, MinIO object storage)
- Admin permission control
- Health check endpoint
- Swagger API documentation

## Directory Structure
```
shared_charge/
├── config/           # Configuration loader
├── controllers/      # Route controllers (business APIs)
├── middleware/       # Gin middleware
├── migrations/       # Database migration SQL
├── models/           # Data models
├── scripts/          # DB migration scripts (Win/Linux)
├── service/          # Business logic layer
├── utils/            # Utilities
├── uploads/          # Uploaded file storage
├── logs/             # Log directory
├── tmp/              # Temp files
├── main.go           # Project entry
├── go.mod/go.sum     # Go dependency management
├── env.example       # Env variable example
├── Dockerfile        # Docker build
└── README.md         # Project docs
```

## Environment Variables
See `env.example`, copy to `.env` and fill in as needed:
- Database (PostgreSQL)
- Server port/mode
- JWT secret & expiration
- WeChat AppID/Secret
- Default price, file upload params
- MinIO config
- Redis config

## Install & Run
1. Install Go 1.18+
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Start service:
   ```bash
   go run main.go
   ```
   Or use Docker:
   ```bash
   docker build -t shared-charge .
   docker run -p 8080:8080 --env-file .env shared-charge
   ```

## Database Migration
- Install [golang-migrate](https://github.com/golang-migrate/migrate)
- Windows:
  ```powershell
  ./scripts/migrate.ps1 up
  ```
- Linux/macOS:
  ```bash
  ./scripts/migrate.sh up
  ```

## Docker Deployment
We provide Docker images via GitHub Actions. Deploy with:

1. Prepare `.env` file (copy from `env.example`)
2. Pull image:
   ```bash
   docker pull itnotf/shared-ev-charging-backend:latest
   ```
3. Run:
   ```bash
   docker run -p 8080:8080 --env-file .env itnotf/shared-ev-charging-backend:latest
   ```

## API Documentation (Swagger)
- Visit [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) after startup

### Main API List

#### Auth
- `POST /api/auth/login` WeChat login
- `POST /api/auth/refresh` Refresh token

#### User
- `GET /api/users/profile` Get user info
- `POST /api/users/profile` Update user info
- `GET /api/users/price` Get user price

#### Reservation
- `GET /api/reservations` List reservations
- `POST /api/reservations` Create reservation
- `DELETE /api/reservations/:id` Delete reservation
- `GET /api/reservations/current` Get current reservation
- `GET /api/reservations/current-status` Get current reservation & charging status

#### Charging Record
- `GET /api/records` List charging records
- `POST /api/records` Create charging record
- `GET /api/records/unsubmitted` List unsubmitted records
- `GET /api/records/list` List records by month
- `GET /api/records/:id` Get record detail
- `PUT /api/records/:id` Update record

#### File Upload
- `POST /api/upload/image` Upload image
- `GET /api/image/:filename` Get image

#### Statistics
- `GET /api/statistics/monthly` Monthly stats
- `GET /api/statistics/daily` Daily stats
- `GET /api/statistics/monthly-shift` Timeslot stats

#### System
- `GET /health` Health check

#### Admin (admin only)
- `GET /api/admin/users` List all users
- `POST /api/admin/user/can_reserve` Change user reservation permission
- `POST /api/admin/user/unit_price` Change user price
- `GET /api/admin/monthly_report` Monthly reconciliation report

### Swagger Doc Generation
This project uses [swag](https://github.com/swaggo/swag) for auto-generating API docs.

1. Install swag:
   ```bash
   go install github.com/swaggo/swag/cmd/swag@latest
   # Or (old Go)
   # go get -u github.com/swaggo/swag/cmd/swag
   # Add GOPATH/bin to PATH
   ```
2. Generate/update docs:
   ```bash
   swag init
   ```
   This will generate a `docs/` folder in the project root.

If you change APIs, rerun `swag init` to keep docs in sync.

## Tech Stack
- Go 1.18+
- Gin Web Framework
- GORM (PostgreSQL)
- JWT Auth
- WeChat Mini Program SDK
- MinIO Object Storage
- Redis
- Swagger API Docs
- Docker
- zap Logger

## Usage Examples
### WeChat Login
```bash
curl -X POST http://localhost:8080/api/auth/login -H "Content-Type: application/json" -d '{"code":"xxx"}'
```

### Get User Info
```bash
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/users/profile
```

### Upload Image
```bash
curl -X POST http://localhost:8080/api/upload/image -H "Authorization: Bearer <token>" -F "file=@test.jpg"
```

### Admin API Examples

#### List all users
```bash
curl -H "Authorization: Bearer <admin_token>" http://localhost:8080/api/admin/users
```

#### Change user reservation permission
```bash
curl -X POST http://localhost:8080/api/admin/user/can_reserve \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "can_reserve": false}'
```

#### Change user price
```bash
curl -X POST http://localhost:8080/api/admin/user/unit_price \
  -H "Authorization: Bearer <admin_token>" \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "unit_price": 0.65}'
```

#### Get monthly report
```bash
curl -H "Authorization: Bearer <admin_token>" \
  "http://localhost:8080/api/admin/monthly_report?month=2024-08"
```

## Contributing
Feel free to submit issues and PRs!

## License
Apache 2.0 
