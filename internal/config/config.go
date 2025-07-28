package config

import (
	"encoding/json"
	"envswitch/internal"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DefaultConfigFile = "config.json"
	DefaultDataDir    = "data"
	DefaultBackupDir  = "backups"
	DefaultWebPort    = 8080
)

var globalConfig *internal.Config

// InitConfig 初始化配置
func InitConfig() error {
	configPath := getConfigPath()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// 创建默认配置
		defaultConfig := &internal.Config{
			DataDir:   DefaultDataDir,
			BackupDir: DefaultBackupDir,
			WebPort:   DefaultWebPort,
		}
		return SaveConfig(defaultConfig)
	}

	config, err := LoadConfig()
	if err != nil {
		return err
	}

	globalConfig = config
	return ensureDirectories(config)
}

// LoadConfig 加载配置文件
func LoadConfig() (*internal.Config, error) {
	configPath := getConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config internal.Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	globalConfig = &config
	return &config, nil
}

// SaveConfig 保存配置文件
func SaveConfig(config *internal.Config) error {
	configPath := getConfigPath()

	// 确保配置目录存在
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	globalConfig = config
	return ensureDirectories(config)
}

// GetConfig 获取当前配置
func GetConfig() *internal.Config {
	if globalConfig == nil {
		// 如果配置未初始化，使用默认配置
		globalConfig = &internal.Config{
			DataDir:   DefaultDataDir,
			BackupDir: DefaultBackupDir,
			WebPort:   DefaultWebPort,
		}
	}
	return globalConfig
}

// UpdateConfig 更新配置
func UpdateConfig(updates map[string]interface{}) error {
	config := GetConfig()

	if dataDir, ok := updates["data_dir"]; ok {
		if dir, ok := dataDir.(string); ok {
			config.DataDir = dir
		}
	}

	if backupDir, ok := updates["backup_dir"]; ok {
		if dir, ok := backupDir.(string); ok {
			config.BackupDir = dir
		}
	}

	if webPort, ok := updates["web_port"]; ok {
		if port, ok := webPort.(int); ok {
			config.WebPort = port
		}
	}

	if defaultProject, ok := updates["default_project"]; ok {
		if proj, ok := defaultProject.(string); ok {
			config.DefaultProject = proj
		}
	}

	return SaveConfig(config)
}

// getConfigPath 获取配置文件路径
func getConfigPath() string {
	// 首先尝试当前目录下的config.json
	if _, err := os.Stat(DefaultConfigFile); err == nil {
		return DefaultConfigFile
	}

	// 然后尝试用户配置目录
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return DefaultConfigFile
	}

	return filepath.Join(homeDir, ".envswitch", DefaultConfigFile)
}

// ensureDirectories 确保必要的目录存在
func ensureDirectories(config *internal.Config) error {
	dirs := []string{
		config.DataDir,
		config.BackupDir,
		filepath.Join(config.DataDir, "projects"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}

// GetDataDir 获取数据目录路径
func GetDataDir() string {
	return GetConfig().DataDir
}

// GetBackupDir 获取备份目录路径
func GetBackupDir() string {
	return GetConfig().BackupDir
}

// GetWebPort 获取Web端口
func GetWebPort() int {
	return GetConfig().WebPort
}

// GetDefaultProject 获取默认项目
func GetDefaultProject() string {
	return GetConfig().DefaultProject
}

// SetDefaultProject 设置默认项目
func SetDefaultProject(projectName string) error {
	return UpdateConfig(map[string]interface{}{
		"default_project": projectName,
	})
}
