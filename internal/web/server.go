package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/zoyopei/envswitch/internal"
	"github.com/zoyopei/envswitch/internal/file"
	"github.com/zoyopei/envswitch/internal/project"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

//go:embed templates/*
var templateFS embed.FS

//go:embed static/*
var staticFS embed.FS

type Server struct {
	projectManager *project.Manager
	fileManager    *file.Manager
	upgrader       websocket.Upgrader
}

// NewServer 创建新的Web服务器实例
func NewServer() *Server {
	return &Server{
		projectManager: project.NewManager(),
		fileManager:    file.NewManager(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(_ *http.Request) bool {
				return true // 在生产环境中应该有更严格的检查
			},
		},
	}
}

// SetupRoutes 设置路由
func (s *Server) SetupRoutes() *gin.Engine {
	// 在生产环境中设置为release模式
	// gin.SetMode(gin.ReleaseMode)

	r := gin.Default()

	// 使用嵌入的静态文件系统，需要去掉前缀
	staticFiles, _ := fs.Sub(staticFS, "static")
	r.StaticFS("/static", http.FS(staticFiles))

	// 使用嵌入的模板文件系统
	tmpl := template.Must(template.New("").ParseFS(templateFS, "templates/*"))
	r.SetHTMLTemplate(tmpl)

	// 页面路由
	r.GET("/", s.indexHandler)
	r.GET("/projects", s.projectsPageHandler)
	r.GET("/projects/:id", s.projectDetailPageHandler)
	r.GET("/environments/:id", s.environmentDetailPageHandler)

	// API路由
	api := r.Group("/api")
	{
		// 项目相关API
		projects := api.Group("/projects")
		{
			projects.GET("", s.listProjectsAPI)
			projects.POST("", s.createProjectAPI)
			projects.GET("/:id", s.getProjectAPI)
			projects.PUT("/:id", s.updateProjectAPI)
			projects.DELETE("/:id", s.deleteProjectAPI)

			// 项目下的环境
			projects.GET("/:id/environments", s.listEnvironmentsAPI)
			projects.POST("/:id/environments", s.createEnvironmentAPI)
		}

		// 环境相关API
		environments := api.Group("/environments")
		{
			environments.GET("/:id", s.getEnvironmentAPI)
			environments.PUT("/:id", s.updateEnvironmentAPI)
			environments.DELETE("/:id", s.deleteEnvironmentAPI)

			// 环境下的文件配置
			environments.POST("/:id/files", s.addFileConfigAPI)
		}

		// 文件配置相关API
		api.PUT("/files/:id", s.updateFileConfigAPI)
		api.DELETE("/files/:id", s.deleteFileConfigAPI)

		// 切换相关API
		api.POST("/switch", s.switchEnvironmentAPI)
		api.GET("/status", s.getStatusAPI)
		api.POST("/rollback", s.rollbackAPI)
	}

	// WebSocket
	r.GET("/ws", s.websocketHandler)

	return r
}

// 获取状态信息的辅助函数
func (s *Server) getStatusData() gin.H {
	state, err := s.fileManager.GetCurrentState()
	if err != nil {
		return gin.H{
			"current_project":     "",
			"current_environment": "",
			"last_switch_at":     "",
			"has_active_env":     false,
		}
	}
	
	currentProjectName := state.CurrentProject
	currentEnvironmentName := state.CurrentEnvironment
	
	// 如果有当前项目ID，尝试获取项目名称
	if state.CurrentProject != "" {
		if project, err := s.projectManager.GetProject(state.CurrentProject); err == nil {
			currentProjectName = project.Name
		}
	}
	
	// 如果有当前环境ID，尝试获取环境名称
	if state.CurrentProject != "" && state.CurrentEnvironment != "" {
		if project, err := s.projectManager.GetProject(state.CurrentProject); err == nil {
			for _, env := range project.Environments {
				if env.ID == state.CurrentEnvironment {
					currentEnvironmentName = env.Name
					break
				}
			}
		}
	}
	
	return gin.H{
		"current_project":     currentProjectName,
		"current_environment": currentEnvironmentName,
		"last_switch_at":     state.LastSwitchAt,
		"has_active_env":     state.CurrentProject != "" && state.CurrentEnvironment != "",
	}
}

// 页面处理器
func (s *Server) indexHandler(c *gin.Context) {
	status := s.getStatusData()
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":  "envswitch - Environment Management",
		"status": status,
	})
}

func (s *Server) projectsPageHandler(c *gin.Context) {
	projects, err := s.projectManager.ListProjects()
	if err != nil {
		status := s.getStatusData()
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error":  err.Error(),
			"status": status,
		})
		return
	}

	status := s.getStatusData()
	c.HTML(http.StatusOK, "projects.html", gin.H{
		"title":    "Projects",
		"projects": projects,
		"status":   status,
	})
}

func (s *Server) projectDetailPageHandler(c *gin.Context) {
	projectID := c.Param("id")

	project, err := s.projectManager.GetProject(projectID)
	if err != nil {
		status := s.getStatusData()
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error":  "Project not found",
			"status": status,
		})
		return
	}

	status := s.getStatusData()
	
	// 获取当前激活的环境ID
	currentEnvID := ""
	if state, err := s.fileManager.GetCurrentState(); err == nil {
		if state.CurrentProject == projectID {
			currentEnvID = state.CurrentEnvironment
		}
	}
	
	c.HTML(http.StatusOK, "project_detail.html", gin.H{
		"title":        "Project: " + project.Name,
		"project":      project,
		"status":       status,
		"current_env":  currentEnvID,
	})
}

func (s *Server) environmentDetailPageHandler(c *gin.Context) {
	envID := c.Param("id")

	// 需要通过项目来查找环境，这里简化处理
	// 在实际应用中，可能需要更复杂的查询逻辑
	projects, err := s.projectManager.ListProjects()
	if err != nil {
		status := s.getStatusData()
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error":  err.Error(),
			"status": status,
		})
		return
	}

	var targetEnv *internal.Environment
	var targetProject *internal.Project

	for _, project := range projects {
		for _, env := range project.Environments {
			if env.ID == envID {
				targetEnv = &env
				targetProject = &project
				break
			}
		}
		if targetEnv != nil {
			break
		}
	}

	if targetEnv == nil {
		status := s.getStatusData()
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error":  "Environment not found",
			"status": status,
		})
		return
	}

	status := s.getStatusData()
	
	// 获取当前激活的环境ID
	currentEnvID := ""
	if state, err := s.fileManager.GetCurrentState(); err == nil {
		if state.CurrentProject == targetProject.ID {
			currentEnvID = state.CurrentEnvironment
		}
	}
	
	c.HTML(http.StatusOK, "environment_detail.html", gin.H{
		"title":        "Environment: " + targetEnv.Name,
		"project":      targetProject,
		"environment":  targetEnv,
		"status":       status,
		"current_env":  currentEnvID,
	})
}

// WebSocket处理器
func (s *Server) websocketHandler(c *gin.Context) {
	conn, err := s.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer func() { _ = conn.Close() }()

	// WebSocket连接处理逻辑
	for {
		// 读取消息
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// 回声消息（简单实现）
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			break
		}
	}
}
