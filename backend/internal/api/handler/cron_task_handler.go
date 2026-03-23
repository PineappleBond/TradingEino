package handler

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/api/request"
	"github.com/PineappleBond/TradingEino/backend/internal/api/response"
	"github.com/PineappleBond/TradingEino/backend/internal/model"
	"github.com/PineappleBond/TradingEino/backend/internal/repository"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/gin-gonic/gin"
)

type CronTaskHandler struct {
	svcCtx     *svc.ServiceContext
	repository *repository.CronTaskRepository
}

func NewCronTaskHandler(svcCtx *svc.ServiceContext) *CronTaskHandler {
	return &CronTaskHandler{
		svcCtx:     svcCtx,
		repository: repository.NewCronTaskRepository(svcCtx),
	}
}

// ListTasks 分页获取任务列表
// @Summary 分页获取定时任务列表
// @Tags crontask
// @Accept json
// @Produce json
// @Param page query int false "页码 (默认 1)"
// @Param pageSize query int false "每页数量 (默认 10)"
// @Param status query string false "状态 (pending/running/completed/stopped/failed)"
// @Param type query string false "类型 (once/recurring)"
// @Param enabled query bool false "是否启用"
// @Success 200 {object} response.Response[response.PagedData[response.CronTaskResponse]]
// @Router /api/crontask [get]
func (h *CronTaskHandler) ListTasks(ctx *gin.Context) {
	var req request.ListTasksRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, err.Error()))
		return
	}

	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	tasks, total, err := h.repository.GetPagedTasks(ctx.Request.Context(), req.Page, req.PageSize, req.Status, req.Type, req.Enabled)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(response.PagedData[*response.CronTaskResponse]{
		Items: response.ToCronTaskListResponse(tasks),
		Page: response.PageInfo{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}))
}

// GetTask 获取任务详情
// @Summary 获取定时任务详情
// @Tags crontask
// @Accept json
// @Produce json
// @Param id path int true "任务 ID"
// @Success 200 {object} response.Response[response.CronTaskResponse]
// @Failure 400 {object} response.Response[any]
// @Failure 404 {object} response.Response[any]
// @Router /api/crontask/{id} [get]
func (h *CronTaskHandler) GetTask(ctx *gin.Context) {
	var req request.GetTaskRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, err.Error()))
		return
	}

	task, err := h.repository.GetByID(ctx.Request.Context(), req.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, response.Error[any](response.CodeResourceNotFound, "task not found"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(response.ToCronTaskResponse(task)))
}

// CreateTask 创建任务
// @Summary 创建定时任务
// @Tags crontask
// @Accept json
// @Produce json
// @Param task body request.CreateTaskRequest true "任务信息"
// @Success 200 {object} response.Response[response.CronTaskResponse]
// @Failure 400 {object} response.Response[any]
// @Router /api/crontask [post]
func (h *CronTaskHandler) CreateTask(ctx *gin.Context) {
	var req request.CreateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, err.Error()))
		return
	}

	task := &model.CronTask{
		Name:            req.Name,
		Spec:            req.Spec,
		Type:            req.Type,
		Status:          model.TaskStatusPending,
		ExecType:        req.ExecType,
		Raw:             req.Raw,
		Enabled:         req.Enabled,
		MaxRetries:      req.MaxRetries,
		TimeoutSeconds:  req.TimeoutSeconds,
		TotalExecutions: 0,
	}

	if req.ValidFrom != nil {
		task.ValidFrom = sql.NullTime{Time: *req.ValidFrom, Valid: true}
	}
	if req.ValidUntil != nil {
		task.ValidUntil = sql.NullTime{Time: *req.ValidUntil, Valid: true}
	}
	if req.NextExecutionAt != nil {
		task.NextExecutionAt = sql.NullTime{Time: *req.NextExecutionAt, Valid: true}
	}

	if err := h.repository.Create(ctx.Request.Context(), task); err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(response.ToCronTaskResponse(task)))
}

// UpdateTask 更新任务
// @Summary 更新定时任务
// @Tags crontask
// @Accept json
// @Produce json
// @Param id path int true "任务 ID"
// @Param task body request.UpdateTaskBody true "任务信息"
// @Success 200 {object} response.Response[response.CronTaskResponse]
// @Failure 400 {object} response.Response[any]
// @Failure 404 {object} response.Response[any]
// @Router /api/crontask/{id} [put]
func (h *CronTaskHandler) UpdateTask(ctx *gin.Context) {
	var uriReq request.UpdateTaskRequest
	if err := ctx.ShouldBindUri(&uriReq); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, err.Error()))
		return
	}

	var bodyReq request.UpdateTaskBody
	if err := ctx.ShouldBindJSON(&bodyReq); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, err.Error()))
		return
	}

	task, err := h.repository.GetByID(ctx.Request.Context(), uriReq.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, response.Error[any](response.CodeResourceNotFound, "task not found"))
		return
	}

	// Update fields
	if bodyReq.Name != "" {
		task.Name = bodyReq.Name
	}
	if bodyReq.Spec != "" {
		task.Spec = bodyReq.Spec
	}
	if bodyReq.Type != "" {
		task.Type = bodyReq.Type
	}
	if bodyReq.ExecType != "" {
		task.ExecType = bodyReq.ExecType
	}
	if bodyReq.Raw != "" {
		task.Raw = bodyReq.Raw
	}
	if bodyReq.ValidFrom != nil {
		task.ValidFrom = sql.NullTime{Time: *bodyReq.ValidFrom, Valid: true}
	} else {
		task.ValidFrom = sql.NullTime{}
	}
	if bodyReq.ValidUntil != nil {
		task.ValidUntil = sql.NullTime{Time: *bodyReq.ValidUntil, Valid: true}
	} else {
		task.ValidUntil = sql.NullTime{}
	}
	if bodyReq.Enabled != nil {
		task.Enabled = *bodyReq.Enabled
	}
	if bodyReq.MaxRetries > 0 {
		task.MaxRetries = bodyReq.MaxRetries
	}
	if bodyReq.TimeoutSeconds > 0 {
		task.TimeoutSeconds = bodyReq.TimeoutSeconds
	}
	if bodyReq.NextExecutionAt != nil {
		task.NextExecutionAt = sql.NullTime{Time: *bodyReq.NextExecutionAt, Valid: true}
	}

	if err := h.repository.Update(ctx.Request.Context(), task); err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(response.ToCronTaskResponse(task)))
}

// DeleteTask 删除任务
// @Summary 删除定时任务
// @Tags crontask
// @Accept json
// @Produce json
// @Param id path int true "任务 ID"
// @Success 200 {object} response.Response[string]
// @Failure 400 {object} response.Response[any]
// @Failure 404 {object} response.Response[any]
// @Router /api/crontask/{id} [delete]
func (h *CronTaskHandler) DeleteTask(ctx *gin.Context) {
	var req request.DeleteTaskRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, err.Error()))
		return
	}

	_, err := h.repository.GetByID(ctx.Request.Context(), req.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, response.Error[any](response.CodeResourceNotFound, "task not found"))
		return
	}

	if err := h.repository.Delete(ctx.Request.Context(), req.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success("deleted"))
}

// EnableTask 启用任务
// @Summary 启用定时任务
// @Tags crontask
// @Accept json
// @Produce json
// @Param id path int true "任务 ID"
// @Success 200 {object} response.Response[string]
// @Failure 400 {object} response.Response[any]
// @Failure 404 {object} response.Response[any]
// @Router /api/crontask/{id}/enable [post]
func (h *CronTaskHandler) EnableTask(ctx *gin.Context) {
	var req request.TaskActionRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, err.Error()))
		return
	}

	_, err := h.repository.GetByID(ctx.Request.Context(), req.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, response.Error[any](response.CodeResourceNotFound, "task not found"))
		return
	}

	if err := h.repository.Enable(ctx.Request.Context(), req.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success("enabled"))
}

// DisableTask 禁用任务
// @Summary 禁用定时任务
// @Tags crontask
// @Accept json
// @Produce json
// @Param id path int true "任务 ID"
// @Success 200 {object} response.Response[string]
// @Failure 400 {object} response.Response[any]
// @Failure 404 {object} response.Response[any]
// @Router /api/crontask/{id}/disable [post]
func (h *CronTaskHandler) DisableTask(ctx *gin.Context) {
	var req request.TaskActionRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, err.Error()))
		return
	}

	_, err := h.repository.GetByID(ctx.Request.Context(), req.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, response.Error[any](response.CodeResourceNotFound, "task not found"))
		return
	}

	if err := h.repository.Disable(ctx.Request.Context(), req.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success("disabled"))
}

// StartTask 启动任务（一次性任务或设置下次执行时间）
// @Summary 启动定时任务
// @Tags crontask
// @Accept json
// @Produce json
// @Param id path int true "任务 ID"
// @Param body body request.StartTaskRequest true "启动参数"
// @Success 200 {object} response.Response[string]
// @Failure 400 {object} response.Response[any]
// @Failure 404 {object} response.Response[any]
// @Router /api/crontask/{id}/start [post]
func (h *CronTaskHandler) StartTask(ctx *gin.Context) {
	var req request.StartTaskRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, err.Error()))
		return
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, err.Error()))
		return
	}

	task, err := h.repository.GetByID(ctx.Request.Context(), req.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, response.Error[any](response.CodeResourceNotFound, "task not found"))
		return
	}

	// Parse next execution time
	nextExecTime, err := time.Parse("2006-01-02 15:04:05", req.NextExecutionTime)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, "invalid time format, use YYYY-MM-DD HH:MM:SS"))
		return
	}

	// Update next execution time and status
	if err := h.repository.UpdateNextExecution(ctx.Request.Context(), req.ID, nextExecTime); err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
		return
	}

	task.Status = model.TaskStatusPending
	task.Enabled = true
	if err := h.repository.Update(ctx.Request.Context(), task); err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success("started"))
}

// StopTask 停止任务
// @Summary 停止定时任务
// @Tags crontask
// @Accept json
// @Produce json
// @Param id path int true "任务 ID"
// @Success 200 {object} response.Response[string]
// @Failure 400 {object} response.Response[any]
// @Failure 404 {object} response.Response[any]
// @Router /api/crontask/{id}/stop [post]
func (h *CronTaskHandler) StopTask(ctx *gin.Context) {
	var req request.TaskActionRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, err.Error()))
		return
	}

	_, err := h.repository.GetByID(ctx.Request.Context(), req.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, response.Error[any](response.CodeResourceNotFound, "task not found"))
		return
	}

	if err := h.repository.MarkAsStopped(ctx.Request.Context(), req.ID); err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success("stopped"))
}
