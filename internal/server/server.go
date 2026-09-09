package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Abbas-Shahneh/taskManager/internal/handler"
	"github.com/Abbas-Shahneh/taskManager/internal/middleware"
	"github.com/Abbas-Shahneh/taskManager/internal/service"
)

func New(
	port int,
	taskService service.TaskService,
) *http.Server {
	router := gin.New()

	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())

	router.GET("/health", healthHandler)

	api := router.Group("/api/v1")

	taskHandler := handler.NewTaskHandler(taskService)
	taskHandler.RegisterRoutes(api)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
}

func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
