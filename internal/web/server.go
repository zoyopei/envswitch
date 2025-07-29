package web

import (
	"embed"
	"html/template"
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

	// 使用嵌入的静态文件系统
	r.StaticFS("/static", http.FS(staticFS))

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

// 页面处理器
func (s *Server) indexHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title": "envswitch - Environment Management",
	})
}

func (s *Server) projectsPageHandler(c *gin.Context) {
	projects, err := s.projectManager.ListProjects()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "projects.html", gin.H{
		"title":    "Projects",
		"projects": projects,
	})
}

func (s *Server) projectDetailPageHandler(c *gin.Context) {
	projectID := c.Param("id")

	project, err := s.projectManager.GetProject(projectID)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Project not found",
		})
		return
	}

	c.HTML(http.StatusOK, "project_detail.html", gin.H{
		"title":   "Project: " + project.Name,
		"project": project,
	})
}

func (s *Server) environmentDetailPageHandler(c *gin.Context) {
	envID := c.Param("id")

	// 需要通过项目来查找环境，这里简化处理
	// 在实际应用中，可能需要更复杂的查询逻辑
	projects, err := s.projectManager.ListProjects()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": err.Error(),
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
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "Environment not found",
		})
		return
	}

	c.HTML(http.StatusOK, "environment_detail.html", gin.H{
		"title":       "Environment: " + targetEnv.Name,
		"project":     targetProject,
		"environment": targetEnv,
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
