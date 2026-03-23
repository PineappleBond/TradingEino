package handler

import (
	"net/http"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/api/response"
	"github.com/PineappleBond/TradingEino/backend/internal/model"
	"github.com/PineappleBond/TradingEino/backend/internal/repository"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/gin-gonic/gin"
)

type CronExecutionHandler struct {
	svcCtx     *svc.ServiceContext
	repository *repository.CronExecutionRepository
}

func NewCronExecutionHandler(svcCtx *svc.ServiceContext) *CronExecutionHandler {
	return &CronExecutionHandler{
		svcCtx:     svcCtx,
		repository: repository.NewCronExecutionRepository(svcCtx),
	}
}

// ListExecutionsRequest 获取执行记录列表请求
type ListExecutionsRequest struct {
	Page      int                      `form:"page"`
	PageSize  int                      `form:"pageSize"`
	TaskID    *uint                    `form:"task_id"`
	Status    *model.ExecutionStatus   `form:"status"`
	StartTime *time.Time               `form:"start_time"`
	EndTime   *time.Time               `form:"end_time"`
}

// ListExecutions 分页获取执行记录列表
// @Summary 分页获取定时任务执行记录列表
// @Tags cronexecution
// @Accept json
// @Produce json
// @Param page query int false "页码 (默认 1)"
// @Param pageSize query int false "每页数量 (默认 10)"
// @Param task_id query uint false "任务 ID"
// @Param status query string false "状态 (pending/running/success/failed/retried/cancelled)"
// @Param start_time query string false "开始时间 (2006-01-02T15:04:05Z07:00)"
// @Param end_time query string false "结束时间 (2006-01-02T15:04:05Z07:00)"
// @Success 200 {object} response.Response[response.PagedData[model.CronExecution]]
// @Router /api/cronexecution [get]
func (h *CronExecutionHandler) ListExecutions(ctx *gin.Context) {
	var req ListExecutionsRequest
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

	executions, total, err := h.repository.GetPagedExecutions(
		ctx.Request.Context(),
		req.Page,
		req.PageSize,
		req.TaskID,
		req.Status,
		req.StartTime,
		req.EndTime,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(response.PagedData[*model.CronExecution]{
		Items: executions,
		Page: response.PageInfo{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}))
}

// GetExecutionRequest 获取执行记录详情请求
type GetExecutionRequest struct {
	ID uint `uri:"id" binding:"required,min=1"`
}

// GetExecution 获取执行记录详情
// @Summary 获取定时任务执行记录详情
// @Tags cronexecution
// @Accept json
// @Produce json
// @Param id path int true "执行记录 ID"
// @Success 200 {object} response.Response[model.CronExecution]
// @Failure 400 {object} response.Response[any]
// @Failure 404 {object} response.Response[any]
// @Router /api/cronexecution/{id} [get]
func (h *CronExecutionHandler) GetExecution(ctx *gin.Context) {
	var req GetExecutionRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, err.Error()))
		return
	}

	execution, err := h.repository.GetByExecutionID(ctx.Request.Context(), req.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, response.Error[any](response.CodeResourceNotFound, "execution not found"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(execution))
}

// GetByTaskIDRequest 获取任务执行记录列表请求
type GetByTaskIDRequest struct {
	TaskID uint `uri:"task_id" binding:"required,min=1"`
	Page   int  `form:"page"`
	PageSize int `form:"pageSize"`
}

// GetByTaskID 分页获取指定任务的所有执行记录
// @Summary 分页获取指定任务的所有执行记录
// @Tags cronexecution
// @Accept json
// @Produce json
// @Param task_id path int true "任务 ID"
// @Param page query int false "页码 (默认 1)"
// @Param pageSize query int false "每页数量 (默认 10)"
// @Success 200 {object} response.Response[response.PagedData[model.CronExecution]]
// @Failure 400 {object} response.Response[any]
// @Router /api/cronexecution/task/{task_id} [get]
func (h *CronExecutionHandler) GetByTaskID(ctx *gin.Context) {
	var req GetByTaskIDRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, err.Error()))
		return
	}
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

	executions, total, err := h.repository.GetPagedExecutions(
		ctx.Request.Context(),
		req.Page,
		req.PageSize,
		&req.TaskID,
		nil,
		nil,
		nil,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(response.PagedData[*model.CronExecution]{
		Items: executions,
		Page: response.PageInfo{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}))
}
