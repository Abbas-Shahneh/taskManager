package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Abbas-Shahneh/taskManager/internal/domain"
	"github.com/Abbas-Shahneh/taskManager/internal/metrics"
	"github.com/Abbas-Shahneh/taskManager/internal/service"
)

type TaskHandler struct {
	service service.TaskService
	metrics *metrics.Metrics
}

func NewTaskHandler(
	taskService service.TaskService,
	appMetrics *metrics.Metrics,
) *TaskHandler {
	return &TaskHandler{
		service: taskService,
		metrics: appMetrics,
	}
}

// RegisterRoutes registers all task-related routes.
func (h *TaskHandler) RegisterRoutes(router *gin.RouterGroup) {
	tasks := router.Group("/tasks")

	tasks.POST("", h.Create)
	tasks.GET("", h.List)
	tasks.GET("/:id", h.GetByID)
	tasks.PUT("/:id", h.Update)
	tasks.DELETE("/:id", h.Delete)
}

// Create handles POST /tasks.
func (h *TaskHandler) Create(c *gin.Context) {
	var request createTaskRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request body",
		)
		return
	}

	input := service.CreateTaskInput{
		Title:       request.Title,
		Description: request.Description,
		Status:      request.Status,
		Assignee:    request.Assignee,
	}

	task, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	h.refreshTaskCount(c)

	c.JSON(http.StatusCreated, taskResponse{
		Task: task,
	})
}

// GetByID handles GET /tasks/:id.
func (h *TaskHandler) GetByID(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}

	task, err := h.service.GetByID(
		c.Request.Context(),
		id,
	)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, taskResponse{
		Task: task,
	})
}

// List handles GET /tasks.
func (h *TaskHandler) List(c *gin.Context) {
	params, err := parseListTasksParams(c)
	if err != nil {
		writeListParamsError(c, err)
		return
	}

	result, err := h.service.List(
		c.Request.Context(),
		params,
	)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, listTasksResponse{
		Tasks:      result.Tasks,
		Total:      result.Total,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	})
}

// Update handles PUT /tasks/:id.
func (h *TaskHandler) Update(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}

	var request updateTaskRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid request body",
		)
		return
	}

	input := service.UpdateTaskInput{
		Title:       request.Title,
		Description: request.Description,
		Status:      request.Status,
		Assignee:    request.Assignee,
	}

	task, err := h.service.Update(
		c.Request.Context(),
		id,
		input,
	)
	if err != nil {
		writeServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, taskResponse{
		Task: task,
	})
}

// Delete handles DELETE /tasks/:id.
func (h *TaskHandler) Delete(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		return
	}

	if err := h.service.Delete(
		c.Request.Context(),
		id,
	); err != nil {
		writeServiceError(c, err)
		return
	}

	h.refreshTaskCount(c)

	c.Status(http.StatusNoContent)
}

type createTaskRequest struct {
	Title       string            `json:"title"`
	Description *string           `json:"description"`
	Status      domain.TaskStatus `json:"status"`
	Assignee    *string           `json:"assignee"`
}

type updateTaskRequest struct {
	Title       string            `json:"title"`
	Description *string           `json:"description"`
	Status      domain.TaskStatus `json:"status"`
	Assignee    *string           `json:"assignee"`
}

type taskResponse struct {
	Task domain.Task `json:"task"`
}

type listTasksResponse struct {
	Tasks      []domain.Task `json:"tasks"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

type errorResponse struct {
	Error errorBody `json:"error"`
}

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func parseUUIDParam(
	c *gin.Context,
	name string,
) (uuid.UUID, error) {
	value := strings.TrimSpace(c.Param(name))

	id, err := uuid.Parse(value)
	if err != nil {
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_TASK_ID",
			"invalid task id",
		)

		return uuid.Nil, service.ErrInvalidTaskID
	}

	return id, nil
}

func parseListTasksParams(
	c *gin.Context,
) (service.ListTasksParams, error) {
	params := service.ListTasksParams{}

	statusValue := strings.TrimSpace(
		c.Query("status"),
	)

	if statusValue != "" {
		status := domain.TaskStatus(statusValue)

		if !domain.IsValidTaskStatus(status) {
			return service.ListTasksParams{},
				domain.ErrInvalidTaskStatus
		}

		params.Status = &status
	}

	assigneeValue := strings.TrimSpace(
		c.Query("assignee"),
	)

	if assigneeValue != "" {
		params.Assignee = &assigneeValue
	}

	pageValue := strings.TrimSpace(
		c.Query("page"),
	)

	if pageValue != "" {
		page, err := strconv.Atoi(pageValue)
		if err != nil {
			return service.ListTasksParams{},
				service.ErrInvalidPage
		}

		params.Page = page
	}

	pageSizeValue := strings.TrimSpace(
		c.Query("page_size"),
	)

	if pageSizeValue != "" {
		pageSize, err := strconv.Atoi(pageSizeValue)
		if err != nil {
			return service.ListTasksParams{},
				service.ErrInvalidPageSize
		}

		params.PageSize = pageSize
	}

	return params, nil
}

func writeListParamsError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidTaskStatus):
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_TASK",
			"invalid task status",
		)

	case errors.Is(err, service.ErrInvalidPage):
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_PAGE",
			"invalid page",
		)

	case errors.Is(err, service.ErrInvalidPageSize):
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_PAGE_SIZE",
			"invalid page size",
		)

	default:
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_REQUEST",
			"invalid query parameters",
		)
	}
}

func writeServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrTaskNotFound):
		writeError(
			c,
			http.StatusNotFound,
			"TASK_NOT_FOUND",
			"task not found",
		)

	case errors.Is(err, domain.ErrInvalidTaskTitle),
		errors.Is(err, domain.ErrInvalidDescription),
		errors.Is(err, domain.ErrInvalidAssignee),
		errors.Is(err, domain.ErrInvalidTaskStatus):
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_TASK",
			"invalid task",
		)

	case errors.Is(err, service.ErrInvalidTaskID):
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_TASK_ID",
			"invalid task id",
		)

	case errors.Is(err, service.ErrInvalidPage):
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_PAGE",
			"invalid page",
		)

	case errors.Is(err, service.ErrInvalidPageSize):
		writeError(
			c,
			http.StatusBadRequest,
			"INVALID_PAGE_SIZE",
			"invalid page size",
		)

	default:
		writeError(
			c,
			http.StatusInternalServerError,
			"INTERNAL_SERVER_ERROR",
			"internal server error",
		)
	}
}

func writeError(
	c *gin.Context,
	statusCode int,
	code string,
	message string,
) {
	c.JSON(statusCode, errorResponse{
		Error: errorBody{
			Code:    code,
			Message: message,
		},
	})
}

func (h *TaskHandler) refreshTaskCount(c *gin.Context) {
	count, err := h.service.Count(c.Request.Context())
	if err != nil {
		return
	}

	h.metrics.SetTasksCount(count)
}
