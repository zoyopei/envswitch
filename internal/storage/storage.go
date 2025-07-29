package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zoyopei/envswitch/internal"
	"github.com/zoyopei/envswitch/internal/config"
)

type Storage struct {
	dataDir string
}

// NewStorage 创建新的存储实例
func NewStorage() *Storage {
	return &Storage{
		dataDir: config.GetDataDir(),
	}
}

// SaveProject 保存项目
func (s *Storage) SaveProject(project *internal.Project) error {
	projectsDir := filepath.Join(s.dataDir, "projects")
	if err := os.MkdirAll(projectsDir, 0755); err != nil {
		return fmt.Errorf("failed to create projects directory: %w", err)
	}

	filename := fmt.Sprintf("%s.json", project.ID)
	filepath := filepath.Join(projectsDir, filename)

	data, err := json.MarshalIndent(project, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal project: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write project file: %w", err)
	}

	return nil
}

// LoadProject 加载项目
func (s *Storage) LoadProject(projectID string) (*internal.Project, error) {
	filename := fmt.Sprintf("%s.json", projectID)
	filepath := filepath.Join(s.dataDir, "projects", filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("project not found: %s", projectID)
		}
		return nil, fmt.Errorf("failed to read project file: %w", err)
	}

	var project internal.Project
	if err := json.Unmarshal(data, &project); err != nil {
		return nil, fmt.Errorf("failed to parse project file: %w", err)
	}

	return &project, nil
}

// LoadProjectByName 通过名称加载项目
func (s *Storage) LoadProjectByName(name string) (*internal.Project, error) {
	projects, err := s.ListProjects()
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		if project.Name == name {
			return &project, nil
		}
	}

	return nil, fmt.Errorf("project not found: %s", name)
}

// ListProjects 列出所有项目
func (s *Storage) ListProjects() ([]internal.Project, error) {
	projectsDir := filepath.Join(s.dataDir, "projects")

	files, err := os.ReadDir(projectsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []internal.Project{}, nil
		}
		return nil, fmt.Errorf("failed to read projects directory: %w", err)
	}

	var projects []internal.Project
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		projectID := file.Name()[:len(file.Name())-5] // 移除.json扩展名
		project, err := s.LoadProject(projectID)
		if err != nil {
			// 跳过无法加载的项目文件，但记录错误
			fmt.Printf("Warning: failed to load project %s: %v\n", projectID, err)
			continue
		}

		projects = append(projects, *project)
	}

	return projects, nil
}

// DeleteProject 删除项目
func (s *Storage) DeleteProject(projectID string) error {
	filename := fmt.Sprintf("%s.json", projectID)
	filepath := filepath.Join(s.dataDir, "projects", filename)

	if err := os.Remove(filepath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("project not found: %s", projectID)
		}
		return fmt.Errorf("failed to delete project file: %w", err)
	}

	return nil
}

// SaveAppState 保存应用状态
func (s *Storage) SaveAppState(state *internal.AppState) error {
	filepath := filepath.Join(s.dataDir, "state.json")

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal app state: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write app state file: %w", err)
	}

	return nil
}

// LoadAppState 加载应用状态
func (s *Storage) LoadAppState() (*internal.AppState, error) {
	filepath := filepath.Join(s.dataDir, "state.json")

	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			// 返回默认状态
			return &internal.AppState{}, nil
		}
		return nil, fmt.Errorf("failed to read app state file: %w", err)
	}

	var state internal.AppState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse app state file: %w", err)
	}

	return &state, nil
}

// SaveBackupInfo 保存备份信息
func (s *Storage) SaveBackupInfo(backup *internal.BackupInfo) error {
	backupDir := config.GetBackupDir()
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	filename := fmt.Sprintf("%s.json", backup.ID)
	filepath := filepath.Join(backupDir, filename)

	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal backup info: %w", err)
	}

	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write backup info file: %w", err)
	}

	return nil
}

// LoadBackupInfo 加载备份信息
func (s *Storage) LoadBackupInfo(backupID string) (*internal.BackupInfo, error) {
	filename := fmt.Sprintf("%s.json", backupID)
	filepath := filepath.Join(config.GetBackupDir(), filename)

	data, err := os.ReadFile(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("backup not found: %s", backupID)
		}
		return nil, fmt.Errorf("failed to read backup info file: %w", err)
	}

	var backup internal.BackupInfo
	if err := json.Unmarshal(data, &backup); err != nil {
		return nil, fmt.Errorf("failed to parse backup info file: %w", err)
	}

	return &backup, nil
}

// ListBackups 列出所有备份
func (s *Storage) ListBackups() ([]internal.BackupInfo, error) {
	backupDir := config.GetBackupDir()

	files, err := os.ReadDir(backupDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []internal.BackupInfo{}, nil
		}
		return nil, fmt.Errorf("failed to read backup directory: %w", err)
	}

	var backups []internal.BackupInfo
	for _, file := range files {
		if filepath.Ext(file.Name()) != ".json" {
			continue
		}

		backupID := file.Name()[:len(file.Name())-5] // 移除.json扩展名
		backup, err := s.LoadBackupInfo(backupID)
		if err != nil {
			// 跳过无法加载的备份文件
			continue
		}

		backups = append(backups, *backup)
	}

	return backups, nil
}

// DeleteBackup 删除备份
func (s *Storage) DeleteBackup(backupID string) error {
	// 删除备份信息文件
	infoFile := fmt.Sprintf("%s.json", backupID)
	infoPath := filepath.Join(config.GetBackupDir(), infoFile)

	// 先获取备份信息以清理备份文件
	backup, err := s.LoadBackupInfo(backupID)
	if err != nil {
		return err
	}

	// 删除所有备份的文件
	for _, backupPath := range backup.Files {
		if err := os.Remove(backupPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: failed to delete backup file %s: %v\n", backupPath, err)
		}
	}

	// 删除备份信息文件
	if err := os.Remove(infoPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("backup not found: %s", backupID)
		}
		return fmt.Errorf("failed to delete backup info file: %w", err)
	}

	return nil
}

// CleanupOldBackups 清理旧备份（保留最近N个）
func (s *Storage) CleanupOldBackups(keepCount int) error {
	backups, err := s.ListBackups()
	if err != nil {
		return err
	}

	if len(backups) <= keepCount {
		return nil
	}

	// 按时间排序，最新的在前
	for i := 0; i < len(backups)-1; i++ {
		for j := i + 1; j < len(backups); j++ {
			if backups[i].Timestamp.Before(backups[j].Timestamp) {
				backups[i], backups[j] = backups[j], backups[i]
			}
		}
	}

	// 删除超出保留数量的备份
	for i := keepCount; i < len(backups); i++ {
		if err := s.DeleteBackup(backups[i].ID); err != nil {
			fmt.Printf("Warning: failed to delete old backup %s: %v\n", backups[i].ID, err)
		}
	}

	return nil
}
