package api

import (
	"net/http"
	"strconv"

	"flowctl/internal/core"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Server struct {
	scheduler *core.Scheduler
	logger    *logrus.Logger
	router    *gin.Engine
}

func NewServer(scheduler *core.Scheduler, logger *logrus.Logger) *Server {
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())

	server := &Server{
		scheduler: scheduler,
		logger:    logger,
		router:    router,
	}

	server.setupRoutes()
	return server
}

func (s *Server) setupRoutes() {
	api := s.router.Group("/api/v1")
	
	api.POST("/workflows", s.createWorkflow)
	api.GET("/workflows/:id", s.getWorkflow)
	api.PUT("/workflows/:id/cancel", s.cancelWorkflow)
	api.GET("/workflows", s.listWorkflows)
	
	api.GET("/tasks/:id", s.getTask)
	api.GET("/workflows/:id/tasks", s.getWorkflowTasks)
	
	api.GET("/health", s.healthCheck)
	api.GET("/metrics", s.getMetrics)

	s.router.Static("/static", "./web/dashboard/build/static")
	s.router.StaticFile("/", "./web/dashboard/build/index.html")
	s.router.NoRoute(func(c *gin.Context) {
		c.File("./web/dashboard/build/index.html")
	})
}

type CreateWorkflowRequest struct {
	Name        string                   `json:"name" binding:"required"`
	Description string                   `json:"description"`
	Tasks       []CreateTaskRequest      `json:"tasks" binding:"required"`
	Config      *core.WorkflowConfig     `json:"config,omitempty"`
}

type CreateTaskRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Type         string                 `json:"type" binding:"required"`
	Payload      map[string]interface{} `json:"payload"`
	MaxRetries   int                    `json:"max_retries,omitempty"`
	Priority     int                    `json:"priority,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
}

func (s *Server) createWorkflow(c *gin.Context) {
	var req CreateWorkflowRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	workflow := core.NewWorkflow(req.Name, req.Description)
	if req.Config != nil {
		workflow.Config = *req.Config
	}

	for _, taskReq := range req.Tasks {
		task := core.NewTask(workflow.ID, taskReq.Name, taskReq.Type, taskReq.Payload)
		
		if taskReq.MaxRetries > 0 {
			task.MaxRetries = taskReq.MaxRetries
		}
		if taskReq.Priority > 0 {
			task.Priority = taskReq.Priority
		}
		if taskReq.Dependencies != nil {
			task.Dependencies = taskReq.Dependencies
		}
		
		workflow.Tasks = append(workflow.Tasks, *task)
	}

	if err := s.scheduler.SubmitWorkflow(c.Request.Context(), workflow); err != nil {
		s.logger.Errorf("Failed to submit workflow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow"})
		return
	}

	c.JSON(http.StatusCreated, workflow)
}

func (s *Server) getWorkflow(c *gin.Context) {
	workflowID := c.Param("id")
	
	workflow, err := s.scheduler.GetWorkflow(workflowID)
	if err != nil {
		s.logger.Errorf("Failed to get workflow %s: %v", workflowID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Workflow not found"})
		return
	}

	c.JSON(http.StatusOK, workflow)
}

func (s *Server) cancelWorkflow(c *gin.Context) {
	workflowID := c.Param("id")
	
	if err := s.scheduler.CancelWorkflow(c.Request.Context(), workflowID); err != nil {
		s.logger.Errorf("Failed to cancel workflow %s: %v", workflowID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel workflow"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Workflow cancelled"})
}

func (s *Server) listWorkflows(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	status := c.Query("status")
	
	_ = page
	_ = limit
	_ = status
	
	c.JSON(http.StatusOK, gin.H{
		"workflows": []core.Workflow{},
		"total":     0,
		"page":      page,
		"limit":     limit,
	})
}

func (s *Server) getTask(c *gin.Context) {
	taskID := c.Param("id")
	
	task, err := s.scheduler.GetTask(taskID)
	if err != nil {
		s.logger.Errorf("Failed to get task %s: %v", taskID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

func (s *Server) getWorkflowTasks(c *gin.Context) {
	workflowID := c.Param("id")
	
	tasks, err := s.scheduler.GetWorkflowTasks(workflowID)
	if err != nil {
		s.logger.Errorf("Failed to get tasks for workflow %s: %v", workflowID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get tasks"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

func (s *Server) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": "2024-01-01T00:00:00Z",
	})
}

func (s *Server) getMetrics(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"workflows": gin.H{
			"total":     0,
			"running":   0,
			"completed": 0,
			"failed":    0,
		},
		"tasks": gin.H{
			"total":     0,
			"pending":   0,
			"running":   0,
			"completed": 0,
			"failed":    0,
		},
		"workers": gin.H{
			"active": 0,
			"idle":   0,
		},
	})
}

func (s *Server) Start(addr string) error {
	s.logger.Infof("Starting API server on %s", addr)
	return s.router.Run(addr)
}
