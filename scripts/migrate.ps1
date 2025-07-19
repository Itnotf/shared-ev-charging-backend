param(
    [Parameter(Position=0)]
    [string]$Action = "up",
    [Parameter(Position=1)]
    [string]$Version
)

# PowerShell 迁移脚本，支持自动加载 .env
# 用法：./scripts/migrate.ps1 up|down|version|force <version>

# 1. 读取 .env 文件
$envFile = ".env"
if (Test-Path $envFile) {
    Get-Content $envFile | ForEach-Object {
        if ($_ -match '^(\w+)=(.*)$') {
            $key = $matches[1]
            $val = $matches[2]
            [System.Environment]::SetEnvironmentVariable($key, $val)
        }
    }
}

# 2. 读取环境变量
$DB_HOST = $env:DB_HOST
$DB_PORT = $env:DB_PORT
$DB_NAME = $env:DB_NAME
$DB_USER = $env:DB_USER
$DB_PASSWORD = $env:DB_PASSWORD



$DATABASE_URL = "postgres://$DB_USER`:$DB_PASSWORD@$DB_HOST`:$DB_PORT/$DB_NAME"+"?sslmode=disable"
$MIGRATIONS_PATH = "migrations"

Write-Host "Database Migration Tool (PowerShell)"
Write-Host "Database: ${DB_HOST}:${DB_PORT}/${DB_NAME}"
Write-Host "Migration Path: $MIGRATIONS_PATH"
Write-Host "$DATABASE_URL"

switch ($Action) {
    "up" {
        Write-Host "Executing migrations..."
        migrate -path $MIGRATIONS_PATH -database $DATABASE_URL up
    }
    "down" {
        Write-Host "Rolling back migrations..."
        migrate -path $MIGRATIONS_PATH -database $DATABASE_URL down
    }
    "version" {
        Write-Host "Current migration version..."
        migrate -path $MIGRATIONS_PATH -database $DATABASE_URL version
    }
    "force" {
        if (-not $Version) {
            Write-Host "Error: Please specify version number"
            Write-Host "Usage: ./scripts/migrate.ps1 force <version>"
            exit 1
        }
        Write-Host "Force setting migration version to $Version..."
        migrate -path $MIGRATIONS_PATH -database $DATABASE_URL force $Version
    }
    default {
        Write-Host "Usage: ./scripts/migrate.ps1 {up|down|version|force <version>}"
        Write-Host ""
        Write-Host "Commands:"
        Write-Host "  up      - Execute all pending migrations"
        Write-Host "  down    - Rollback last migration"
        Write-Host "  version - Show current migration version"
        Write-Host "  force   - Force set migration version"
        exit 1
    }
} 