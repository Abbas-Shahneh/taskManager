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
	dto "github.com/prometheus/client_model/go"
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

func TestTaskHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	taskID := uuid.New()

	description := "updated description"
	assignee := "bob"

	task := domain.Task{
		ID:          taskID,
		Title:       "Updated task",
		Description: &description,
		Status:      domain.TaskStatusInProgress,
		Assignee:    &assignee,
	}

	mockService := &mockTaskService{
		UpdateFunc: func(
			_ context.Context,
			id uuid.UUID,
			input service.UpdateTaskInput,
		) (domain.Task, error) {
			assert.Equal(t, taskID, id)
			assert.Equal(t, "Updated task", input.Title)
			assert.Equal(
				t,
				domain.TaskStatusInProgress,
				input.Status,
			)

			return task, nil
		},
	}

	router := gin.New()
	handler := NewTaskHandler(mockService, metrics.New())

	router.PUT("/tasks/:id", handler.Update)

	body := `{
		"title": "Updated task",
		"description": "updated description",
		"status": "in_progress",
		"assignee": "bob"
	}`

	request := httptest.NewRequest(
		http.MethodPut,
		"/tasks/"+taskID.String(),
		bytes.NewBufferString(body),
	)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)

	var response taskResponse
	require.NoError(
		t,
		json.Unmarshal(recorder.Body.Bytes(), &response),
	)

	assert.Equal(t, taskID, response.Task.ID)
	assert.Equal(t, "Updated task", response.Task.Title)
	assert.Equal(
		t,
		domain.TaskStatusInProgress,
		response.Task.Status,
	)
}

func TestTaskHandler_Update_InvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockTaskService{
		UpdateFunc: func(
			context.Context,
			uuid.UUID,
			service.UpdateTaskInput,
		) (domain.Task, error) {
			t.Fatal("service should not be called")
			return domain.Task{}, nil
		},
	}

	router := gin.New()
	handler := NewTaskHandler(mockService, metrics.New())

	router.PUT("/tasks/:id", handler.Update)

	request := httptest.NewRequest(
		http.MethodPut,
		"/tasks/not-a-uuid",
		bytes.NewBufferString(`{"title":"test"}`),
	)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusBadRequest, recorder.Code)

	var response errorResponse
	require.NoError(
		t,
		json.Unmarshal(recorder.Body.Bytes(), &response),
	)

	assert.Equal(t, "INVALID_TASK_ID", response.Error.Code)
}

func TestTaskHandler_Update_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockService := &mockTaskService{
		UpdateFunc: func(
			context.Context,
			uuid.UUID,
			service.UpdateTaskInput,
		) (domain.Task, error) {
			t.Fatal("service should not be called")
			return domain.Task{}, nil
		},
	}

	router := gin.New()
	handler := NewTaskHandler(mockService, metrics.New())

	router.PUT("/tasks/:id", handler.Update)

	request := httptest.NewRequest(
		http.MethodPut,
		"/tasks/"+uuid.New().String(),
		bytes.NewBufferString(`{"title":`),
	)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusBadRequest, recorder.Code)

	var response errorResponse
	require.NoError(
		t,
		json.Unmarshal(recorder.Body.Bytes(), &response),
	)

	assert.Equal(t, "INVALID_REQUEST", response.Error.Code)
}

func TestTaskHandler_Update_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	taskID := uuid.New()

	mockService := &mockTaskService{
		UpdateFunc: func(
			context.Context,
			uuid.UUID,
			service.UpdateTaskInput,
		) (domain.Task, error) {
			return domain.Task{}, domain.ErrTaskNotFound
		},
	}

	router := gin.New()
	handler := NewTaskHandler(mockService, metrics.New())

	router.PUT("/tasks/:id", handler.Update)

	request := httptest.NewRequest(
		http.MethodPut,
		"/tasks/"+taskID.String(),
		bytes.NewBufferString(`{"title":"test","status":"pending"}`),
	)
	request.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNotFound, recorder.Code)

	var response errorResponse
	require.NoError(
		t,
		json.Unmarshal(recorder.Body.Bytes(), &response),
	)

	assert.Equal(t, "TASK_NOT_FOUND", response.Error.Code)
}

func TestTaskHandler_List_InvalidQueryParameters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		query    string
		wantCode int
		wantErr  string
	}{
		{
			name:     "invalid status",
			query:    "?status=invalid",
			wantCode: http.StatusBadRequest,
			wantErr:  "INVALID_TASK",
		},
		{
			name:     "invalid page",
			query:    "?page=abc",
			wantCode: http.StatusBadRequest,
			wantErr:  "INVALID_PAGE",
		},
		{
			name:     "invalid page size",
			query:    "?page_size=abc",
			wantCode: http.StatusBadRequest,
			wantErr:  "INVALID_PAGE_SIZE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &mockTaskService{
				ListFunc: func(
					context.Context,
					service.ListTasksParams,
				) (service.ListTasksResult, error) {
					t.Fatal("service should not be called")
					return service.ListTasksResult{}, nil
				},
			}

			router := gin.New()
			handler := NewTaskHandler(mockService, metrics.New())

			router.GET("/tasks", handler.List)

			request := httptest.NewRequest(
				http.MethodGet,
				"/tasks"+tt.query,
				nil,
			)

			recorder := httptest.NewRecorder()

			router.ServeHTTP(recorder, request)

			require.Equal(t, tt.wantCode, recorder.Code)

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
				tt.wantErr,
				response.Error.Code,
			)
		})
	}
}

func TestTaskHandler_ServiceErrorMapping(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		err      error
		wantCode int
		wantErr  string
	}{
		{
			name:     "task not found",
			err:      domain.ErrTaskNotFound,
			wantCode: http.StatusNotFound,
			wantErr:  "TASK_NOT_FOUND",
		},
		{
			name:     "invalid title",
			err:      domain.ErrInvalidTaskTitle,
			wantCode: http.StatusBadRequest,
			wantErr:  "INVALID_TASK",
		},
		{
			name:     "invalid description",
			err:      domain.ErrInvalidDescription,
			wantCode: http.StatusBadRequest,
			wantErr:  "INVALID_TASK",
		},
		{
			name:     "invalid assignee",
			err:      domain.ErrInvalidAssignee,
			wantCode: http.StatusBadRequest,
			wantErr:  "INVALID_TASK",
		},
		{
			name:     "invalid status",
			err:      domain.ErrInvalidTaskStatus,
			wantCode: http.StatusBadRequest,
			wantErr:  "INVALID_TASK",
		},
		{
			name:     "invalid task id",
			err:      service.ErrInvalidTaskID,
			wantCode: http.StatusBadRequest,
			wantErr:  "INVALID_TASK_ID",
		},
		{
			name:     "invalid page",
			err:      service.ErrInvalidPage,
			wantCode: http.StatusBadRequest,
			wantErr:  "INVALID_PAGE",
		},
		{
			name:     "invalid page size",
			err:      service.ErrInvalidPageSize,
			wantCode: http.StatusBadRequest,
			wantErr:  "INVALID_PAGE_SIZE",
		},
		{
			name:     "unknown error",
			err:      errors.New("unexpected failure"),
			wantCode: http.StatusInternalServerError,
			wantErr:  "INTERNAL_SERVER_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(recorder)

			writeServiceError(c, tt.err)

			require.Equal(t, tt.wantCode, recorder.Code)

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
				tt.wantErr,
				response.Error.Code,
			)
		})
	}
}

func TestTaskHandler_RefreshTaskCount(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("updates metric", func(t *testing.T) {
		serviceMock := &mockTaskService{
			CountFunc: func(ctx context.Context) (int, error) {
				return 42, nil
			},
		}

		appMetrics := metrics.New()
		h := NewTaskHandler(serviceMock, appMetrics)

		router := gin.New()
		router.GET("/test", func(c *gin.Context) {
			h.refreshTaskCount(c)
			c.Status(http.StatusNoContent)
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		metric := &dto.Metric{}
		err := appMetrics.TasksCount.Write(metric)
		require.NoError(t, err)

		assert.Equal(t, float64(42), metric.GetGauge().GetValue())
	})
}
