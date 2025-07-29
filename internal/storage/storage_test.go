package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/zoyopei/EnvSwitch/internal"
	"github.com/zoyopei/EnvSwitch/internal/config"
)

func setupStorageTest(t *testing.T) *Storage {
	tempDir := t.TempDir()

	// 保存原始配置并在测试结束后恢复
	originalConfig := config.GetConfig()
	t.Cleanup(func() {
		_ = config.SaveConfig(originalConfig)
	})

	// 设置临时配置，使用临时目录中的路径
	testConfig := &internal.Config{
		DataDir:   filepath.Join(tempDir, "data"),
		BackupDir: filepath.Join(tempDir, "backups"),
		WebPort:   8080,
	}

	err := config.SaveConfig(testConfig)
	if err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}

	// 确保目录存在
	_ = os.MkdirAll(testConfig.DataDir, 0755)
	_ = os.MkdirAll(testConfig.BackupDir, 0755)

	return NewStorage()
}

func createTestProject() *internal.Project {
	return &internal.Project{
		ID:          "test-project-id",
		Name:        "Test Project",
		Description: "Test project description",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Environments: []internal.Environment{
			{
				ID:          "test-env-id",
				Name:        "test-env",
				Description: "Test environment",
				Tags:        []string{"test", "development"},
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				Files: []internal.FileConfig{
					{
						ID:          "test-file-id",
						SourcePath:  "/source/config.json",
						TargetPath:  "/target/config.json",
						Description: "Test config file",
					},
				},
			},
		},
	}
}

func TestSaveAndLoadProject(t *testing.T) {
	storage := setupStorageTest(t)
	project := createTestProject()

	// 保存项目
	err := storage.SaveProject(project)
	if err != nil {
		t.Fatalf("SaveProject() error = %v", err)
	}

	// 加载项目
	loadedProject, err := storage.LoadProject(project.ID)
	if err != nil {
		t.Fatalf("LoadProject() error = %v", err)
	}

	// 验证项目内容
	if loadedProject.ID != project.ID {
		t.Errorf("Expected project ID = %s, got %s", project.ID, loadedProject.ID)
	}
	if loadedProject.Name != project.Name {
		t.Errorf("Expected project name = %s, got %s", project.Name, loadedProject.Name)
	}
	if len(loadedProject.Environments) != len(project.Environments) {
		t.Errorf("Expected %d environments, got %d", len(project.Environments), len(loadedProject.Environments))
	}

	// 验证环境内容
	if len(loadedProject.Environments) > 0 {
		env := loadedProject.Environments[0]
		originalEnv := project.Environments[0]
		if env.ID != originalEnv.ID {
			t.Errorf("Expected environment ID = %s, got %s", originalEnv.ID, env.ID)
		}
		if len(env.Files) != len(originalEnv.Files) {
			t.Errorf("Expected %d files, got %d", len(originalEnv.Files), len(env.Files))
		}
	}
}

func TestLoadProjectByName(t *testing.T) {
	storage := setupStorageTest(t)
	project := createTestProject()

	// 保存项目
	err := storage.SaveProject(project)
	if err != nil {
		t.Fatalf("SaveProject() error = %v", err)
	}

	// 通过名称加载项目
	loadedProject, err := storage.LoadProjectByName(project.Name)
	if err != nil {
		t.Fatalf("LoadProjectByName() error = %v", err)
	}

	if loadedProject.ID != project.ID {
		t.Errorf("Expected project ID = %s, got %s", project.ID, loadedProject.ID)
	}

	// 测试加载不存在的项目
	_, err = storage.LoadProjectByName("non-existent-project")
	if err == nil {
		t.Error("Expected error when loading non-existent project by name")
	}
}

func TestListProjects(t *testing.T) {
	storage := setupStorageTest(t)

	// 初始时应该没有项目
	projects, err := storage.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("Expected 0 projects initially, got %d", len(projects))
	}

	// 保存几个项目
	for i := 0; i < 3; i++ {
		project := createTestProject()
		project.ID = project.ID + "-" + string(rune('0'+i))
		project.Name = project.Name + " " + string(rune('0'+i))

		err := storage.SaveProject(project)
		if err != nil {
			t.Fatalf("Failed to save project %d: %v", i, err)
		}
	}

	// 列出项目
	projects, err = storage.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(projects) != 3 {
		t.Errorf("Expected 3 projects, got %d", len(projects))
	}
}

func TestDeleteProject(t *testing.T) {
	storage := setupStorageTest(t)
	project := createTestProject()

	// 保存项目
	err := storage.SaveProject(project)
	if err != nil {
		t.Fatalf("SaveProject() error = %v", err)
	}

	// 删除项目
	err = storage.DeleteProject(project.ID)
	if err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}

	// 验证项目已被删除
	_, err = storage.LoadProject(project.ID)
	if err == nil {
		t.Error("Expected error when loading deleted project")
	}

	// 测试删除不存在的项目
	err = storage.DeleteProject("non-existent-id")
	if err == nil {
		t.Error("Expected error when deleting non-existent project")
	}
}

func TestSaveAndLoadAppState(t *testing.T) {
	storage := setupStorageTest(t)

	state := &internal.AppState{
		CurrentProject:     "test-project",
		CurrentEnvironment: "test-env",
		LastSwitchAt:       &time.Time{},
		BackupID:           "test-backup-id",
	}
	now := time.Now()
	state.LastSwitchAt = &now

	// 保存状态
	err := storage.SaveAppState(state)
	if err != nil {
		t.Fatalf("SaveAppState() error = %v", err)
	}

	// 加载状态
	loadedState, err := storage.LoadAppState()
	if err != nil {
		t.Fatalf("LoadAppState() error = %v", err)
	}

	if loadedState.CurrentProject != state.CurrentProject {
		t.Errorf("Expected current project = %s, got %s", state.CurrentProject, loadedState.CurrentProject)
	}
	if loadedState.CurrentEnvironment != state.CurrentEnvironment {
		t.Errorf("Expected current environment = %s, got %s", state.CurrentEnvironment, loadedState.CurrentEnvironment)
	}
	if loadedState.BackupID != state.BackupID {
		t.Errorf("Expected backup ID = %s, got %s", state.BackupID, loadedState.BackupID)
	}
}

func TestSaveAndLoadBackupInfo(t *testing.T) {
	storage := setupStorageTest(t)

	backup := &internal.BackupInfo{
		ID:        "test-backup-id",
		Timestamp: time.Now(),
		Files: map[string]string{
			"/target/config.json": "/backup/config.json.bak",
			"/target/app.conf":    "/backup/app.conf.bak",
		},
		ProjectID: "test-project-id",
		EnvID:     "test-env-id",
	}

	// 保存备份信息
	err := storage.SaveBackupInfo(backup)
	if err != nil {
		t.Fatalf("SaveBackupInfo() error = %v", err)
	}

	// 加载备份信息
	loadedBackup, err := storage.LoadBackupInfo(backup.ID)
	if err != nil {
		t.Fatalf("LoadBackupInfo() error = %v", err)
	}

	if loadedBackup.ID != backup.ID {
		t.Errorf("Expected backup ID = %s, got %s", backup.ID, loadedBackup.ID)
	}
	if loadedBackup.ProjectID != backup.ProjectID {
		t.Errorf("Expected project ID = %s, got %s", backup.ProjectID, loadedBackup.ProjectID)
	}
	if len(loadedBackup.Files) != len(backup.Files) {
		t.Errorf("Expected %d files, got %d", len(backup.Files), len(loadedBackup.Files))
	}

	// 验证文件映射
	for target, backupPath := range backup.Files {
		if loadedBackup.Files[target] != backupPath {
			t.Errorf("Expected backup path for %s = %s, got %s", target, backupPath, loadedBackup.Files[target])
		}
	}
}

func TestListBackups(t *testing.T) {
	storage := setupStorageTest(t)

	// 初始时应该没有备份
	backups, err := storage.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups() error = %v", err)
	}
	if len(backups) != 0 {
		t.Errorf("Expected 0 backups initially, got %d", len(backups))
	}

	// 创建几个备份
	for i := 0; i < 3; i++ {
		backup := &internal.BackupInfo{
			ID:        "backup-" + string(rune('0'+i)),
			Timestamp: time.Now().Add(time.Duration(i) * time.Hour),
			Files:     map[string]string{"/test": "/backup"},
			ProjectID: "project-id",
			EnvID:     "env-id",
		}

		err := storage.SaveBackupInfo(backup)
		if err != nil {
			t.Fatalf("Failed to save backup %d: %v", i, err)
		}
	}

	// 列出备份
	backups, err = storage.ListBackups()
	if err != nil {
		t.Fatalf("ListBackups() error = %v", err)
	}
	if len(backups) != 3 {
		t.Errorf("Expected 3 backups, got %d", len(backups))
	}
}
