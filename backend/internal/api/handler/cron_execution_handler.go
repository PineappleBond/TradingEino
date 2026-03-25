package handler

import (
	"net/http"

	"github.com/PineappleBond/TradingEino/backend/internal/api/request"
	"github.com/PineappleBond/TradingEino/backend/internal/api/response"
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
// @Success 200 {object} response.Response[response.PagedData[response.CronExecutionResponse]]
// @Router /api/cronexecution [get]
func (h *CronExecutionHandler) ListExecutions(ctx *gin.Context) {
	var req request.ListExecutionsRequest
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

	ctx.JSON(http.StatusOK, response.Success(response.PagedData[*response.CronExecutionResponse]{
		Items: response.ToCronExecutionListResponse(executions),
		Page: response.PageInfo{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}))
}

// GetExecution 获取执行记录详情
// @Summary 获取定时任务执行记录详情
// @Tags cronexecution
// @Accept json
// @Produce json
// @Param id path int true "执行记录 ID"
// @Success 200 {object} response.Response[response.CronExecutionResponse]
// @Failure 400 {object} response.Response[any]
// @Failure 404 {object} response.Response[any]
// @Router /api/cronexecution/{id} [get]
func (h *CronExecutionHandler) GetExecution(ctx *gin.Context) {
	var req request.GetExecutionRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, err.Error()))
		return
	}

	execution, err := h.repository.GetByExecutionID(ctx.Request.Context(), req.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, response.Error[any](response.CodeResourceNotFound, "execution not found"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(response.ToCronExecutionResponse(execution)))
}

// GetByTaskID 分页获取指定任务的所有执行记录
// @Summary 分页获取指定任务的所有执行记录
// @Tags cronexecution
// @Accept json
// @Produce json
// @Param task_id path int true "任务 ID"
// @Param page query int false "页码 (默认 1)"
// @Param pageSize query int false "每页数量 (默认 10)"
// @Success 200 {object} response.Response[response.PagedData[response.CronExecutionResponse]]
// @Failure 400 {object} response.Response[any]
// @Router /api/cronexecution/task/{task_id} [get]
func (h *CronExecutionHandler) GetByTaskID(ctx *gin.Context) {
	var req request.GetByTaskIDRequest
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

	ctx.JSON(http.StatusOK, response.Success(response.PagedData[*response.CronExecutionResponse]{
		Items: response.ToCronExecutionListResponse(executions),
		Page: response.PageInfo{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}))
}
