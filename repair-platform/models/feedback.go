package models

import (
	"time"
)

// Feedback 定义了反馈的数据模型
type Feedback struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time `sql:"index"`
	UserID    uint       // 用户ID
	RepairID  uint       // 关联的维修请求ID
	Rating    int        // 评分
	Comments  string     `gorm:"type:varchar(255)"` // 反馈评论
}

// SetRating 设置反馈评分的逻辑
func (f *Feedback) SetRating(rating int) {
	if rating < 1 {
		f.Rating = 1
	} else if rating > 5 {
		f.Rating = 5
	} else {
		f.Rating = rating
	}
}
