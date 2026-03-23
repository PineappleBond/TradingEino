package handler

import (
	"net/http"

	"github.com/PineappleBond/TradingEino/backend/internal/api/response"
	"github.com/PineappleBond/TradingEino/backend/internal/model"
	"github.com/PineappleBond/TradingEino/backend/internal/repository"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/gin-gonic/gin"
)

type CronExecutionLogHandler struct {
	svcCtx     *svc.ServiceContext
	repository *repository.CronExecutionLogRepository
}

func NewCronExecutionLogHandler(svcCtx *svc.ServiceContext) *CronExecutionLogHandler {
	return &CronExecutionLogHandler{
		svcCtx:     svcCtx,
		repository: repository.NewCronExecutionLogRepository(svcCtx),
	}
}

// ListLogsRequest 获取日志记录列表请求
type ListLogsRequest struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"pageSize"`
	ExecutionID *uint `form:"execution_id"`
	Level      *string `form:"level"`
	From       *string `form:"from"`
}

// ListLogs 分页获取日志记录列表
// @Summary 分页获取定时任务执行日志列表
// @Tags cronexecutionlog
// @Accept json
// @Produce json
// @Param page query int false "页码 (默认 1)"
// @Param pageSize query int false "每页数量 (默认 10)"
// @Param execution_id query uint false "执行记录 ID"
// @Param level query string false "日志级别 (info/warn/error/debug)"
// @Param from query string false "日志来源"
// @Success 200 {object} response.Response[response.PagedData[model.CronExecutionLog]]
// @Router /api/cronexecutionlog [get]
func (h *CronExecutionLogHandler) ListLogs(ctx *gin.Context) {
	var req ListLogsRequest
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

	logs, total, err := h.repository.GetPagedLogs(
		ctx.Request.Context(),
		req.Page,
		req.PageSize,
		req.ExecutionID,
		req.Level,
		req.From,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(response.PagedData[*model.CronExecutionLog]{
		Items: logs,
		Page: response.PageInfo{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}))
}

// GetLogRequest 获取日志记录详情请求
type GetLogRequest struct {
	ID uint `uri:"id" binding:"required,min=1"`
}

// GetLog 获取日志记录详情
// @Summary 获取定时任务执行日志详情
// @Tags cronexecutionlog
// @Accept json
// @Produce json
// @Param id path int true "日志记录 ID"
// @Success 200 {object} response.Response[model.CronExecutionLog]
// @Failure 400 {object} response.Response[any]
// @Failure 404 {object} response.Response[any]
// @Router /api/cronexecutionlog/{id} [get]
func (h *CronExecutionLogHandler) GetLog(ctx *gin.Context) {
	var req GetLogRequest
	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, err.Error()))
		return
	}

	log, err := h.repository.GetByID(ctx.Request.Context(), req.ID)
	if err != nil {
		ctx.JSON(http.StatusNotFound, response.Error[any](response.CodeResourceNotFound, "log not found"))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(log))
}

// GetByExecutionIDRequest 根据执行 ID 获取日志列表请求
type GetByExecutionIDRequest struct {
	ExecutionID uint `uri:"execution_id" binding:"required,min=1"`
	Page        int  `form:"page"`
	PageSize    int  `form:"pageSize"`
}

// GetByExecutionID 分页获取指定执行记录的所有日志
// @Summary 分页获取指定执行记录的所有日志
// @Tags cronexecutionlog
// @Accept json
// @Produce json
// @Param execution_id path int true "执行记录 ID"
// @Param page query int false "页码 (默认 1)"
// @Param pageSize query int false "每页数量 (默认 10)"
// @Success 200 {object} response.Response[response.PagedData[model.CronExecutionLog]]
// @Failure 400 {object} response.Response[any]
// @Router /api/cronexecutionlog/execution/{execution_id} [get]
func (h *CronExecutionLogHandler) GetByExecutionID(ctx *gin.Context) {
	var req GetByExecutionIDRequest
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

	logs, total, err := h.repository.GetPagedLogs(
		ctx.Request.Context(),
		req.Page,
		req.PageSize,
		&req.ExecutionID,
		nil,
		nil,
	)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(response.PagedData[*model.CronExecutionLog]{
		Items: logs,
		Page: response.PageInfo{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}))
}
