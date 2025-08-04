package project

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zoyopei/envswitch/internal"
	"github.com/zoyopei/envswitch/internal/config"
)

func setupTest(t *testing.T) (*Manager, string) {
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

	manager := NewManager()
	return manager, tempDir
}

func TestCreateProject(t *testing.T) {
	manager, _ := setupTest(t)

	// 测试创建项目
	project, err := manager.CreateProject("test-project", "Test project description")
	if err != nil {
		t.Fatalf("CreateProject() error = %v", err)
	}

	if project.Name != "test-project" {
		t.Errorf("Expected project name = test-project, got %s", project.Name)
	}
	if project.Description != "Test project description" {
		t.Errorf("Expected project description = Test project description, got %s", project.Description)
	}
	if project.ID == "" {
		t.Error("Project ID should not be empty")
	}
	if len(project.Environments) != 0 {
		t.Errorf("Expected 0 environments, got %d", len(project.Environments))
	}

	// 测试创建重复项目名称
	_, err = manager.CreateProject("test-project", "Another description")
	if err == nil {
		t.Error("Expected error when creating project with duplicate name")
	}

	// 测试创建空名称项目
	_, err = manager.CreateProject("", "Empty name test")
	if err == nil {
		t.Error("Expected error when creating project with empty name")
	}
}

func TestGetProject(t *testing.T) {
	manager, _ := setupTest(t)

	// 创建测试项目
	originalProject, err := manager.CreateProject("get-test-project", "Get test description")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// 通过ID获取项目
	project, err := manager.GetProject(originalProject.ID)
	if err != nil {
		t.Fatalf("GetProject() by ID error = %v", err)
	}
	if project.ID != originalProject.ID {
		t.Errorf("Expected project ID = %s, got %s", originalProject.ID, project.ID)
	}

	// 通过名称获取项目
	project, err = manager.GetProject("get-test-project")
	if err != nil {
		t.Fatalf("GetProject() by name error = %v", err)
	}
	if project.Name != "get-test-project" {
		t.Errorf("Expected project name = get-test-project, got %s", project.Name)
	}

	// 测试获取不存在的项目
	_, err = manager.GetProject("non-existent-project")
	if err == nil {
		t.Error("Expected error when getting non-existent project")
	}
}

func TestListProjects(t *testing.T) {
	manager, _ := setupTest(t)

	// 初始时应该没有项目
	projects, err := manager.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("Expected 0 projects initially, got %d", len(projects))
	}

	// 创建几个项目
	projectNames := []string{"project1", "project2", "project3"}
	for _, name := range projectNames {
		_, err := manager.CreateProject(name, "Description for "+name)
		if err != nil {
			t.Fatalf("Failed to create project %s: %v", name, err)
		}
	}

	// 再次列出项目
	projects, err = manager.ListProjects()
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if len(projects) != len(projectNames) {
		t.Errorf("Expected %d projects, got %d", len(projectNames), len(projects))
	}

	// 验证项目名称
	foundNames := make(map[string]bool)
	for _, project := range projects {
		foundNames[project.Name] = true
	}
	for _, name := range projectNames {
		if !foundNames[name] {
			t.Errorf("Project %s not found in list", name)
		}
	}
}

func TestUpdateProject(t *testing.T) {
	manager, _ := setupTest(t)

	// 创建测试项目
	project, err := manager.CreateProject("update-test", "Original description")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// 更新项目
	updates := map[string]interface{}{
		"name":        "updated-test",
		"description": "Updated description",
	}

	updatedProject, err := manager.UpdateProject(project.ID, updates)
	if err != nil {
		t.Fatalf("UpdateProject() error = %v", err)
	}

	if updatedProject.Name != "updated-test" {
		t.Errorf("Expected updated name = updated-test, got %s", updatedProject.Name)
	}
	if updatedProject.Description != "Updated description" {
		t.Errorf("Expected updated description = Updated description, got %s", updatedProject.Description)
	}

	// 测试更新为已存在的名称
	_, err = manager.CreateProject("another-project", "Another project")
	if err != nil {
		t.Fatalf("Failed to create another project: %v", err)
	}

	updates["name"] = "another-project"
	_, err = manager.UpdateProject(updatedProject.ID, updates)
	if err == nil {
		t.Error("Expected error when updating to existing project name")
	}
}

func TestDeleteProject(t *testing.T) {
	manager, _ := setupTest(t)

	// 创建测试项目
	project, err := manager.CreateProject("delete-test", "Delete test description")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// 删除项目
	err = manager.DeleteProject(project.ID)
	if err != nil {
		t.Fatalf("DeleteProject() error = %v", err)
	}

	// 验证项目已被删除
	_, err = manager.GetProject(project.ID)
	if err == nil {
		t.Error("Expected error when getting deleted project")
	}

	// 测试删除不存在的项目
	err = manager.DeleteProject("non-existent-id")
	if err == nil {
		t.Error("Expected error when deleting non-existent project")
	}
}

func TestAddEnvironment(t *testing.T) {
	manager, _ := setupTest(t)

	// 创建测试项目
	project, err := manager.CreateProject("env-test", "Environment test project")
	if err != nil {
		t.Fatalf("Failed to create test project: %v", err)
	}

	// 添加环境
	env := &internal.Environment{
		Name:        "test-env",
		Description: "Test environment",
		Tags:        []string{"development", "local"},
		Files:       []internal.FileConfig{},
	}

	err = manager.AddEnvironment(project.ID, env)
	if err != nil {
		t.Fatalf("AddEnvironment() error = %v", err)
	}

	// 验证环境已添加
	updatedProject, err := manager.GetProject(project.ID)
	if err != nil {
		t.Fatalf("Failed to get updated project: %v", err)
	}

	if len(updatedProject.Environments) != 1 {
		t.Errorf("Expected 1 environment, got %d", len(updatedProject.Environments))
	}

	addedEnv := updatedProject.Environments[0]
	if addedEnv.Name != "test-env" {
		t.Errorf("Expected environment name = test-env, got %s", addedEnv.Name)
	}
	if addedEnv.ID == "" {
		t.Error("Environment ID should not be empty")
	}

	// 测试添加重复环境名称
	duplicateEnv := &internal.Environment{
		Name:  "test-env",
		Files: []internal.FileConfig{},
	}
	err = manager.AddEnvironment(project.ID, duplicateEnv)
	if err == nil {
		t.Error("Expected error when adding environment with duplicate name")
	}
}
