package internal

import (
	"time"
)

// Project 项目结构
type Project struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Environments []Environment `json:"environments"`
}

// Environment 环境结构
type Environment struct {
	ID           string       `json:"id"`
	Name         string       `json:"name"`
	Description  string       `json:"description"`
	Tags         []string     `json:"tags"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
	LastSwitchAt *time.Time   `json:"last_switch_at,omitempty"`
	Files        []FileConfig `json:"files"`
}

// FileConfig 文件配置结构
type FileConfig struct {
	ID          string `json:"id"`
	SourcePath  string `json:"source_path"` // 模板文件路径
	TargetPath  string `json:"target_path"` // 目标替换路径
	BackupPath  string `json:"backup_path"` // 备份文件路径
	Description string `json:"description"`
}

// Config 全局配置结构
type Config struct {
	DataDir            string   `json:"data_dir"`
	BackupDir          string   `json:"backup_dir"`
	WebPort            int      `json:"web_port"`
	DefaultProject     string   `json:"default_project"`
	OriginalDataDir    string   `json:"original_data_dir,omitempty"` // 原始数据目录路径
	DataDirHistory     []string `json:"data_dir_history,omitempty"`  // 历史数据目录记录
	EnableDataDirCheck bool     `json:"enable_data_dir_check"`       // 是否启用数据目录变更检查
}

// AppState 应用状态
type AppState struct {
	CurrentProject     string     `json:"current_project"`
	CurrentEnvironment string     `json:"current_environment"`
	LastSwitchAt       *time.Time `json:"last_switch_at,omitempty"`
	BackupID           string     `json:"backup_id,omitempty"`
}

// SwitchRequest 切换请求
type SwitchRequest struct {
	ProjectID     string `json:"project_id"`
	EnvironmentID string `json:"environment_id"`
}

// BackupInfo 备份信息
type BackupInfo struct {
	ID        string            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	Files     map[string]string `json:"files"` // target_path -> backup_path
	ProjectID string            `json:"project_id"`
	EnvID     string            `json:"env_id"`
}
