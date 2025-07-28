package project

import (
	"envswitch/internal"
	"envswitch/internal/storage"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Manager struct {
	storage *storage.Storage
}

// NewManager 创建新的项目管理器
func NewManager() *Manager {
	return &Manager{
		storage: storage.NewStorage(),
	}
}

// CreateProject 创建新项目
func (m *Manager) CreateProject(name, description string) (*internal.Project, error) {
	if name == "" {
		return nil, fmt.Errorf("project name cannot be empty")
	}

	// 检查项目名称是否已存在
	_, err := m.storage.LoadProjectByName(name)
	if err == nil {
		return nil, fmt.Errorf("project with name '%s' already exists", name)
	}

	project := &internal.Project{
		ID:           uuid.New().String(),
		Name:         name,
		Description:  description,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Environments: []internal.Environment{},
	}

	if err := m.storage.SaveProject(project); err != nil {
		return nil, fmt.Errorf("failed to save project: %w", err)
	}

	return project, nil
}

// GetProject 获取项目（通过ID或名称）
func (m *Manager) GetProject(identifier string) (*internal.Project, error) {
	// 首先尝试通过ID获取
	project, err := m.storage.LoadProject(identifier)
	if err == nil {
		return project, nil
	}

	// 如果通过ID获取失败，尝试通过名称获取
	project, err = m.storage.LoadProjectByName(identifier)
	if err != nil {
		return nil, fmt.Errorf("project not found: %s", identifier)
	}

	return project, nil
}

// ListProjects 列出所有项目
func (m *Manager) ListProjects() ([]internal.Project, error) {
	return m.storage.ListProjects()
}

// UpdateProject 更新项目信息
func (m *Manager) UpdateProject(identifier string, updates map[string]interface{}) (*internal.Project, error) {
	project, err := m.GetProject(identifier)
	if err != nil {
		return nil, err
	}

	// 应用更新
	if name, ok := updates["name"]; ok {
		if nameStr, ok := name.(string); ok && nameStr != "" {
			// 检查新名称是否已被其他项目使用
			existingProject, err := m.storage.LoadProjectByName(nameStr)
			if err == nil && existingProject.ID != project.ID {
				return nil, fmt.Errorf("project with name '%s' already exists", nameStr)
			}
			project.Name = nameStr
		}
	}

	if description, ok := updates["description"]; ok {
		if descStr, ok := description.(string); ok {
			project.Description = descStr
		}
	}

	project.UpdatedAt = time.Now()

	if err := m.storage.SaveProject(project); err != nil {
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	return project, nil
}

// DeleteProject 删除项目
func (m *Manager) DeleteProject(identifier string) error {
	project, err := m.GetProject(identifier)
	if err != nil {
		return err
	}

	return m.storage.DeleteProject(project.ID)
}

// AddEnvironment 向项目添加环境
func (m *Manager) AddEnvironment(projectIdentifier string, env *internal.Environment) error {
	project, err := m.GetProject(projectIdentifier)
	if err != nil {
		return err
	}

	// 检查环境名称是否已存在
	for _, existingEnv := range project.Environments {
		if existingEnv.Name == env.Name {
			return fmt.Errorf("environment with name '%s' already exists in project '%s'", env.Name, project.Name)
		}
	}

	// 确保环境有ID和时间戳
	if env.ID == "" {
		env.ID = uuid.New().String()
	}
	if env.CreatedAt.IsZero() {
		env.CreatedAt = time.Now()
	}
	env.UpdatedAt = time.Now()

	project.Environments = append(project.Environments, *env)
	project.UpdatedAt = time.Now()

	return m.storage.SaveProject(project)
}

// UpdateEnvironment 更新环境
func (m *Manager) UpdateEnvironment(projectIdentifier, envIdentifier string, updates map[string]interface{}) (*internal.Environment, error) {
	project, err := m.GetProject(projectIdentifier)
	if err != nil {
		return nil, err
	}

	// 找到环境
	var envIndex = -1
	for i, env := range project.Environments {
		if env.ID == envIdentifier || env.Name == envIdentifier {
			envIndex = i
			break
		}
	}

	if envIndex == -1 {
		return nil, fmt.Errorf("environment not found: %s", envIdentifier)
	}

	env := &project.Environments[envIndex]

	// 应用更新
	if name, ok := updates["name"]; ok {
		if nameStr, ok := name.(string); ok && nameStr != "" {
			// 检查新名称是否已被其他环境使用
			for i, existingEnv := range project.Environments {
				if i != envIndex && existingEnv.Name == nameStr {
					return nil, fmt.Errorf("environment with name '%s' already exists", nameStr)
				}
			}
			env.Name = nameStr
		}
	}

	if description, ok := updates["description"]; ok {
		if descStr, ok := description.(string); ok {
			env.Description = descStr
		}
	}

	if tags, ok := updates["tags"]; ok {
		if tagsSlice, ok := tags.([]string); ok {
			env.Tags = tagsSlice
		}
	}

	env.UpdatedAt = time.Now()
	project.UpdatedAt = time.Now()

	if err := m.storage.SaveProject(project); err != nil {
		return nil, fmt.Errorf("failed to update environment: %w", err)
	}

	return env, nil
}

// RemoveEnvironment 从项目移除环境
func (m *Manager) RemoveEnvironment(projectIdentifier, envIdentifier string) error {
	project, err := m.GetProject(projectIdentifier)
	if err != nil {
		return err
	}

	// 找到环境
	var envIndex = -1
	for i, env := range project.Environments {
		if env.ID == envIdentifier || env.Name == envIdentifier {
			envIndex = i
			break
		}
	}

	if envIndex == -1 {
		return fmt.Errorf("environment not found: %s", envIdentifier)
	}

	// 移除环境
	project.Environments = append(project.Environments[:envIndex], project.Environments[envIndex+1:]...)
	project.UpdatedAt = time.Now()

	return m.storage.SaveProject(project)
}

// GetEnvironment 获取环境
func (m *Manager) GetEnvironment(projectIdentifier, envIdentifier string) (*internal.Environment, error) {
	project, err := m.GetProject(projectIdentifier)
	if err != nil {
		return nil, err
	}

	for _, env := range project.Environments {
		if env.ID == envIdentifier || env.Name == envIdentifier {
			return &env, nil
		}
	}

	return nil, fmt.Errorf("environment not found: %s", envIdentifier)
}

// ListEnvironments 列出项目的所有环境
func (m *Manager) ListEnvironments(projectIdentifier string) ([]internal.Environment, error) {
	project, err := m.GetProject(projectIdentifier)
	if err != nil {
		return nil, err
	}

	return project.Environments, nil
}
