package web

import (
	"net/http"

	"github.com/zoyopei/envswitch/internal"

	"github.com/gin-gonic/gin"
)

// 项目相关API

func (s *Server) listProjectsAPI(c *gin.Context) {
	projects, err := s.projectManager.ListProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"projects": projects,
	})
}

func (s *Server) createProjectAPI(c *gin.Context) {
	var request struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	project, err := s.projectManager.CreateProject(request.Name, request.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, project)
}

func (s *Server) getProjectAPI(c *gin.Context) {
	projectID := c.Param("id")

	project, err := s.projectManager.GetProject(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Project not found",
		})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (s *Server) updateProjectAPI(c *gin.Context) {
	projectID := c.Param("id")

	var request struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	updates := make(map[string]interface{})
	if request.Name != "" {
		updates["name"] = request.Name
	}
	if request.Description != "" {
		updates["description"] = request.Description
	}

	project, err := s.projectManager.UpdateProject(projectID, updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, project)
}

func (s *Server) deleteProjectAPI(c *gin.Context) {
	projectID := c.Param("id")

	err := s.projectManager.DeleteProject(projectID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project deleted successfully",
	})
}

// 环境相关API

func (s *Server) listEnvironmentsAPI(c *gin.Context) {
	projectID := c.Param("id")

	environments, err := s.projectManager.ListEnvironments(projectID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"environments": environments,
	})
}

func (s *Server) createEnvironmentAPI(c *gin.Context) {
	projectID := c.Param("id")

	var request struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	env := &internal.Environment{
		Name:        request.Name,
		Description: request.Description,
		Tags:        request.Tags,
		Files:       []internal.FileConfig{},
	}

	err := s.projectManager.AddEnvironment(projectID, env)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, env)
}

func (s *Server) getEnvironmentAPI(c *gin.Context) {
	envID := c.Param("id")

	// 查找环境（需要遍历所有项目）
	projects, err := s.projectManager.ListProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	for _, project := range projects {
		for _, env := range project.Environments {
			if env.ID == envID {
				c.JSON(http.StatusOK, env)
				return
			}
		}
	}

	c.JSON(http.StatusNotFound, gin.H{
		"error": "Environment not found",
	})
}

func (s *Server) updateEnvironmentAPI(c *gin.Context) {
	envID := c.Param("id")

	var request struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 找到环境所属的项目
	projects, err := s.projectManager.ListProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var projectID string
	for _, project := range projects {
		for _, env := range project.Environments {
			if env.ID == envID {
				projectID = project.ID
				break
			}
		}
		if projectID != "" {
			break
		}
	}

	if projectID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Environment not found",
		})
		return
	}

	updates := make(map[string]interface{})
	if request.Name != "" {
		updates["name"] = request.Name
	}
	if request.Description != "" {
		updates["description"] = request.Description
	}
	if request.Tags != nil {
		updates["tags"] = request.Tags
	}

	env, err := s.projectManager.UpdateEnvironment(projectID, envID, updates)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, env)
}

func (s *Server) deleteEnvironmentAPI(c *gin.Context) {
	envID := c.Param("id")

	// 找到环境所属的项目
	projects, err := s.projectManager.ListProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var projectID string
	for _, project := range projects {
		for _, env := range project.Environments {
			if env.ID == envID {
				projectID = project.ID
				break
			}
		}
		if projectID != "" {
			break
		}
	}

	if projectID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Environment not found",
		})
		return
	}

	err = s.projectManager.RemoveEnvironment(projectID, envID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Environment deleted successfully",
	})
}

// 文件配置相关API

func (s *Server) addFileConfigAPI(c *gin.Context) {
	envID := c.Param("id")

	var request struct {
		SourcePath  string `json:"source_path" binding:"required"`
		TargetPath  string `json:"target_path" binding:"required"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// 找到环境所属的项目
	projects, err := s.projectManager.ListProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var projectID string
	for _, project := range projects {
		for _, env := range project.Environments {
			if env.ID == envID {
				projectID = project.ID
				break
			}
		}
		if projectID != "" {
			break
		}
	}

	if projectID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Environment not found",
		})
		return
	}

	err = s.fileManager.AddFileConfig(projectID, envID, request.SourcePath, request.TargetPath, request.Description)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "File configuration added successfully",
	})
}

func (s *Server) updateFileConfigAPI(c *gin.Context) {
	_ = c.Param("id") // fileID，暂时未使用

	// 这里需要实现文件配置更新逻辑
	// 为简化，现在返回未实现
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "File config update not implemented yet",
	})
}

func (s *Server) deleteFileConfigAPI(c *gin.Context) {
	fileID := c.Param("id")

	// 找到文件配置所属的项目和环境
	projects, err := s.projectManager.ListProjects()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	var projectID, envID string
	for _, project := range projects {
		for _, env := range project.Environments {
			for _, file := range env.Files {
				if file.ID == fileID {
					projectID = project.ID
					envID = env.ID
					break
				}
			}
			if projectID != "" {
				break
			}
		}
		if projectID != "" {
			break
		}
	}

	if projectID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File configuration not found",
		})
		return
	}

	err = s.fileManager.RemoveFileConfig(projectID, envID, fileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File configuration deleted successfully",
	})
}

// 切换相关API

func (s *Server) switchEnvironmentAPI(c *gin.Context) {
	var request struct {
		ProjectID     string `json:"project_id" binding:"required"`
		EnvironmentID string `json:"environment_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err := s.fileManager.SwitchEnvironment(request.ProjectID, request.EnvironmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Environment switched successfully",
	})
}

func (s *Server) getStatusAPI(c *gin.Context) {
	state, err := s.fileManager.GetCurrentState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, state)
}

func (s *Server) rollbackAPI(c *gin.Context) {
	var request struct {
		BackupID string `json:"backup_id"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var backupID string
	if request.BackupID != "" {
		backupID = request.BackupID
	} else {
		// 使用当前状态中的备份ID
		state, err := s.fileManager.GetCurrentState()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		backupID = state.BackupID
	}

	if backupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "No backup ID provided and no backup available",
		})
		return
	}

	err := s.fileManager.RollbackFromBackup(backupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rollback completed successfully",
	})
}
