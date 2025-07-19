package config

import (
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Wechat   WechatConfig
	App      AppConfig
	MinIO    MinIOConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

type JWTConfig struct {
	Secret      string
	ExpireHours int
}

type WechatConfig struct {
	AppID  string
	Secret string
}

type AppConfig struct {
	DefaultUnitPrice float64
	MaxFileSize      int64
	UploadPath       string
}

type MinIOConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
	UseSSL     bool
}

type LogConfig struct {
	Mode     string
	Level    string
	FilePath string
}

var config *Config

// 环境变量缓存
var envCache = make(map[string]string)
var envCacheMutex sync.RWMutex

func LoadConfig() {
	// 加载.env文件
	godotenv.Load()

	config = &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Mode: getEnv("SERVER_MODE", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "shared_charge"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:      getEnv("JWT_SECRET", "your-jwt-secret-key"),
			ExpireHours: getEnvAsInt("JWT_EXPIRE_HOURS", 24),
		},
		Wechat: WechatConfig{
			AppID:  getEnv("WECHAT_APPID", ""),
			Secret: getEnv("WECHAT_SECRET", ""),
		},
		App: AppConfig{
			DefaultUnitPrice: getEnvAsFloat("DEFAULT_UNIT_PRICE", 0.7),
			MaxFileSize:      getEnvAsInt64("MAX_FILE_SIZE", 10485760), // 10MB
			UploadPath:       getEnv("UPLOAD_PATH", "./uploads"),
		},
		MinIO: MinIOConfig{
			Endpoint:   getEnv("MINIO_ENDPOINT", "localhost:9002"),
			AccessKey:  getEnv("MINIO_ACCESS_KEY", "your-access-key"),
			SecretKey:  getEnv("MINIO_SECRET_KEY", "your-secret-key"),
			BucketName: getEnv("MINIO_BUCKET_NAME", "shared-charge"),
			UseSSL:     getEnvAsBool("MINIO_USE_SSL", false),
		},
		Log: LogConfig{
			Mode:     getEnv("LOG_MODE", "debug"),
			Level:    getEnv("LOG_LEVEL", "debug"),
			FilePath: getEnv("LOG_FILE_PATH", "./logs/app.log"),
		},
	}
}

func GetConfig() *Config {
	return config
}

func getEnv(key, defaultValue string) string {
	// 先从缓存获取
	envCacheMutex.RLock()
	if value, exists := envCache[key]; exists {
		envCacheMutex.RUnlock()
		return value
	}
	envCacheMutex.RUnlock()

	// 缓存未命中，读取环境变量
	value := os.Getenv(key)
	if value == "" {
		value = defaultValue
	}

	// 更新缓存
	envCacheMutex.Lock()
	envCache[key] = value
	envCacheMutex.Unlock()

	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := getEnv(key, ""); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsInt64(key string, defaultValue int64) int64 {
	if value := getEnv(key, ""); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	if value := getEnv(key, ""); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := getEnv(key, ""); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
