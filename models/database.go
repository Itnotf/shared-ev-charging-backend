package models

import (
	"fmt"
	"shared-charge/config"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() {
	cfg := config.GetConfig()

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Shanghai",
		cfg.Database.Host,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
		cfg.Database.Port,
		cfg.Database.SSLMode,
	)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent), // 完全禁用SQL日志
	})

	if err != nil {
	}

	// 设置数据库连接池参数
	sqlDB, err := DB.DB()
	if err != nil {
	}

	// 优化连接池配置
	sqlDB.SetMaxIdleConns(20)                  // 增加空闲连接数
	sqlDB.SetMaxOpenConns(200)                 // 增加最大连接数
	sqlDB.SetConnMaxLifetime(time.Hour)        // 连接最大复用时间
	sqlDB.SetConnMaxIdleTime(30 * time.Minute) // 空闲连接超时时间

}
