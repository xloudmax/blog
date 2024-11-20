package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const defaultBasePath = "uploads/test"

// getBasePath 动态获取基础路径
func getBasePath() string {
	if basePath := os.Getenv("BASE_PATH"); basePath != "" {
		return basePath
	}
	return defaultBasePath
}

// isPathInsideBase 验证路径是否在基础路径内
func isPathInsideBase(path, base string) bool {
	absBase, _ := filepath.Abs(base)
	absPath, _ := filepath.Abs(path)
	return strings.HasPrefix(absPath, absBase)
}

// ensureFolderExists 确保文件夹路径存在
func ensureFolderExists(folderPath string) error {
	return os.MkdirAll(folderPath, os.ModePerm)
}

// isValidFolderName 验证文件夹名称是否有效
func isValidFolderName(name string) bool {
	// 支持字母、数字、下划线和中文字符
	for _, char := range name {
		if !(char == '_' || (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || (char >= '0' && char <= '9') || (char >= 0x4E00 && char <= 0x9FA5)) {
			return false
		}
	}
	return true
}

// requireAdmin 检查管理员权限
func requireAdmin(c *gin.Context) {
	role, _ := c.Get("role")
	if role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "权限不足"})
		c.Abort()
	}
}

// GetFolders 返回现有文件夹列表
func GetFolders(c *gin.Context) {
	requireAdmin(c)
	if c.IsAborted() {
		return
	}

	basePath := getBasePath()
	var folders []string
	files, err := os.ReadDir(basePath)
	if err != nil {
		zap.S().Error("无法读取文件夹列表: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取文件夹列表"})
		return
	}

	for _, file := range files {
		if file.IsDir() {
			folders = append(folders, file.Name())
		}
	}

	c.JSON(http.StatusOK, gin.H{"folders": folders})
}

// CreateFolder 创建文件夹
func CreateFolder(c *gin.Context) {
	requireAdmin(c)
	if c.IsAborted() {
		return
	}

	var request struct {
		Folder string `json:"folder" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的文件夹名称"})
		return
	}

	if !isValidFolderName(request.Folder) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件夹名称包含非法字符"})
		return
	}

	basePath := getBasePath()
	folderPath := filepath.Join(basePath, request.Folder)
	if err := ensureFolderExists(folderPath); err != nil {
		zap.S().Error("创建文件夹失败: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建文件夹失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件夹创建成功"})
}

// UploadFile 处理文件上传
func UploadFile(c *gin.Context) {
	requireAdmin(c)
	if c.IsAborted() {
		return
	}

	title := c.PostForm("title")
	folder := c.PostForm("folder")
	if title == "" || folder == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件标题和文件夹名称不能为空"})
		return
	}

	if !isValidFolderName(folder) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件夹名称包含非法字符"})
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件上传失败"})
		return
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != ".md" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持上传 Markdown 文件 (.md)"})
		return
	}

	basePath := getBasePath()
	folderPath := filepath.Join(basePath, folder)
	if !isPathInsideBase(folderPath, basePath) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "非法路径"})
		return
	}

	if err := ensureFolderExists(folderPath); err != nil {
		zap.S().Error("确保文件夹存在失败: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件夹创建失败"})
		return
	}

	filePath := filepath.Join(folderPath, fmt.Sprintf("%s.md", title))
	if _, err := os.Stat(filePath); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "文件已存在，请使用其他名称"})
		return
	}

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		zap.S().Error("文件保存失败: ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "文件保存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "文件上传成功", "file_path": filePath})
}

// GetMarkdownContent 返回指定 Markdown 文件的内容
func GetMarkdownContent(c *gin.Context) {
	folder := c.Query("folder")
	fileName := c.Param("file")

	if folder == "" || fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件夹或文件名不能为空"})
		return
	}

	basePath := "uploads/test"
	filePath := filepath.Join(basePath, folder, fileName)

	if !isPathInsideBase(filePath, basePath) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "非法路径"})
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
		}
		return
	}

	// 验证是否是纯字符串
	response := string(content)
	if len(response) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件内容为空"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"content": response})
}

// ListMarkdownFiles 返回指定文件夹下的 Markdown 文件列表
func ListMarkdownFiles(c *gin.Context) {
	folder := c.Param("folder")
	if folder == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件夹名称不能为空"})
		return
	}

	basePath := "uploads/test"
	folderPath := filepath.Join(basePath, folder)

	if !isPathInsideBase(folderPath, basePath) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "非法路径"})
		return
	}

	files, err := os.ReadDir(folderPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "文件夹不存在"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件夹内容失败"})
		}
		return
	}

	var mdFiles []string
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".md" {
			mdFiles = append(mdFiles, file.Name())
		}
	}

	c.JSON(http.StatusOK, gin.H{"files": mdFiles})
}
