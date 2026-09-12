package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Abbas-Shahneh/taskManager/internal/cache"
	"github.com/Abbas-Shahneh/taskManager/internal/config"
	"github.com/Abbas-Shahneh/taskManager/internal/database"
	"github.com/Abbas-Shahneh/taskManager/internal/repository"
	"github.com/Abbas-Shahneh/taskManager/internal/server"
	"github.com/Abbas-Shahneh/taskManager/internal/service"
	"github.com/Abbas-Shahneh/taskManager/internal/tracing"
	"github.com/redis/go-redis/v9"
)

func main() {
	logger := slog.New(
		slog.NewJSONHandler(
			os.Stdout,
			&slog.HandlerOptions{
				Level: slog.LevelInfo,
			},
		),
	)

	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error(
			"failed to load configuration",
			"error",
			err,
		)
		os.Exit(1)
	}

	var redisClient *redis.Client

	if cfg.Redis.Enabled {
		redisClient = redis.NewClient(&redis.Options{
			Addr: fmt.Sprintf(
				"%s:%d",
				cfg.Redis.Host,
				cfg.Redis.Port,
			),
		})

		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			logger.Error(
				"redis unavailable, continuing without cache",
				"error",
				err,
			)

			_ = redisClient.Close()
			redisClient = nil
		}
	}

	if redisClient != nil {
		defer redisClient.Close()
	}

	shutdownTracing := tracing.Init()
	defer shutdownTracing(context.Background())

	ctx := context.Background()

	db, err := database.NewPostgresPool(
		ctx,
		cfg.Database,
	)
	if err != nil {
		logger.Error(
			"failed to connect to database",
			"error",
			err,
		)
		os.Exit(1)
	}
	defer db.Close()

	taskRepository := repository.NewPostgresTaskRepository(db)

	var taskService service.TaskService

	if redisClient != nil {
		taskCache := cache.NewRedisTaskListCache(redisClient)

		taskService = service.NewTaskServiceWithCache(
			taskRepository,
			taskCache,
			time.Duration(cfg.Redis.CacheTTL)*time.Second,
		)
	} else {
		taskService = service.NewTaskService(taskRepository)
	}

	httpServer := server.New(
		cfg.Port,
		taskService,
		logger,
	)

	serverErrors := make(chan error, 1)

	go func() {
		logger.Info(
			"starting HTTP server",
			"port",
			cfg.Port,
			"environment",
			cfg.AppEnv,
		)

		if err := httpServer.ListenAndServe(); err != nil &&
			!errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
		}
	}()

	shutdownSignals := make(chan os.Signal, 1)

	signal.Notify(
		shutdownSignals,
		syscall.SIGINT,
		syscall.SIGTERM,
	)

	select {
	case err := <-serverErrors:
		logger.Error(
			"HTTP server failed",
			"error",
			err,
		)
		os.Exit(1)

	case signal := <-shutdownSignals:
		logger.Info(
			"shutdown signal received",
			"signal",
			signal,
		)
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error(
			"HTTP server shutdown failed",
			"error",
			err,
		)
		os.Exit(1)
	}

	logger.Info("HTTP server stopped")
}
