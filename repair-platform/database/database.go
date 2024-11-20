package database

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"repair-platform/models"
)

// InitDB 初始化数据库连接并自动迁移模型
func InitDB() (*gorm.DB, error) {
	// 自定义日志配置
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold: time.Second,   // 慢 SQL 阈值
			LogLevel:      logger.Silent, // 记录所有SQL
			Colorful:      true,          // 禁用彩色打印
		},
	)

	// 配置数据库连接选项
	db, err := gorm.Open(sqlite.Open("repair_platform.db"), &gorm.Config{
		Logger: newLogger,
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// 自动迁移数据库结构
	if err := db.AutoMigrate(&models.User{}, &models.RepairRequest{}, &models.Feedback{}, &models.PasswordResetToken{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	log.Println("Database connection and migration successful.")
	return db, nil
}

// CloseDB 关闭数据库连接
func CloseDB(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		log.Println("Failed to retrieve generic database object:", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		log.Println("Failed to close database:", err)
	} else {
		log.Println("Database connection closed successfully.")
	}
}
