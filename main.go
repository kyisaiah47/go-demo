package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Task represents a task in our system
type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title" validate:"required,min=1,max=100"`
	Description string    `json:"description" validate:"required,min=1,max=500"`
	Priority    string    `json:"priority" validate:"required,oneof=low medium high"`
	Status      string    `json:"status" validate:"oneof=pending in-progress completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateTaskRequest represents the request payload for creating a task
type CreateTaskRequest struct {
	Title       string `json:"title" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"required,min=1,max=500"`
	Priority    string `json:"priority" validate:"required,oneof=low medium high"`
	Status      string `json:"status,omitempty" validate:"omitempty,oneof=pending in-progress completed"`
}

// UpdateTaskRequest represents the request payload for updating a task
type UpdateTaskRequest struct {
	Title       *string `json:"title,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,min=1,max=500"`
	Priority    *string `json:"priority,omitempty" validate:"omitempty,oneof=low medium high"`
	Status      *string `json:"status,omitempty" validate:"omitempty,oneof=pending in-progress completed"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// In-memory storage (like our other examples)
var tasks = make(map[string]*Task)
var validator_instance = validator.New()

// Initialize with sample data
func init() {
	sampleTask := &Task{
		ID:          uuid.New().String(),
		Title:       "Sample Task",
		Description: "This is a sample task",
		Priority:    "high",
		Status:      "pending",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	tasks[sampleTask.ID] = sampleTask
}

// Validation helper function
func validateStruct(s interface{}) error {
	return validator_instance.Struct(s)
}

// Health check endpoint
func healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	})
}

// Get all tasks
func getTasks(c *gin.Context) {
	taskList := make([]*Task, 0, len(tasks))
	for _, task := range tasks {
		taskList = append(taskList, task)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"tasks": taskList,
		"count": len(taskList),
	})
}

// Get a single task by ID
func getTask(c *gin.Context) {
	id := c.Param("id")
	
	task, exists := tasks[id]
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Task not found",
			Message: "Task with ID " + id + " does not exist",
		})
		return
	}
	
	c.JSON(http.StatusOK, task)
}

// Create a new task
func createTask(c *gin.Context) {
	var req CreateTaskRequest
	
	// Bind JSON to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid JSON",
			Message: err.Error(),
		})
		return
	}
	
	// Validate the request
	if err := validateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}
	
	// Set default status if not provided
	status := req.Status
	if status == "" {
		status = "pending"
	}
	
	// Create new task
	task := &Task{
		ID:          uuid.New().String(),
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		Status:      status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// Store the task
	tasks[task.ID] = task
	
	c.JSON(http.StatusCreated, task)
}

// Update an existing task
func updateTask(c *gin.Context) {
	id := c.Param("id")
	
	// Check if task exists
	task, exists := tasks[id]
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Task not found",
			Message: "Task with ID " + id + " does not exist",
		})
		return
	}
	
	var req UpdateTaskRequest
	
	// Bind JSON to struct
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid JSON",
			Message: err.Error(),
		})
		return
	}
	
	// Validate the request
	if err := validateStruct(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Validation failed",
			Message: err.Error(),
		})
		return
	}
	
	// Update fields if provided (partial update)
	if req.Title != nil {
		task.Title = *req.Title
	}
	if req.Description != nil {
		task.Description = *req.Description
	}
	if req.Priority != nil {
		task.Priority = *req.Priority
	}
	if req.Status != nil {
		task.Status = *req.Status
	}
	
	// Update timestamp
	task.UpdatedAt = time.Now()
	
	c.JSON(http.StatusOK, task)
}

// Delete a task
func deleteTask(c *gin.Context) {
	id := c.Param("id")
	
	// Check if task exists
	_, exists := tasks[id]
	if !exists {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Task not found",
			Message: "Task with ID " + id + " does not exist",
		})
		return
	}
	
	// Delete the task
	delete(tasks, id)
	
	c.JSON(http.StatusNoContent, nil)
}

// Get tasks statistics
func getTaskStats(c *gin.Context) {
	stats := map[string]int{
		"total":       len(tasks),
		"pending":     0,
		"in-progress": 0,
		"completed":   0,
		"low":         0,
		"medium":      0,
		"high":        0,
	}
	
	for _, task := range tasks {
		stats[task.Status]++
		stats[task.Priority]++
	}
	
	c.JSON(http.StatusOK, gin.H{
		"statistics": stats,
		"timestamp":  time.Now(),
	})
}

// CORS middleware
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// Logging middleware
func loggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

func main() {
	// Create Gin router
	r := gin.New()
	
	// Add middleware
	r.Use(loggingMiddleware())
	r.Use(corsMiddleware())
	r.Use(gin.Recovery())
	
	// Welcome endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Welcome to Go Task Management API",
			"version": "1.0.0",
			"docs":    "Visit /health for health check",
		})
	})
	
	// Health check
	r.GET("/health", healthCheck)
	
	// API routes group
	api := r.Group("/api")
	{
		// Task routes
		tasks := api.Group("/tasks")
		{
			tasks.GET("", getTasks)           // GET /api/tasks
			tasks.POST("", createTask)        // POST /api/tasks
			tasks.GET("/:id", getTask)        // GET /api/tasks/:id
			tasks.PUT("/:id", updateTask)     // PUT /api/tasks/:id
			tasks.DELETE("/:id", deleteTask)  // DELETE /api/tasks/:id
		}
		
		// Statistics route
		api.GET("/stats", getTaskStats)  // GET /api/stats
	}
	
	// Start server
	port := ":8080"
	fmt.Printf("🚀 Server starting on http://localhost%s\n", port)
	fmt.Printf("📚 API endpoints available at http://localhost%s/api/tasks\n", port)
	fmt.Printf("💚 Health check at http://localhost%s/health\n", port)
	
	// This blocks until the server stops
	if err := r.Run(port); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}