package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Abbas-Shahneh/taskManager/internal/domain"
	"github.com/Abbas-Shahneh/taskManager/internal/metrics"
	"github.com/Abbas-Shahneh/taskManager/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	taskID := uuid.New()

	task := domain.Task{
		ID:     taskID,
		Title:  "Test task",
		Status: domain.TaskStatusPending,
	}

	mockService := &mockTaskService{
		CreateFunc: func(
			_ context.Context,
			input service.CreateTaskInput,
		) (domain.Task, error) {
			assert.Equal(t, "Test task", input.Title)

			return task, nil
		},
	}

	router := gin.New()

	appMetrics := metrics.New()

	handler := NewTaskHandler(
		mockService,
		appMetrics,
	)

	router.POST("/tasks", handler.Create)

	body := `{
		"title": "Test task"
	}`

	request := httptest.NewRequest(
		http.MethodPost,
		"/tasks",
		bytes.NewBufferString(body),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(
		t,
		http.StatusCreated,
		recorder.Code,
	)

	var response taskResponse

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	require.NoError(t, err)

	assert.Equal(t, taskID, response.Task.ID)
	assert.Equal(t, "Test task", response.Task.Title)
	assert.Equal(
		t,
		domain.TaskStatusPending,
		response.Task.Status,
	)
}

func TestTaskHandler_Create_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockTaskService{
		CreateFunc: func(
			context.Context,
			service.CreateTaskInput,
		) (domain.Task, error) {
			t.Fatal("service should not be called")
			return domain.Task{}, nil
		},
	}

	router := gin.New()

	appMetrics := metrics.New()

	handler := NewTaskHandler(
		mockService,
		appMetrics,
	)

	router.POST("/tasks", handler.Create)

	request := httptest.NewRequest(
		http.MethodPost,
		"/tasks",
		bytes.NewBufferString(`{"title":`),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(
		t,
		http.StatusBadRequest,
		recorder.Code,
	)
}

func TestTaskHandler_GetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	taskID := uuid.New()

	task := domain.Task{
		ID:     taskID,
		Title:  "Find me",
		Status: domain.TaskStatusPending,
	}

	mockService := &mockTaskService{
		GetByIDFunc: func(
			_ context.Context,
			id uuid.UUID,
		) (domain.Task, error) {
			assert.Equal(t, taskID, id)

			return task, nil
		},
	}

	router := gin.New()

	appMetrics := metrics.New()

	handler := NewTaskHandler(
		mockService,
		appMetrics,
	)

	router.GET("/tasks/:id", handler.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/tasks/"+taskID.String(),
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	var response taskResponse

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	require.NoError(t, err)

	assert.Equal(t, taskID, response.Task.ID)
	assert.Equal(t, "Find me", response.Task.Title)
}

func TestTaskHandler_GetByID_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockTaskService{
		GetByIDFunc: func(
			context.Context,
			uuid.UUID,
		) (domain.Task, error) {
			t.Fatal("service should not be called")
			return domain.Task{}, nil
		},
	}

	router := gin.New()

	appMetrics := metrics.New()

	handler := NewTaskHandler(
		mockService,
		appMetrics,
	)

	router.GET("/tasks/:id", handler.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/tasks/not-a-uuid",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(
		t,
		http.StatusBadRequest,
		recorder.Code,
	)
}

func TestTaskHandler_GetByID_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	taskID := uuid.New()

	mockService := &mockTaskService{
		GetByIDFunc: func(
			context.Context,
			uuid.UUID,
		) (domain.Task, error) {
			return domain.Task{}, domain.ErrTaskNotFound
		},
	}

	router := gin.New()

	appMetrics := metrics.New()

	handler := NewTaskHandler(
		mockService,
		appMetrics,
	)

	router.GET("/tasks/:id", handler.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/tasks/"+taskID.String(),
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(
		t,
		http.StatusNotFound,
		recorder.Code,
	)
}

func TestTaskHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockTaskService{
		ListFunc: func(
			_ context.Context,
			params service.ListTasksParams,
		) (service.ListTasksResult, error) {
			assert.Equal(t, 2, params.Page)
			assert.Equal(t, 10, params.PageSize)

			return service.ListTasksResult{
				Tasks: []domain.Task{
					{
						ID:     uuid.New(),
						Title:  "Task 1",
						Status: domain.TaskStatusPending,
					},
				},
				Total:      21,
				Page:       2,
				PageSize:   10,
				TotalPages: 3,
			}, nil
		},
	}

	router := gin.New()

	appMetrics := metrics.New()

	handler := NewTaskHandler(
		mockService,
		appMetrics,
	)

	router.GET("/tasks", handler.List)

	request := httptest.NewRequest(
		http.MethodGet,
		"/tasks?page=2&page_size=10",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	var response listTasksResponse

	err := json.Unmarshal(
		recorder.Body.Bytes(),
		&response,
	)

	require.NoError(t, err)

	assert.Equal(t, 21, response.Total)
	assert.Equal(t, 2, response.Page)
	assert.Equal(t, 10, response.PageSize)
	assert.Equal(t, 3, response.TotalPages)
	assert.Len(t, response.Tasks, 1)
}

func TestTaskHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	taskID := uuid.New()

	deleted := false

	mockService := &mockTaskService{
		DeleteFunc: func(
			_ context.Context,
			id uuid.UUID,
		) error {
			assert.Equal(t, taskID, id)

			deleted = true

			return nil
		},
	}

	router := gin.New()

	appMetrics := metrics.New()

	handler := NewTaskHandler(
		mockService,
		appMetrics,
	)

	router.DELETE("/tasks/:id", handler.Delete)

	request := httptest.NewRequest(
		http.MethodDelete,
		"/tasks/"+taskID.String(),
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	assert.Equal(
		t,
		http.StatusNoContent,
		recorder.Code,
	)

	assert.True(t, deleted)
}

func TestTaskHandler_GetByID_InvalidUUID_ErrorContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockTaskService{
		GetByIDFunc: func(
			context.Context,
			uuid.UUID,
		) (domain.Task, error) {
			t.Fatal("service should not be called")
			return domain.Task{}, nil
		},
	}

	router := gin.New()
	appMetrics := metrics.New()

	handler := NewTaskHandler(
		mockService,
		appMetrics,
	)

	router.GET("/tasks/:id", handler.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/tasks/not-a-uuid",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusBadRequest, recorder.Code)

	var response errorResponse

	require.NoError(
		t,
		json.Unmarshal(
			recorder.Body.Bytes(),
			&response,
		),
	)

	assert.Equal(
		t,
		"INVALID_TASK_ID",
		response.Error.Code,
	)

	assert.Equal(
		t,
		"invalid task id",
		response.Error.Message,
	)
}

func TestTaskHandler_GetByID_NotFound_ErrorContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	taskID := uuid.New()

	mockService := &mockTaskService{
		GetByIDFunc: func(
			context.Context,
			uuid.UUID,
		) (domain.Task, error) {
			return domain.Task{}, domain.ErrTaskNotFound
		},
	}

	router := gin.New()
	appMetrics := metrics.New()

	handler := NewTaskHandler(
		mockService,
		appMetrics,
	)

	router.GET("/tasks/:id", handler.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/tasks/"+taskID.String(),
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNotFound, recorder.Code)

	var response errorResponse

	require.NoError(
		t,
		json.Unmarshal(
			recorder.Body.Bytes(),
			&response,
		),
	)

	assert.Equal(
		t,
		"TASK_NOT_FOUND",
		response.Error.Code,
	)

	assert.Equal(
		t,
		"task not found",
		response.Error.Message,
	)
}

func TestTaskHandler_GetByID_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	taskID := uuid.New()

	mockService := &mockTaskService{
		GetByIDFunc: func(
			context.Context,
			uuid.UUID,
		) (domain.Task, error) {
			return domain.Task{}, errors.New("database connection failed")
		},
	}

	router := gin.New()
	appMetrics := metrics.New()

	handler := NewTaskHandler(
		mockService,
		appMetrics,
	)

	router.GET("/tasks/:id", handler.GetByID)

	request := httptest.NewRequest(
		http.MethodGet,
		"/tasks/"+taskID.String(),
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(
		t,
		http.StatusInternalServerError,
		recorder.Code,
	)

	var response errorResponse

	require.NoError(
		t,
		json.Unmarshal(
			recorder.Body.Bytes(),
			&response,
		),
	)

	assert.Equal(
		t,
		"INTERNAL_SERVER_ERROR",
		response.Error.Code,
	)

	assert.Equal(
		t,
		"internal server error",
		response.Error.Message,
	)

	assert.NotContains(
		t,
		recorder.Body.String(),
		"database connection failed",
	)
}
