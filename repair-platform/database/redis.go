package database

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"os"
	"time"
)

var redisClient *redis.Client
var ctx = context.Background()

// InitRedis 初始化 Redis 客户端
func InitRedis() {
	// 从环境变量读取 Redis 配置信息
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0 // 使用默认数据库

	// 创建 Redis 客户端
	redisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword, // 没有密码时为空字符串
		DB:       redisDB,       // 使用默认数据库
	})

	// 测试 Redis 连接
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("无法连接到 Redis: %v", err)
	}

	fmt.Println("Redis 连接成功")
}

// GetRedisClient 返回全局 Redis 客户端实例
func GetRedisClient() *redis.Client {
	return redisClient
}

// SetVerificationCode 在 Redis 中设置验证码，过期时间为 15 分钟
func SetVerificationCode(ctx context.Context, email string, code string) error {
	err := redisClient.Set(ctx, email, code, 15*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("设置验证码失败: %v", err)
	}
	return nil
}

// GetVerificationCode 从 Redis 中获取验证码
func GetVerificationCode(ctx context.Context, email string) (string, error) {
	code, err := redisClient.Get(ctx, email).Result()
	if errors.Is(err, redis.Nil) {
		return "", fmt.Errorf("验证码不存在或已过期")
	} else if err != nil {
		return "", fmt.Errorf("获取验证码失败: %v", err)
	}
	return code, nil
}

// DeleteVerificationCode 删除 Redis 中的验证码
func DeleteVerificationCode(ctx context.Context, email string) error {
	err := redisClient.Del(ctx, email).Err()
	if err != nil {
		return fmt.Errorf("删除验证码失败: %v", err)
	}
	return nil
}
