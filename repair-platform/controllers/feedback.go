package controllers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"net/http"
	"repair-platform/models"
)

// SubmitFeedback 允许用户为特定的维修请求提交反馈
// @Summary 提交维修反馈
// @Description 允许用户为特定的维修请求提交反馈
// @Tags 反馈
// @Accept json
// @Produce json
// @Param feedback body models.Feedback true "反馈内容"
// @Success 200 {object} map[string]interface{} "反馈提交成功"
// @Failure 400 {object} map[string]interface{} "输入数据无效"
// @Failure 404 {object} map[string]interface{} "未找到维修请求"
// @Failure 500 {object} map[string]interface{} "提交反馈失败"
// @Router /feedback [post]
func SubmitFeedback(c *gin.Context) {
	var feedback models.Feedback

	// 绑定 JSON 数据到反馈模型
	if err := c.ShouldBindJSON(&feedback); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "输入数据无效: " + err.Error()})
		return
	}

	db := c.MustGet("db").(*gorm.DB)

	// 验证相关的维修请求是否存在
	var repairRequest models.RepairRequest
	if err := db.Where("id = ?", feedback.RepairID).First(&repairRequest).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "未找到维修请求"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "检查维修请求失败: " + err.Error()})
		}
		return
	}

	// 设置反馈的评分
	feedback.SetRating(feedback.Rating)

	// 将反馈数据保存到数据库
	if err := db.Create(&feedback).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "提交反馈失败: " + err.Error()})
		return
	}

	// 反馈提交成功
	c.JSON(http.StatusOK, gin.H{"message": "反馈提交成功", "feedback": feedback})
}

// GetFeedbackByRepairID 允许管理员或用户查看特定维修请求的反馈
// @Summary 查看维修反馈
// @Description 允许管理员或用户查看特定维修请求的反馈
// @Tags 反馈
// @Produce json
// @Param id path string true "维修请求ID"
// @Success 200 {array} models.Feedback "反馈记录"
// @Failure 404 {object} map[string]interface{} "未找到此维修请求的反馈"
// @Failure 500 {object} map[string]interface{} "检索反馈失败"
// @Router /feedback/{id} [get]
func GetFeedbackByRepairID(c *gin.Context) {
	repairID := c.Param("id")

	db := c.MustGet("db").(*gorm.DB)
	var feedbacks []models.Feedback

	// 查找与特定维修请求关联的所有反馈记录
	if err := db.Where("repair_id = ?", repairID).Find(&feedbacks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检索反馈失败: " + err.Error()})
		return
	}

	// 如果没有找到反馈
	if len(feedbacks) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "未找到此维修请求的反馈"})
		return
	}

	// 返回所有找到的反馈记录
	c.JSON(http.StatusOK, feedbacks)
}
