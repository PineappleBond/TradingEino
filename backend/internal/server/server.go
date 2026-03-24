package server

import (
	"context"
	"net/http"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/api"
	"github.com/PineappleBond/TradingEino/backend/internal/api/middleware"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/gin-gonic/gin"
)

type Server struct {
	ServiceContext *svc.ServiceContext
	Engine         *gin.Engine
	httpServer     *http.Server
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

	srv := &Server{
		ServiceContext: serviceContext,
		Engine:         engine,
	}

	// Create http.Server for graceful shutdown support
	srv.httpServer = &http.Server{
		Addr:    serviceContext.Config.Server.ListenOn,
		Handler: engine,
	}

	return srv
}

func (s *Server) Start() error {
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server with timeout
func (s *Server) Shutdown(ctx context.Context) error {
	// Create a timeout context for shutdown
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	return s.httpServer.Shutdown(shutdownCtx)
}
