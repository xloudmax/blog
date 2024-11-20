package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"repair-platform/database"
	"repair-platform/routes"
	"repair-platform/service"
	"testing"
)

var testRouter *gin.Engine

// setupTest 初始化测试环境
func setupTest() {
	// 初始化数据库
	db, err := database.InitDB()
	if err != nil {
		panic("数据库初始化失败")
	}

	// 初始化 Email 服务
	emailService := service.NewEmailService(nil)

	// 初始化路由
	testRouter = gin.Default()
	routes.SetupRoutes(testRouter, db, emailService)

	// 设置为测试模式
	gin.SetMode(gin.TestMode)
}

// performRequest 执行测试 HTTP 请求
func performRequest(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody []byte
	if body != nil {
		reqBody, _ = json.Marshal(body)
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)
	return rec
}

// uniqueUsername 动态生成唯一用户名
func uniqueUsername() string {
	return fmt.Sprintf("testuser_%d", rand.Intn(100000))
}

// uniqueEmail 动态生成唯一邮箱
func uniqueEmail() string {
	return fmt.Sprintf("testuser_%d@example.com", rand.Intn(100000))
}

func TestRegister(t *testing.T) {
	setupTest()

	body := map[string]string{
		"username":    uniqueUsername(),
		"email":       uniqueEmail(),
		"password":    "password123",
		"invite_code": "",
	}

	resp := performRequest("POST", "/api/register", body, "")
	if resp.Code != http.StatusOK {
		t.Fatalf("Expected status %d but got %d", http.StatusOK, resp.Code)
	}
	t.Log("用户注册成功")
}

func TestVerifyEmail(t *testing.T) {
	setupTest()

	// 动态生成用户数据
	username := uniqueUsername()
	email := uniqueEmail()

	// 注册用户
	registerBody := map[string]string{
		"username":    username,
		"email":       email,
		"password":    "password123",
		"invite_code": "",
	}
	registerResp := performRequest("POST", "/api/register", registerBody, "")
	if registerResp.Code != http.StatusOK {
		t.Fatalf("Register failed, status: %d", registerResp.Code)
	}

	// 模拟邮箱验证
	verifyBody := map[string]string{
		"email": email,
		"code":  "123456", // 在测试模式下，任何验证码都可以通过
	}
	verifyResp := performRequest("POST", "/api/verify_email", verifyBody, "")
	if verifyResp.Code != http.StatusOK {
		t.Fatalf("Verify email failed, status: %d", verifyResp.Code)
	}
	t.Log("邮箱验证成功")
}

func TestLogin(t *testing.T) {
	setupTest()

	// 动态生成用户数据
	username := uniqueUsername()
	email := uniqueEmail()

	// 注册用户
	registerBody := map[string]string{
		"username":    username,
		"email":       email,
		"password":    "password123",
		"invite_code": "",
	}
	registerResp := performRequest("POST", "/api/register", registerBody, "")
	if registerResp.Code != http.StatusOK {
		t.Fatalf("Register failed, status: %d", registerResp.Code)
	}

	// 模拟邮箱验证
	verifyBody := map[string]string{
		"email": email,
		"code":  "123456",
	}
	verifyResp := performRequest("POST", "/api/verify_email", verifyBody, "")
	if verifyResp.Code != http.StatusOK {
		t.Fatalf("Verify email failed, status: %d", verifyResp.Code)
	}

	// 登录用户
	loginBody := map[string]string{
		"username": username,
		"password": "password123",
	}
	loginResp := performRequest("POST", "/api/login", loginBody, "")
	if loginResp.Code != http.StatusOK {
		t.Fatalf("Login failed, status: %d", loginResp.Code)
	}

	var responseBody map[string]interface{}
	err := json.Unmarshal(loginResp.Body.Bytes(), &responseBody)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	token, ok := responseBody["data"].(map[string]interface{})["token"].(string)
	if !ok {
		t.Fatalf("Failed to extract token from response")
	}

	t.Logf("用户登录成功, token: %s", token)
}

func TestFileUpload(t *testing.T) {
	setupTest()

	// 动态生成用户数据
	username := uniqueUsername()
	email := uniqueEmail()

	// 注册用户
	registerBody := map[string]string{
		"username":    username,
		"email":       email,
		"password":    "password123",
		"invite_code": "",
	}
	registerResp := performRequest("POST", "/api/register", registerBody, "")
	if registerResp.Code != http.StatusOK {
		t.Fatalf("Register failed, status: %d", registerResp.Code)
	}

	// 模拟邮箱验证
	verifyBody := map[string]string{
		"email": email,
		"code":  "123456",
	}
	verifyResp := performRequest("POST", "/api/verify_email", verifyBody, "")
	if verifyResp.Code != http.StatusOK {
		t.Fatalf("Verify email failed, status: %d", verifyResp.Code)
	}

	// 登录用户
	loginBody := map[string]string{
		"username": username,
		"password": "password123",
	}
	loginResp := performRequest("POST", "/api/login", loginBody, "")
	if loginResp.Code != http.StatusOK {
		t.Fatalf("Login failed, status: %d", loginResp.Code)
	}

	var loginResponseBody map[string]interface{}
	err := json.Unmarshal(loginResp.Body.Bytes(), &loginResponseBody)
	if err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	token := loginResponseBody["data"].(map[string]interface{})["token"].(string)

	// 测试上传文件
	body := map[string]string{
		"title":  "testfile",
		"folder": "testfolder",
	}
	resp := performRequest("POST", "/api/upload", body, token)
	if resp.Code != http.StatusOK {
		t.Fatalf("Expected status %d but got %d", http.StatusOK, resp.Code)
	}
	t.Log("文件上传测试通过")
}
