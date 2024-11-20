package models

import (
	"time"

	"gorm.io/gorm"
)

// PasswordResetToken 用于存储密码重置或验证邮件的验证码
type PasswordResetToken struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	Token     string    `gorm:"not null;unique"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	ExpiresAt time.Time `gorm:"not null;index"`
}

// TableName 自定义表名
func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

// IsExpired 检查令牌是否已过期
func (t *PasswordResetToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// CreateToken 生成并存储新的密码重置令牌
func CreateToken(db *gorm.DB, userID uint, token string, duration time.Duration) (*PasswordResetToken, error) {
	resetToken := &PasswordResetToken{
		UserID:    userID,
		Token:     token,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}
	if err := db.Create(resetToken).Error; err != nil {
		return nil, err
	}
	return resetToken, nil
}

// DeleteExpiredTokens 清理已过期的令牌
func DeleteExpiredTokens(db *gorm.DB) error {
	if err := db.Where("expires_at < ?", time.Now()).Delete(&PasswordResetToken{}).Error; err != nil {
		return err
	}
	return nil
}
