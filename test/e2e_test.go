package test

import (
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCLIEndToEnd(t *testing.T) {
	// 设置测试环境
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// 构建可执行文件（从项目根目录构建）
	projectRoot := filepath.Dir(originalDir)
	execName := "envswitch_test"
	if runtime.GOOS == "windows" {
		execName = "envswitch_test.exe"
	}
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tempDir, execName), ".")
	buildCmd.Dir = projectRoot
	err := buildCmd.Run()
	if err != nil {
		t.Fatalf("Failed to build envswitch: %v", err)
	}

	binary := "./" + execName

	// 1. 测试项目创建
	t.Run("CreateProject", func(t *testing.T) {
		cmd := exec.Command(binary, "project", "create", "test-project", "--description", "Test project for E2E")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to create project: %v\nOutput: %s", err, string(output))
		}

		if !strings.Contains(string(output), "successfully") {
			t.Errorf("Expected success message, got: %s", string(output))
		}
	})

	// 2. 测试项目列表
	t.Run("ListProjects", func(t *testing.T) {
		cmd := exec.Command(binary, "project", "list")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to list projects: %v\nOutput: %s", err, string(output))
		}

		if !strings.Contains(string(output), "test-project") {
			t.Errorf("Expected to find 'test-project' in output: %s", string(output))
		}
	})

	// 3. 测试设置默认项目
	t.Run("SetDefaultProject", func(t *testing.T) {
		cmd := exec.Command(binary, "project", "set-default", "test-project")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to set default project: %v\nOutput: %s", err, string(output))
		}

		if !strings.Contains(string(output), "Default project set") {
			t.Errorf("Expected success message, got: %s", string(output))
		}
	})

	// 4. 测试环境创建
	t.Run("CreateEnvironment", func(t *testing.T) {
		cmd := exec.Command(binary, "env", "create", "test-project", "dev", "--description", "Development environment", "--tags", "development,local")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to create environment: %v\nOutput: %s", err, string(output))
		}

		if !strings.Contains(string(output), "created") {
			t.Errorf("Expected success message, got: %s", string(output))
		}
	})

	// 5. 测试环境列表
	t.Run("ListEnvironments", func(t *testing.T) {
		cmd := exec.Command(binary, "env", "list", "test-project")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to list environments: %v\nOutput: %s", err, string(output))
		}

		if !strings.Contains(string(output), "dev") {
			t.Errorf("Expected to find 'dev' environment in output: %s", string(output))
		}
	})

	// 6. 测试创建测试文件和添加文件配置
	t.Run("AddFileConfig", func(t *testing.T) {
		// 创建源文件
		sourceDir := filepath.Join(tempDir, "configs")
		err := os.MkdirAll(sourceDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create source directory: %v", err)
		}

		sourceFile := filepath.Join(sourceDir, "dev.json")
		err = os.WriteFile(sourceFile, []byte(`{"env": "development", "debug": true}`), 0644)
		if err != nil {
			t.Fatalf("Failed to create source file: %v", err)
		}

		// 创建目标目录
		targetDir := filepath.Join(tempDir, "app")
		err = os.MkdirAll(targetDir, 0755)
		if err != nil {
			t.Fatalf("Failed to create target directory: %v", err)
		}

		targetFile := filepath.Join(targetDir, "config.json")

		// 添加文件配置
		cmd := exec.Command(binary, "env", "add-file", "test-project", "dev", sourceFile, targetFile, "--description", "Development config")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to add file config: %v\nOutput: %s", err, string(output))
		}

		if !strings.Contains(string(output), "added") {
			t.Errorf("Expected success message, got: %s", string(output))
		}
	})

	// 7. 测试环境切换
	t.Run("SwitchEnvironment", func(t *testing.T) {
		cmd := exec.Command(binary, "switch", "test-project", "dev")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to switch environment: %v\nOutput: %s", err, string(output))
		}

		if !strings.Contains(string(output), "Successfully switched") {
			t.Errorf("Expected success message, got: %s", string(output))
		}

		// 验证目标文件是否被创建
		targetFile := filepath.Join(tempDir, "app", "config.json")
		if _, err := os.Stat(targetFile); os.IsNotExist(err) {
			t.Error("Target file should be created after environment switch")
		}

		// 验证文件内容
		content, err := os.ReadFile(targetFile)
		if err != nil {
			t.Fatalf("Failed to read target file: %v", err)
		}

		if !strings.Contains(string(content), "development") {
			t.Errorf("Target file should contain 'development', got: %s", string(content))
		}
	})

	// 8. 测试状态查看
	t.Run("CheckStatus", func(t *testing.T) {
		cmd := exec.Command(binary, "status")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to check status: %v\nOutput: %s", err, string(output))
		}

		if !strings.Contains(string(output), "test-project") {
			t.Errorf("Expected to find current project in status: %s", string(output))
		}
		if !strings.Contains(string(output), "dev") {
			t.Errorf("Expected to find current environment in status: %s", string(output))
		}
	})

	// 9. 测试回滚
	t.Run("Rollback", func(t *testing.T) {
		cmd := exec.Command(binary, "rollback", "--force")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to rollback: %v\nOutput: %s", err, string(output))
		}

		if !strings.Contains(string(output), "completed") {
			t.Errorf("Expected success message, got: %s", string(output))
		}
	})

	// 10. 测试环境删除
	t.Run("DeleteEnvironment", func(t *testing.T) {
		cmd := exec.Command(binary, "env", "delete", "test-project", "dev", "--force")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to delete environment: %v\nOutput: %s", err, string(output))
		}

		if !strings.Contains(string(output), "deleted") {
			t.Errorf("Expected success message, got: %s", string(output))
		}
	})

	// 11. 测试项目删除
	t.Run("DeleteProject", func(t *testing.T) {
		cmd := exec.Command(binary, "project", "delete", "test-project", "--force")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to delete project: %v\nOutput: %s", err, string(output))
		}

		if !strings.Contains(string(output), "deleted") {
			t.Errorf("Expected success message, got: %s", string(output))
		}
	})
}

func TestWebServerEndToEnd(t *testing.T) {
	// 设置测试环境
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// 构建可执行文件（从项目根目录构建）
	projectRoot := filepath.Dir(originalDir)
	execName := "envswitch_test"
	if runtime.GOOS == "windows" {
		execName = "envswitch_test.exe"
	}
	buildCmd := exec.Command("go", "build", "-o", filepath.Join(tempDir, execName), ".")
	buildCmd.Dir = projectRoot
	err := buildCmd.Run()
	if err != nil {
		t.Fatalf("Failed to build envswitch: %v", err)
	}

	binary := "./" + execName

	// 启动Web服务器
	cmd := exec.Command(binary, "server", "--port", "8081")
	err = cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer func() { _ = cmd.Process.Kill() }()

	// 等待服务器启动
	time.Sleep(2 * time.Second)

	// 测试Web服务器是否正常运行
	t.Run("ServerHealthCheck", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8081/")
		if err != nil {
			t.Fatalf("Failed to connect to server: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}
	})

	// 测试API端点
	t.Run("APIHealthCheck", func(t *testing.T) {
		resp, err := http.Get("http://localhost:8081/api/status")
		if err != nil {
			t.Fatalf("Failed to connect to API: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}
	})
}

func TestFileOperations(t *testing.T) {
	tempDir := t.TempDir()

	// 创建测试文件结构
	configsDir := filepath.Join(tempDir, "configs")
	appDir := filepath.Join(tempDir, "app")

	err := os.MkdirAll(configsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create configs directory: %v", err)
	}

	err = os.MkdirAll(appDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create app directory: %v", err)
	}

	// 创建源配置文件
	devConfig := filepath.Join(configsDir, "dev.json")
	prodConfig := filepath.Join(configsDir, "prod.json")
	targetConfig := filepath.Join(appDir, "config.json")

	devContent := `{"env": "development", "debug": true, "port": 3000}`
	prodContent := `{"env": "production", "debug": false, "port": 80}`

	err = os.WriteFile(devConfig, []byte(devContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write dev config: %v", err)
	}

	err = os.WriteFile(prodConfig, []byte(prodContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write prod config: %v", err)
	}

	// 创建初始目标文件
	initialContent := `{"env": "initial", "debug": false}`
	err = os.WriteFile(targetConfig, []byte(initialContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write initial config: %v", err)
	}

	t.Run("CompleteWorkflow", func(t *testing.T) {
		originalDir, _ := os.Getwd()
		defer func() { _ = os.Chdir(originalDir) }()
		_ = os.Chdir(tempDir)

		// 构建二进制文件（从项目根目录构建）
		projectRoot := filepath.Dir(originalDir)
		execName := "envswitch_test"
		if runtime.GOOS == "windows" {
			execName = "envswitch_test.exe"
		}
		buildCmd := exec.Command("go", "build", "-o", filepath.Join(tempDir, execName), ".")
		buildCmd.Dir = projectRoot
		err := buildCmd.Run()
		if err != nil {
			t.Fatalf("Failed to build envswitch: %v", err)
		}

		binary := "./" + execName

		// 创建项目
		cmd := exec.Command(binary, "project", "create", "file-test", "--description", "File operations test")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to create project: %v\nOutput: %s", err, string(output))
		}

		// 创建环境
		cmd = exec.Command(binary, "env", "create", "file-test", "dev", "--description", "Development")
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to create dev environment: %v\nOutput: %s", err, string(output))
		}

		cmd = exec.Command(binary, "env", "create", "file-test", "prod", "--description", "Production")
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to create prod environment: %v\nOutput: %s", err, string(output))
		}

		// 添加文件配置
		cmd = exec.Command(binary, "env", "add-file", "file-test", "dev", devConfig, targetConfig)
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to add dev file config: %v\nOutput: %s", err, string(output))
		}

		cmd = exec.Command(binary, "env", "add-file", "file-test", "prod", prodConfig, targetConfig)
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to add prod file config: %v\nOutput: %s", err, string(output))
		}

		// 切换到开发环境
		cmd = exec.Command(binary, "switch", "file-test", "dev")
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to switch to dev: %v\nOutput: %s", err, string(output))
		}

		// 验证文件内容
		content, err := os.ReadFile(targetConfig)
		if err != nil {
			t.Fatalf("Failed to read target config: %v", err)
		}

		if !strings.Contains(string(content), "development") {
			t.Errorf("Expected development config, got: %s", string(content))
		}

		// 切换到生产环境
		cmd = exec.Command(binary, "switch", "file-test", "prod")
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to switch to prod: %v\nOutput: %s", err, string(output))
		}

		// 验证文件内容
		content, err = os.ReadFile(targetConfig)
		if err != nil {
			t.Fatalf("Failed to read target config: %v", err)
		}

		if !strings.Contains(string(content), "production") {
			t.Errorf("Expected production config, got: %s", string(content))
		}

		// 回滚
		cmd = exec.Command(binary, "rollback", "--force")
		output, err = cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("Failed to rollback: %v\nOutput: %s", err, string(output))
		}

		// 验证回滚后的内容
		content, err = os.ReadFile(targetConfig)
		if err != nil {
			t.Fatalf("Failed to read target config after rollback: %v", err)
		}

		if !strings.Contains(string(content), "development") {
			t.Errorf("Expected development config after rollback, got: %s", string(content))
		}
	})
}
