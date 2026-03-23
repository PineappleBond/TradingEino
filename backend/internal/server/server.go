package server

import (
	"github.com/PineappleBond/TradingEino/backend/internal/api"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/gin-gonic/gin"
)

type Server struct {
	ServiceContext *svc.ServiceContext
	Engine         *gin.Engine
}

func NewServer(serviceContext *svc.ServiceContext) *Server {
	engine := gin.Default()
	// CORS middleware is applied in routes if needed
	api.Routes(engine, serviceContext)
	return &Server{
		ServiceContext: serviceContext,
		Engine:         engine,
	}
}

func (s *Server) Start() error {
	return s.Engine.Run(":8080")
}
