package models

import (
	"gorm.io/gorm" // 导入新版 GORM 库
	"time"         // 用于处理时间类型
)

// RepairRequest 表示一个用户提交的维修请求
type RepairRequest struct {
	gorm.Model
	UserID       uint       `json:"user_id"`       // 提交请求的用户 ID，外键
	TechnicianID uint       `json:"technician_id"` // 负责处理的维修人员 ID（可为空），外键
	Description  string     `json:"description"`   // 报修的描述
	Status       string     `json:"status"`        // 报修状态：pending, in_progress, completed
	Location     string     `json:"location"`      // 报修的位置
	Priority     string     `json:"priority"`      // 紧急程度：low, medium, high
	ImageURL     string     `json:"image_url"`     // 上传的报修相关图片（可选）
	CompletedAt  *time.Time `json:"completed_at"`  // 任务完成时间（可为空）
}

// Possible statuses for a repair request
const (
	StatusPending    = "pending"     // 等待处理
	StatusInProgress = "in_progress" // 正在处理
	StatusCompleted  = "completed"   // 已完成
)

// SetStatus 更新报修请求的状态，并在状态为“completed”时设置完成时间
func (r *RepairRequest) SetStatus(status string) {
	r.Status = status
	if status == StatusCompleted {
		now := time.Now()
		r.CompletedAt = &now
	}
}
