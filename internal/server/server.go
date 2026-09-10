package server

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/Abbas-Shahneh/taskManager/internal/handler"
	"github.com/Abbas-Shahneh/taskManager/internal/metrics"
	"github.com/Abbas-Shahneh/taskManager/internal/middleware"
	"github.com/Abbas-Shahneh/taskManager/internal/service"
)

func New(
	port int,
	taskService service.TaskService,
	logger *slog.Logger,
) *http.Server {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Tracing())

	appMetrics := metrics.New()

	registry := prometheus.NewRegistry()

	if err := appMetrics.Register(registry); err != nil {
		panic(fmt.Errorf("register metrics: %w", err))
	}

	router.Use(middleware.RequestLogger(logger))
	router.Use(middleware.Metrics(appMetrics))

	router.GET("/health", healthHandler)

	router.GET(
		"/metrics",
		gin.WrapH(promhttp.HandlerFor(
			registry,
			promhttp.HandlerOpts{},
		)),
	)

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
