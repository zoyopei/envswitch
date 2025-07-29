package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zoyopei/EnvSwitch/internal"
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
		homeDir, err := os.UserHomeDir()
		var defaultDataDir, defaultBackupDir string
		if err != nil {
			// 如果无法获取用户目录，使用当前目录
			defaultDataDir = DefaultDataDir
			defaultBackupDir = DefaultBackupDir
		} else {
			// 使用用户主目录下的 .envswitch
			defaultDataDir = filepath.Join(homeDir, ".envswitch", "data")
			defaultBackupDir = filepath.Join(homeDir, ".envswitch", "backups")
		}

		defaultConfig := &internal.Config{
			DataDir:            defaultDataDir,
			BackupDir:          defaultBackupDir,
			WebPort:            DefaultWebPort,
			OriginalDataDir:    defaultDataDir,
			EnableDataDirCheck: true, // 默认启用数据目录检查
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
		homeDir, err := os.UserHomeDir()
		var defaultDataDir, defaultBackupDir string
		if err != nil {
			// 如果无法获取用户目录，使用当前目录
			defaultDataDir = DefaultDataDir
			defaultBackupDir = DefaultBackupDir
		} else {
			// 使用用户主目录下的 .envswitch
			defaultDataDir = filepath.Join(homeDir, ".envswitch", "data")
			defaultBackupDir = filepath.Join(homeDir, ".envswitch", "backups")
		}

		globalConfig = &internal.Config{
			DataDir:            defaultDataDir,
			BackupDir:          defaultBackupDir,
			WebPort:            DefaultWebPort,
			OriginalDataDir:    defaultDataDir,
			EnableDataDirCheck: true,
		}
	}
	return globalConfig
}

// UpdateConfig 更新配置
func UpdateConfig(updates map[string]interface{}) error {
	config := GetConfig()

	// 检查是否尝试更新 data_dir
	if newDataDir, ok := updates["data_dir"]; ok {
		if dir, ok := newDataDir.(string); ok && dir != config.DataDir {
			// 检测到数据目录变更，进行安全检查
			if err := handleDataDirChange(config, dir); err != nil {
				return err
			}
		}
	}

	// 其他配置更新
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

	if enableCheck, ok := updates["enable_data_dir_check"]; ok {
		if enable, ok := enableCheck.(bool); ok {
			config.EnableDataDirCheck = enable
		}
	}

	return SaveConfig(config)
}

// handleDataDirChange 处理数据目录变更
func handleDataDirChange(config *internal.Config, newDataDir string) error {
	// 检查是否启用了数据目录检查
	if !config.EnableDataDirCheck {
		fmt.Println("⚠️  警告: 数据目录检查已禁用，直接更新数据目录路径")
		config.DataDir = newDataDir
		return nil
	}

	currentDataDir := config.DataDir

	// 检查当前数据目录是否存在且包含数据
	hasData, err := CheckDataDirHasData(currentDataDir)
	if err != nil {
		return fmt.Errorf("检查当前数据目录失败: %w", err)
	}

	// 如果当前数据目录没有数据，直接更新
	if !hasData {
		fmt.Printf("✅ 当前数据目录 '%s' 为空，安全更新到 '%s'\n", currentDataDir, newDataDir)
		config.DataDir = newDataDir
		updateDataDirHistory(config, currentDataDir)
		return nil
	}

	// 有数据的情况下，需要用户确认
	fmt.Printf("⚠️  危险操作: 检测到数据目录变更!\n")
	fmt.Printf("   当前数据目录: %s (包含项目数据)\n", currentDataDir)
	fmt.Printf("   新数据目录:   %s\n", newDataDir)
	fmt.Printf("\n")
	fmt.Printf("🔥 警告: 更改数据目录将导致无法访问当前的所有项目和环境数据!\n")
	fmt.Printf("\n")
	fmt.Printf("可选操作:\n")
	fmt.Printf("  1. 取消更改 (推荐)\n")
	fmt.Printf("  2. 迁移数据到新目录\n")
	fmt.Printf("  3. 强制更改 (当前数据将丢失)\n")
	fmt.Printf("\n")

	choice, err := promptUser("请选择操作 (1/2/3): ")
	if err != nil {
		return err
	}

	switch strings.TrimSpace(choice) {
	case "1":
		return fmt.Errorf("用户取消了数据目录更改")
	case "2":
		return migrateDataDir(config, currentDataDir, newDataDir)
	case "3":
		return forceUpdateDataDir(config, currentDataDir, newDataDir)
	default:
		return fmt.Errorf("无效的选择，操作已取消")
	}
}

// CheckDataDirHasData 检查数据目录是否包含数据 (导出函数)
func CheckDataDirHasData(dataDir string) (bool, error) {
	projectsDir := filepath.Join(dataDir, "projects")

	// 检查项目目录是否存在
	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		return false, nil
	}

	// 检查是否有项目文件
	files, err := os.ReadDir(projectsDir)
	if err != nil {
		return false, err
	}

	// 检查是否有 .json 文件（项目文件）
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			return true, nil
		}
	}

	return false, nil
}

// updateDataDirHistory 更新数据目录历史
func updateDataDirHistory(config *internal.Config, oldDataDir string) {
	// 设置原始数据目录（如果还没有设置）
	if config.OriginalDataDir == "" {
		config.OriginalDataDir = oldDataDir
	}

	// 添加到历史记录
	if config.DataDirHistory == nil {
		config.DataDirHistory = []string{}
	}

	// 避免重复记录
	for _, dir := range config.DataDirHistory {
		if dir == oldDataDir {
			return
		}
	}

	config.DataDirHistory = append(config.DataDirHistory, oldDataDir)
}

// migrateDataDir 迁移数据目录
func migrateDataDir(config *internal.Config, oldDataDir, newDataDir string) error {
	fmt.Printf("\n🔄 开始迁移数据从 '%s' 到 '%s'...\n", oldDataDir, newDataDir)

	// 创建新数据目录
	if err := os.MkdirAll(newDataDir, 0755); err != nil {
		return fmt.Errorf("创建新数据目录失败: %w", err)
	}

	// 检查新目录是否为空
	if newDirHasData, err := CheckDataDirHasData(newDataDir); err != nil {
		return fmt.Errorf("检查新数据目录失败: %w", err)
	} else if newDirHasData {
		confirm, err := promptUser("⚠️  新数据目录已包含数据，是否覆盖? (y/N): ")
		if err != nil {
			return err
		}
		if strings.ToLower(strings.TrimSpace(confirm)) != "y" {
			return fmt.Errorf("用户取消了数据迁移")
		}
	}

	// 创建备份
	timestamp := time.Now().Format("20060102_150405")
	backupDir := fmt.Sprintf("%s_backup_%s", oldDataDir, timestamp)

	fmt.Printf("📦 创建数据备份到: %s\n", backupDir)
	if err := copyDir(oldDataDir, backupDir); err != nil {
		return fmt.Errorf("创建备份失败: %w", err)
	}

	// 迁移数据
	fmt.Printf("📁 迁移数据...\n")
	if err := copyDir(oldDataDir, newDataDir); err != nil {
		return fmt.Errorf("数据迁移失败: %w", err)
	}

	// 更新配置
	config.DataDir = newDataDir
	updateDataDirHistory(config, oldDataDir)

	fmt.Printf("✅ 数据迁移完成!\n")
	fmt.Printf("   原数据备份: %s\n", backupDir)
	fmt.Printf("   新数据目录: %s\n", newDataDir)
	fmt.Printf("\n💡 提示: 确认新目录工作正常后，可以删除备份目录\n")

	return nil
}

// forceUpdateDataDir 强制更新数据目录
func forceUpdateDataDir(config *internal.Config, oldDataDir, newDataDir string) error {
	confirm, err := promptUser("\n⚠️  确认强制更改数据目录? 这将导致当前数据无法访问 (输入 'CONFIRM' 确认): ")
	if err != nil {
		return err
	}

	if strings.TrimSpace(confirm) != "CONFIRM" {
		return fmt.Errorf("用户取消了强制更改")
	}

	// 更新配置
	config.DataDir = newDataDir
	updateDataDirHistory(config, oldDataDir)

	fmt.Printf("⚠️  数据目录已强制更改为: %s\n", newDataDir)
	fmt.Printf("💡 原数据目录 '%s' 的数据仍然存在，可以手动恢复\n", oldDataDir)

	return nil
}

// promptUser 提示用户输入
func promptUser(prompt string) (string, error) {
	fmt.Print(prompt)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text(), nil
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("读取用户输入失败")
}

// copyDir 复制目录
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算目标路径
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		// 复制文件
		return copyFile(path, dstPath)
	})
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	// 获取源文件信息并设置权限
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	// 复制内容
	buf := make([]byte, 64*1024) // 64KB buffer
	for {
		n, err := sourceFile.Read(buf)
		if n > 0 {
			if _, writeErr := destFile.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return err
		}
	}

	// 设置文件权限
	return os.Chmod(dst, sourceInfo.Mode())
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
