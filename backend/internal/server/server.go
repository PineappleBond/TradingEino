package server

import (
	"github.com/PineappleBond/TradingEino/backend/internal/api"
	"github.com/PineappleBond/TradingEino/backend/internal/api/middleware"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/gin-gonic/gin"
)

type Server struct {
	ServiceContext *svc.ServiceContext
	Engine         *gin.Engine
}

func NewServer(serviceContext *svc.ServiceContext) *Server {
	if serviceContext.Config.Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else if serviceContext.Config.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else if serviceContext.Config.Server.Mode == "test" {
		gin.SetMode(gin.TestMode)
	}
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(middleware.Logger(serviceContext))
	engine.Use(middleware.Cors(serviceContext))
	// CORS middleware is applied in routes if needed
	api.Routes(engine, serviceContext)
	return &Server{
		ServiceContext: serviceContext,
		Engine:         engine,
	}
}

func (s *Server) Start() error {
	return s.Engine.Run(s.ServiceContext.Config.Server.ListenOn)
}
