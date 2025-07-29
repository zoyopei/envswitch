package config

import (
	"os"
	"testing"

	"github.com/zoyopei/envswitch/internal"
)

func TestInitConfig(t *testing.T) {
	// 创建临时目录用于测试
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	// 保存原始配置
	originalConfig := globalConfig
	defer func() { globalConfig = originalConfig }()

	_ = os.Chdir(tempDir)

	// 重置全局配置
	globalConfig = nil

	// 在当前目录创建一个 config.json 文件，确保使用默认配置
	defaultConfig := &internal.Config{
		DataDir:   DefaultDataDir,
		BackupDir: DefaultBackupDir,
		WebPort:   DefaultWebPort,
	}
	err := SaveConfig(defaultConfig)
	if err != nil {
		t.Fatalf("Failed to save default config: %v", err)
	}

	// 重置全局配置，然后测试初始化
	globalConfig = nil
	err = InitConfig()
	if err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}

	// 验证配置文件是否创建或全局配置是否已设置
	if globalConfig == nil {
		t.Error("Global config was not initialized")
	}

	// 验证默认配置值
	config := GetConfig()
	if config.DataDir != DefaultDataDir {
		t.Errorf("Expected DataDir = %s, got %s", DefaultDataDir, config.DataDir)
	}
	if config.BackupDir != DefaultBackupDir {
		t.Errorf("Expected BackupDir = %s, got %s", DefaultBackupDir, config.BackupDir)
	}
	if config.WebPort != DefaultWebPort {
		t.Errorf("Expected WebPort = %d, got %d", DefaultWebPort, config.WebPort)
	}
}

func TestLoadAndSaveConfig(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	// 保存原始配置
	originalConfig := globalConfig
	defer func() { globalConfig = originalConfig }()

	_ = os.Chdir(tempDir)

	// 重置全局配置
	globalConfig = nil

	// 创建测试配置
	testConfig := &internal.Config{
		DataDir:        "test_data",
		BackupDir:      "test_backups",
		WebPort:        9999,
		DefaultProject: "test_project",
	}

	// 保存配置
	err := SaveConfig(testConfig)
	if err != nil {
		t.Fatalf("SaveConfig() error = %v", err)
	}

	// 加载配置
	loadedConfig, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// 验证配置内容
	if loadedConfig.DataDir != testConfig.DataDir {
		t.Errorf("Expected DataDir = %s, got %s", testConfig.DataDir, loadedConfig.DataDir)
	}
	if loadedConfig.BackupDir != testConfig.BackupDir {
		t.Errorf("Expected BackupDir = %s, got %s", testConfig.BackupDir, loadedConfig.BackupDir)
	}
	if loadedConfig.WebPort != testConfig.WebPort {
		t.Errorf("Expected WebPort = %d, got %d", testConfig.WebPort, loadedConfig.WebPort)
	}
	if loadedConfig.DefaultProject != testConfig.DefaultProject {
		t.Errorf("Expected DefaultProject = %s, got %s", testConfig.DefaultProject, loadedConfig.DefaultProject)
	}
}

func TestUpdateConfig(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	// 保存原始配置
	originalConfig := globalConfig
	defer func() { globalConfig = originalConfig }()

	_ = os.Chdir(tempDir)

	// 重置全局配置
	globalConfig = nil

	// 初始化配置
	err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}

	// 更新配置
	updates := map[string]interface{}{
		"web_port":        8888,
		"default_project": "updated_project",
	}

	err = UpdateConfig(updates)
	if err != nil {
		t.Fatalf("UpdateConfig() error = %v", err)
	}

	// 验证更新结果
	config := GetConfig()
	if config.WebPort != 8888 {
		t.Errorf("Expected WebPort = 8888, got %d", config.WebPort)
	}
	if config.DefaultProject != "updated_project" {
		t.Errorf("Expected DefaultProject = updated_project, got %s", config.DefaultProject)
	}
}

func TestSetDefaultProject(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()

	// 保存原始配置
	originalConfig := globalConfig
	defer func() { globalConfig = originalConfig }()

	_ = os.Chdir(tempDir)

	// 重置全局配置
	globalConfig = nil

	// 初始化配置
	err := InitConfig()
	if err != nil {
		t.Fatalf("InitConfig() error = %v", err)
	}

	// 设置默认项目
	projectName := "test_default_project"
	err = SetDefaultProject(projectName)
	if err != nil {
		t.Fatalf("SetDefaultProject() error = %v", err)
	}

	// 验证设置结果
	if GetDefaultProject() != projectName {
		t.Errorf("Expected default project = %s, got %s", projectName, GetDefaultProject())
	}
}
