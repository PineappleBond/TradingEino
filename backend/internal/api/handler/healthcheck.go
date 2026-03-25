package handler

import (
	"net/http"

	"github.com/PineappleBond/TradingEino/backend/internal/api/response"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/gin-gonic/gin"
)

type HealthCheckHandler struct {
	svcCtx *svc.ServiceContext
}

func NewHealthCheckHandler(svcCtx *svc.ServiceContext) *HealthCheckHandler {
	return &HealthCheckHandler{svcCtx: svcCtx}
}

func (h *HealthCheckHandler) HealthCheck(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, response.Success("ok"))
}
