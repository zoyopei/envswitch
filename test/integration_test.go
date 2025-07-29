package test

import (
	"bytes"
	"encoding/json"
	"envswitch/internal"
	"envswitch/internal/config"
	"envswitch/internal/web"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func setupIntegrationTest(t *testing.T) (*web.Server, string) {
	tempDir := t.TempDir()

	// 设置临时配置
	testConfig := &internal.Config{
		DataDir:   tempDir + "/data",
		BackupDir: tempDir + "/backups",
		WebPort:   8080,
	}

	err := config.SaveConfig(testConfig)
	if err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// 保存当前目录
	originalDir, _ := os.Getwd()

	// 切换到项目根目录以确保模板文件路径正确
	projectRoot := filepath.Dir(originalDir)
	os.Chdir(projectRoot)

	// 恢复目录的函数
	t.Cleanup(func() {
		os.Chdir(originalDir)
	})

	server := web.NewServer()
	return server, tempDir
}

func TestAPIProjectLifecycle(t *testing.T) {
	server, _ := setupIntegrationTest(t)
	router := server.SetupRoutes()

	// 1. 测试列出空项目列表
	req, _ := http.NewRequest("GET", "/api/projects", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	projectsInterface, exists := response["projects"]
	if !exists {
		t.Error("Response missing 'projects' field")
	}

	if projectsInterface == nil {
		// 如果返回null，则表示0个项目
		return
	}

	projects, ok := projectsInterface.([]interface{})
	if !ok {
		t.Errorf("Projects field is not an array: %T", projectsInterface)
	}

	if len(projects) != 0 {
		t.Errorf("Expected 0 projects initially, got %d", len(projects))
	}

	// 2. 测试创建项目
	createReq := map[string]interface{}{
		"name":        "test-api-project",
		"description": "Test API project description",
	}
	reqBody, _ := json.Marshal(createReq)
	req, _ = http.NewRequest("POST", "/api/projects", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code 201, got %d. Response: %s", w.Code, w.Body.String())
	}

	var createdProject internal.Project
	_ = json.Unmarshal(w.Body.Bytes(), &createdProject)
	projectID := createdProject.ID

	if createdProject.Name != "test-api-project" {
		t.Errorf("Expected project name = test-api-project, got %s", createdProject.Name)
	}

	// 3. 测试获取项目详情
	req, _ = http.NewRequest("GET", "/api/projects/"+projectID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	var fetchedProject internal.Project
	_ = json.Unmarshal(w.Body.Bytes(), &fetchedProject)
	if fetchedProject.ID != projectID {
		t.Errorf("Expected project ID = %s, got %s", projectID, fetchedProject.ID)
	}

	// 4. 测试更新项目
	updateReq := map[string]interface{}{
		"name":        "updated-api-project",
		"description": "Updated description",
	}
	reqBody, _ = json.Marshal(updateReq)
	req, _ = http.NewRequest("PUT", "/api/projects/"+projectID, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d. Response: %s", w.Code, w.Body.String())
	}

	var updatedProject internal.Project
	_ = json.Unmarshal(w.Body.Bytes(), &updatedProject)
	if updatedProject.Name != "updated-api-project" {
		t.Errorf("Expected updated name = updated-api-project, got %s", updatedProject.Name)
	}

	// 5. 测试创建环境
	createEnvReq := map[string]interface{}{
		"name":        "test-environment",
		"description": "Test environment description",
		"tags":        []string{"development", "test"},
	}
	reqBody, _ = json.Marshal(createEnvReq)
	req, _ = http.NewRequest("POST", "/api/projects/"+projectID+"/environments", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status code 201, got %d. Response: %s", w.Code, w.Body.String())
	}

	var createdEnv internal.Environment
	json.Unmarshal(w.Body.Bytes(), &createdEnv)
	_ = createdEnv.ID // 忽略未使用的变量

	// 6. 测试列出环境
	req, _ = http.NewRequest("GET", "/api/projects/"+projectID+"/environments", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	var envResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &envResponse)
	if err != nil {
		t.Errorf("Failed to unmarshal environment response: %v", err)
	}

	envInterface, exists := envResponse["environments"]
	if !exists {
		t.Error("Response missing 'environments' field")
	}

	environments, ok := envInterface.([]interface{})
	if !ok {
		t.Errorf("Environments field is not an array: %T", envInterface)
	}

	if len(environments) != 1 {
		t.Errorf("Expected 1 environment, got %d", len(environments))
	}

	// 7. 测试删除项目
	req, _ = http.NewRequest("DELETE", "/api/projects/"+projectID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d. Response: %s", w.Code, w.Body.String())
	}

	// 验证项目已删除
	req, _ = http.NewRequest("GET", "/api/projects/"+projectID, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code 404 for deleted project, got %d", w.Code)
	}
}

func TestAPIErrorHandling(t *testing.T) {
	server, _ := setupIntegrationTest(t)
	router := server.SetupRoutes()

	// 测试创建项目时缺少必要字段
	req, _ := http.NewRequest("POST", "/api/projects", bytes.NewBuffer([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code 400 for missing required fields, got %d", w.Code)
	}

	// 测试获取不存在的项目
	req, _ = http.NewRequest("GET", "/api/projects/non-existent-id", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status code 404 for non-existent project, got %d", w.Code)
	}

	// 测试无效的JSON
	req, _ = http.NewRequest("POST", "/api/projects", bytes.NewBuffer([]byte(`invalid json`)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status code 400 for invalid JSON, got %d", w.Code)
	}
}

func TestAPIStatus(t *testing.T) {
	server, _ := setupIntegrationTest(t)
	router := server.SetupRoutes()

	// 测试获取状态
	req, _ := http.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", w.Code)
	}

	var status internal.AppState
	json.Unmarshal(w.Body.Bytes(), &status)

	// 初始状态应该是空的
	if status.CurrentProject != "" {
		t.Errorf("Expected empty current project initially, got %s", status.CurrentProject)
	}
}

func TestWebPages(t *testing.T) {
	server, _ := setupIntegrationTest(t)
	router := server.SetupRoutes()

	// 测试主页
	req, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200 for home page, got %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("EnvSwitch")) {
		t.Error("Home page should contain 'EnvSwitch'")
	}

	// 测试项目管理页面
	req, _ = http.NewRequest("GET", "/projects", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code 200 for projects page, got %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("项目管理")) {
		t.Error("Projects page should contain '项目管理'")
	}
}

func BenchmarkAPIProjectCreation(b *testing.B) {
	server, _ := setupIntegrationTest(&testing.T{})
	router := server.SetupRoutes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		createReq := map[string]interface{}{
			"name":        fmt.Sprintf("benchmark-project-%d", i),
			"description": "Benchmark test project",
		}
		reqBody, _ := json.Marshal(createReq)
		req, _ := http.NewRequest("POST", "/api/projects", bytes.NewBuffer(reqBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			b.Errorf("Expected status code 201, got %d", w.Code)
		}
	}
}
