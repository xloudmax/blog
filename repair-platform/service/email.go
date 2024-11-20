package service

import (
	"context"
	"fmt"
	"math/rand"
	"net/smtp"
	"repair-platform/database"
	"time"

	"github.com/redis/go-redis/v9" // 导入 redis 包
	"go.uber.org/zap"
)

type EmailService interface {
	SendVerificationCode(ctx context.Context, to string) error
	VerifyVerificationCode(ctx context.Context, email string, code string) bool
}

type emailService struct {
	logger *zap.SugaredLogger
}

// NewEmailService 创建一个新的EmailService实例
func NewEmailService(logger *zap.SugaredLogger) EmailService {
	return &emailService{
		logger: logger,
	}
}

// SendVerificationCode 发送验证码到用户的邮箱
func (e *emailService) SendVerificationCode(ctx context.Context, to string) error {
	code, err := generateVerificationCode()
	if err != nil {
		e.logger.Errorf("Failed to generate verification code: %v", err)
		return fmt.Errorf("failed to generate verification code: %v", err)
	}

	// 调用发送邮件的内部函数
	if err := e.sendVerificationCode(to, code); err != nil {
		e.logger.Errorf("Failed to send verification code: %v", err)
		return err
	}

	// 将验证码存储到 Redis 中，设置5分钟过期
	err = database.GetRedisClient().Set(ctx, to, code, 5*time.Minute).Err()
	if err != nil {
		e.logger.Errorf("Failed to store verification code in Redis: %v", err)
		return fmt.Errorf("failed to store verification code in Redis: %v", err)
	}

	return nil
}

// sendVerificationCode 内部函数，用于发送邮件
func (e *emailService) sendVerificationCode(to string, code string) error {
	from := "xloudmaxx@gmail.com"
	password := "mbbf hrde wlpk bphe"

	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// 构建邮件内容
	msg := fmt.Sprintf("From: %s\nTo: %s\nSubject: Verification Code\n\nYour verification code is: %s", from, to, code)

	// 发送邮件
	auth := smtp.PlainAuth("", from, password, smtpHost)
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, []string{to}, []byte(msg))
	if err != nil {
		e.logger.Errorf("Failed to send email: %v", err)
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

// generateVerificationCode 生成6位验证码
func generateVerificationCode() (string, error) {
	rand.Seed(time.Now().UnixNano())
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	return code, nil
}

// VerifyVerificationCode 验证用户输入的验证码是否正确
func (e *emailService) VerifyVerificationCode(ctx context.Context, email string, code string) bool {
	storedCode, err := database.GetRedisClient().Get(ctx, email).Result()
	if err == redis.Nil {
		e.logger.Warnf("Verification code not found for email: %s", email)
		return false
	} else if err != nil {
		e.logger.Errorf("Failed to retrieve verification code from Redis: %v", err)
		return false
	}

	if storedCode != code {
		e.logger.Warnf("Verification code mismatch for email: %s", email)
		return false
	}

	return true
}
