package file

import (
	"envswitch/internal"
	"envswitch/internal/config"
	"envswitch/internal/storage"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

type Manager struct {
	storage *storage.Storage
}

// NewManager 创建新的文件管理器
func NewManager() *Manager {
	return &Manager{
		storage: storage.NewStorage(),
	}
}

// SwitchEnvironment 切换到指定环境
func (m *Manager) SwitchEnvironment(projectID, environmentID string) error {
	// 首先创建备份
	backupID, err := m.CreateBackup(projectID, environmentID)
	if err != nil {
		return fmt.Errorf("failed to create backup: %w", err)
	}

	// 加载项目和环境信息
	storage := storage.NewStorage()
	project, err := storage.LoadProject(projectID)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	var environment *internal.Environment
	for _, env := range project.Environments {
		if env.ID == environmentID {
			environment = &env
			break
		}
	}

	if environment == nil {
		return fmt.Errorf("environment not found: %s", environmentID)
	}

	// 执行文件切换
	for _, fileConfig := range environment.Files {
		if err := m.switchFile(&fileConfig); err != nil {
			// 如果切换失败，尝试回滚
			_ = m.RollbackFromBackup(backupID)
			return fmt.Errorf("failed to switch file %s: %w", fileConfig.TargetPath, err)
		}
	}

	// 更新环境的最后切换时间
	now := time.Now()
	environment.LastSwitchAt = &now

	// 保存项目更新
	if err := storage.SaveProject(project); err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	// 更新应用状态
	state := &internal.AppState{
		CurrentProject:     projectID,
		CurrentEnvironment: environmentID,
		LastSwitchAt:       &now,
		BackupID:           backupID,
	}

	if err := storage.SaveAppState(state); err != nil {
		return fmt.Errorf("failed to save app state: %w", err)
	}

	return nil
}

// switchFile 切换单个文件
func (m *Manager) switchFile(fileConfig *internal.FileConfig) error {
	// 检查源文件是否存在
	if _, err := os.Stat(fileConfig.SourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", fileConfig.SourcePath)
	}

	// 确保目标目录存在
	targetDir := filepath.Dir(fileConfig.TargetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	// 复制文件
	if err := m.copyFile(fileConfig.SourcePath, fileConfig.TargetPath); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// copyFile 复制文件
func (m *Manager) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// 复制文件权限
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}

// CreateBackup 创建备份
func (m *Manager) CreateBackup(projectID, environmentID string) (string, error) {
	storage := storage.NewStorage()
	project, err := storage.LoadProject(projectID)
	if err != nil {
		return "", err
	}

	var environment *internal.Environment
	for _, env := range project.Environments {
		if env.ID == environmentID {
			environment = &env
			break
		}
	}

	if environment == nil {
		return "", fmt.Errorf("environment not found: %s", environmentID)
	}

	backupID := uuid.New().String()
	backupDir := filepath.Join(config.GetBackupDir(), backupID)

	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create backup directory: %w", err)
	}

	backupFiles := make(map[string]string)

	// 备份每个目标文件
	for _, fileConfig := range environment.Files {
		if _, err := os.Stat(fileConfig.TargetPath); os.IsNotExist(err) {
			// 目标文件不存在，跳过备份
			continue
		}

		// 生成备份文件路径
		backupFileName := fmt.Sprintf("%s_%s", filepath.Base(fileConfig.TargetPath), uuid.New().String())
		backupFilePath := filepath.Join(backupDir, backupFileName)

		// 复制文件到备份位置
		if err := m.copyFile(fileConfig.TargetPath, backupFilePath); err != nil {
			return "", fmt.Errorf("failed to backup file %s: %w", fileConfig.TargetPath, err)
		}

		backupFiles[fileConfig.TargetPath] = backupFilePath
	}

	// 保存备份信息
	backupInfo := &internal.BackupInfo{
		ID:        backupID,
		Timestamp: time.Now(),
		Files:     backupFiles,
		ProjectID: projectID,
		EnvID:     environmentID,
	}

	if err := storage.SaveBackupInfo(backupInfo); err != nil {
		return "", fmt.Errorf("failed to save backup info: %w", err)
	}

	return backupID, nil
}

// RollbackFromBackup 从备份回滚
func (m *Manager) RollbackFromBackup(backupID string) error {
	storage := storage.NewStorage()
	backup, err := storage.LoadBackupInfo(backupID)
	if err != nil {
		return fmt.Errorf("failed to load backup info: %w", err)
	}

	// 恢复每个文件
	for targetPath, backupPath := range backup.Files {
		if err := m.copyFile(backupPath, targetPath); err != nil {
			return fmt.Errorf("failed to restore file %s: %w", targetPath, err)
		}
	}

	// 更新应用状态，清除当前环境信息
	state := &internal.AppState{}
	if err := storage.SaveAppState(state); err != nil {
		return fmt.Errorf("failed to update app state: %w", err)
	}

	return nil
}

// GetCurrentState 获取当前状态
func (m *Manager) GetCurrentState() (*internal.AppState, error) {
	return m.storage.LoadAppState()
}

// ValidateFileConfig 验证文件配置
func (m *Manager) ValidateFileConfig(fileConfig *internal.FileConfig) error {
	if fileConfig.SourcePath == "" {
		return fmt.Errorf("source path cannot be empty")
	}

	if fileConfig.TargetPath == "" {
		return fmt.Errorf("target path cannot be empty")
	}

	// 检查源文件是否存在
	if _, err := os.Stat(fileConfig.SourcePath); os.IsNotExist(err) {
		return fmt.Errorf("source file does not exist: %s", fileConfig.SourcePath)
	}

	// 检查目标路径是否有效
	targetDir := filepath.Dir(fileConfig.TargetPath)
	if targetDir != "." {
		if err := os.MkdirAll(targetDir, 0755); err != nil {
			return fmt.Errorf("cannot create target directory %s: %w", targetDir, err)
		}
	}

	return nil
}

// AddFileConfig 添加文件配置到环境
func (m *Manager) AddFileConfig(projectID, environmentID, sourcePath, targetPath, description string) error {
	fileConfig := &internal.FileConfig{
		ID:          uuid.New().String(),
		SourcePath:  sourcePath,
		TargetPath:  targetPath,
		Description: description,
	}

	// 验证文件配置
	if err := m.ValidateFileConfig(fileConfig); err != nil {
		return err
	}

	// 加载项目
	project, err := m.storage.LoadProject(projectID)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	// 找到环境
	var envIndex = -1
	for i, env := range project.Environments {
		if env.ID == environmentID {
			envIndex = i
			break
		}
	}

	if envIndex == -1 {
		return fmt.Errorf("environment not found: %s", environmentID)
	}

	// 检查是否已存在相同的目标路径
	for _, existingFile := range project.Environments[envIndex].Files {
		if existingFile.TargetPath == targetPath {
			return fmt.Errorf("file config with target path '%s' already exists", targetPath)
		}
	}

	// 添加文件配置
	project.Environments[envIndex].Files = append(project.Environments[envIndex].Files, *fileConfig)
	project.Environments[envIndex].UpdatedAt = time.Now()
	project.UpdatedAt = time.Now()

	return m.storage.SaveProject(project)
}

// RemoveFileConfig 从环境移除文件配置
func (m *Manager) RemoveFileConfig(projectID, environmentID, fileID string) error {
	project, err := m.storage.LoadProject(projectID)
	if err != nil {
		return fmt.Errorf("failed to load project: %w", err)
	}

	// 找到环境
	var envIndex = -1
	for i, env := range project.Environments {
		if env.ID == environmentID {
			envIndex = i
			break
		}
	}

	if envIndex == -1 {
		return fmt.Errorf("environment not found: %s", environmentID)
	}

	// 找到文件配置
	var fileIndex = -1
	for i, file := range project.Environments[envIndex].Files {
		if file.ID == fileID {
			fileIndex = i
			break
		}
	}

	if fileIndex == -1 {
		return fmt.Errorf("file config not found: %s", fileID)
	}

	// 移除文件配置
	files := project.Environments[envIndex].Files
	project.Environments[envIndex].Files = append(files[:fileIndex], files[fileIndex+1:]...)
	project.Environments[envIndex].UpdatedAt = time.Now()
	project.UpdatedAt = time.Now()

	return m.storage.SaveProject(project)
}

// CleanupOldBackups 清理旧备份
func (m *Manager) CleanupOldBackups(keepCount int) error {
	return m.storage.CleanupOldBackups(keepCount)
}
