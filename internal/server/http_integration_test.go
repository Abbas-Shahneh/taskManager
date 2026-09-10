package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/Abbas-Shahneh/taskManager/internal/repository"
	"github.com/Abbas-Shahneh/taskManager/internal/service"
)

func setupHTTPIntegrationTest(t *testing.T) (*httptest.Server, *pgxpool.Pool) {
	t.Helper()

	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("TEST_DATABASE_URL is not set; skipping HTTP integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)

	pool, err := pgxpool.New(ctx, databaseURL)
	require.NoError(t, err)

	require.NoError(t, pool.Ping(ctx))

	t.Cleanup(pool.Close)

	_, err = pool.Exec(ctx, `TRUNCATE TABLE tasks`)
	require.NoError(t, err)

	taskRepository := repository.NewPostgresTaskRepository(pool)
	taskService := service.NewTaskService(taskRepository)

	gin.SetMode(gin.TestMode)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	httpServer := New(8080, taskService, logger)

	testServer := httptest.NewServer(httpServer.Handler)

	t.Cleanup(func() {
		testServer.Close()
	})

	return testServer, pool
}

func TestHTTPIntegration_CreateAndGetTask(t *testing.T) {
	testServer, _ := setupHTTPIntegrationTest(t)

	payload := map[string]any{
		"title":       "HTTP integration test",
		"description": "created through the complete HTTP stack",
		"status":      "pending",
		"assignee":    "alice",
	}

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	response, err := http.Post(
		testServer.URL+"/api/v1/tasks",
		"application/json",
		bytes.NewReader(body),
	)
	require.NoError(t, err)
	defer response.Body.Close()

	require.Equal(t, http.StatusCreated, response.StatusCode)

	var createResponse struct {
		Task struct {
			ID          string  `json:"id"`
			Title       string  `json:"title"`
			Description *string `json:"description"`
			Status      string  `json:"status"`
			Assignee    *string `json:"assignee"`
		} `json:"task"`
	}

	require.NoError(
		t,
		json.NewDecoder(response.Body).Decode(&createResponse),
	)

	require.NotEmpty(t, createResponse.Task.ID)
	require.Equal(t, "HTTP integration test", createResponse.Task.Title)
	require.Equal(t, "pending", createResponse.Task.Status)

	getResponse, err := http.Get(
		testServer.URL + "/api/v1/tasks/" + createResponse.Task.ID,
	)
	require.NoError(t, err)
	defer getResponse.Body.Close()

	require.Equal(t, http.StatusOK, getResponse.StatusCode)

	var getResult struct {
		Task struct {
			ID     string `json:"id"`
			Title  string `json:"title"`
			Status string `json:"status"`
		} `json:"task"`
	}

	require.NoError(
		t,
		json.NewDecoder(getResponse.Body).Decode(&getResult),
	)

	require.Equal(t, createResponse.Task.ID, getResult.Task.ID)
	require.Equal(t, "HTTP integration test", getResult.Task.Title)
	require.Equal(t, "pending", getResult.Task.Status)
}

func TestHTTPIntegration_ListTasksWithFiltersAndPagination(t *testing.T) {
	testServer, _ := setupHTTPIntegrationTest(t)

	tasks := []map[string]any{
		{
			"title":    "Pending Alice",
			"status":   "pending",
			"assignee": "alice",
		},
		{
			"title":    "Completed Alice",
			"status":   "completed",
			"assignee": "alice",
		},
		{
			"title":    "Pending Bob",
			"status":   "pending",
			"assignee": "bob",
		},
	}

	for _, task := range tasks {
		body, err := json.Marshal(task)
		require.NoError(t, err)

		response, err := http.Post(
			testServer.URL+"/api/v1/tasks",
			"application/json",
			bytes.NewReader(body),
		)
		require.NoError(t, err)
		response.Body.Close()

		require.Equal(t, http.StatusCreated, response.StatusCode)
	}

	response, err := http.Get(
		testServer.URL + "/api/v1/tasks?status=pending&assignee=alice&page=1&page_size=1",
	)
	require.NoError(t, err)
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)

	var result struct {
		Tasks []struct {
			Title  string `json:"title"`
			Status string `json:"status"`
		} `json:"tasks"`
		Total      int `json:"total"`
		Page       int `json:"page"`
		PageSize   int `json:"page_size"`
		TotalPages int `json:"total_pages"`
	}

	require.NoError(t, json.NewDecoder(response.Body).Decode(&result))

	require.Len(t, result.Tasks, 1)
	require.Equal(t, 1, result.Total)
	require.Equal(t, 1, result.Page)
	require.Equal(t, 1, result.PageSize)
	require.Equal(t, 1, result.TotalPages)

	require.Equal(t, "Pending Alice", result.Tasks[0].Title)
	require.Equal(t, "pending", result.Tasks[0].Status)
}

func TestHTTPIntegration_UpdateAndDeleteTask(t *testing.T) {
	testServer, _ := setupHTTPIntegrationTest(t)

	createPayload := map[string]any{
		"title":    "Task before update",
		"status":   "pending",
		"assignee": "alice",
	}

	body, err := json.Marshal(createPayload)
	require.NoError(t, err)

	createResponse, err := http.Post(
		testServer.URL+"/api/v1/tasks",
		"application/json",
		bytes.NewReader(body),
	)
	require.NoError(t, err)
	defer createResponse.Body.Close()

	require.Equal(t, http.StatusCreated, createResponse.StatusCode)

	var createResult struct {
		Task struct {
			ID string `json:"id"`
		} `json:"task"`
	}

	require.NoError(
		t,
		json.NewDecoder(createResponse.Body).Decode(&createResult),
	)

	taskID := createResult.Task.ID
	require.NotEmpty(t, taskID)

	updatePayload := map[string]any{
		"title":    "Task after update",
		"status":   "in_progress",
		"assignee": "bob",
	}

	updateBody, err := json.Marshal(updatePayload)
	require.NoError(t, err)

	request, err := http.NewRequest(
		http.MethodPut,
		testServer.URL+"/api/v1/tasks/"+taskID,
		bytes.NewReader(updateBody),
	)
	require.NoError(t, err)

	request.Header.Set("Content-Type", "application/json")

	updateResponse, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer updateResponse.Body.Close()

	require.Equal(t, http.StatusOK, updateResponse.StatusCode)

	var updateResult struct {
		Task struct {
			ID       string `json:"id"`
			Title    string `json:"title"`
			Status   string `json:"status"`
			Assignee string `json:"assignee"`
		} `json:"task"`
	}

	require.NoError(
		t,
		json.NewDecoder(updateResponse.Body).Decode(&updateResult),
	)

	require.Equal(t, taskID, updateResult.Task.ID)
	require.Equal(t, "Task after update", updateResult.Task.Title)
	require.Equal(t, "in_progress", updateResult.Task.Status)
	require.Equal(t, "bob", updateResult.Task.Assignee)

	deleteRequest, err := http.NewRequest(
		http.MethodDelete,
		testServer.URL+"/api/v1/tasks/"+taskID,
		nil,
	)
	require.NoError(t, err)

	deleteResponse, err := http.DefaultClient.Do(deleteRequest)
	require.NoError(t, err)
	defer deleteResponse.Body.Close()

	require.Equal(t, http.StatusNoContent, deleteResponse.StatusCode)

	getResponse, err := http.Get(
		testServer.URL + "/api/v1/tasks/" + taskID,
	)
	require.NoError(t, err)
	defer getResponse.Body.Close()

	require.Equal(t, http.StatusNotFound, getResponse.StatusCode)
}

func TestHTTPIntegration_ValidationAndErrors(t *testing.T) {
	testServer, _ := setupHTTPIntegrationTest(t)

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{
			name:       "invalid JSON",
			method:     http.MethodPost,
			path:       "/api/v1/tasks",
			body:       `{"title":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid UUID",
			method:     http.MethodGet,
			path:       "/api/v1/tasks/not-a-uuid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing task",
			method:     http.MethodGet,
			path:       "/api/v1/tasks/00000000-0000-0000-0000-000000000001",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid task",
			method:     http.MethodPost,
			path:       "/api/v1/tasks",
			body:       `{"title":"","status":"pending"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(
				tt.method,
				testServer.URL+tt.path,
				bytes.NewBufferString(tt.body),
			)
			require.NoError(t, err)

			if tt.body != "" {
				request.Header.Set("Content-Type", "application/json")
			}

			response, err := http.DefaultClient.Do(request)
			require.NoError(t, err)
			defer response.Body.Close()

			require.Equal(t, tt.wantStatus, response.StatusCode)
		})
	}
}

func TestHTTPIntegration_HealthAndRequestID(t *testing.T) {
	testServer, _ := setupHTTPIntegrationTest(t)

	request, err := http.NewRequest(
		http.MethodGet,
		testServer.URL+"/health",
		nil,
	)
	require.NoError(t, err)

	request.Header.Set("X-Request-ID", "integration-test-request-id")

	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)
	require.Equal(
		t,
		"integration-test-request-id",
		response.Header.Get("X-Request-ID"),
	)

	var result map[string]string
	require.NoError(t, json.NewDecoder(response.Body).Decode(&result))

	require.Equal(t, "ok", result["status"])
}
