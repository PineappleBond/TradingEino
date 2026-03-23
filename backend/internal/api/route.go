package api

import (
	"io/fs"
	"log"
	"net/http"
	"strings"

	"github.com/PineappleBond/TradingEino/backend/internal/api/handler"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/PineappleBond/TradingEino/backend/web"
	"github.com/gin-gonic/gin"
)

func Routes(engine *gin.Engine, ctx *svc.ServiceContext) {
	setUpWeb(engine)
	api := engine.Group("/api")
	{
		// health check
		checkHandler := handler.NewHealthCheckHandler(ctx)
		api.GET("/health", checkHandler.HealthCheck)
	}
	{
		// crontask
		cronTaskHandler := handler.NewCronTaskHandler(ctx)
		api.GET("/crontask", cronTaskHandler.ListTasks)
		api.GET("/crontask/:id", cronTaskHandler.GetTask)
		api.POST("/crontask", cronTaskHandler.CreateTask)
		api.PUT("/crontask/:id", cronTaskHandler.UpdateTask)
		api.DELETE("/crontask/:id", cronTaskHandler.DeleteTask)
		api.POST("/crontask/:id/enable", cronTaskHandler.EnableTask)
		api.POST("/crontask/:id/disable", cronTaskHandler.DisableTask)
		api.POST("/crontask/:id/start", cronTaskHandler.StartTask)
		api.POST("/crontask/:id/stop", cronTaskHandler.StopTask)
	}
	{
		// cronexecution
		cronExecutionHandler := handler.NewCronExecutionHandler(ctx)
		api.GET("/cronexecution", cronExecutionHandler.ListExecutions)
		api.GET("/cronexecution/:id", cronExecutionHandler.GetExecution)
		api.GET("/cronexecution/task/:task_id", cronExecutionHandler.GetByTaskID)
	}
	{
		// cronexecutionlog
		cronExecutionLogHandler := handler.NewCronExecutionLogHandler(ctx)
		api.GET("/cronexecutionlog", cronExecutionLogHandler.ListLogs)
		api.GET("/cronexecutionlog/:id", cronExecutionLogHandler.GetLog)
		api.GET("/cronexecutionlog/execution/:execution_id", cronExecutionLogHandler.GetByExecutionID)
	}
	{
		// systemlog 读取日志输出目录，解析 jsonl，分页返回
		systemLogHandler := handler.NewSystemLogHandler(ctx)
		api.GET("/systemlog/files", systemLogHandler.ListLogFiles)
		api.GET("/systemlog/files/:filename", systemLogHandler.GetLogContent)
		api.GET("/systemlog/search", systemLogHandler.SearchLogs)
		api.GET("/systemlog/stats", systemLogHandler.GetLogStats)
	}
}

func setUpWeb(r *gin.Engine) {
	// 静态文件服务 - 嵌入前端资源
	staticFS, err := fs.Sub(web.DistFS, "dist")
	if err != nil {
		log.Fatalf("Failed to create static filesystem: %v", err)
	}

	// 调试：列出嵌入的文件
	log.Println("Embedded files:")
	fs.WalkDir(staticFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err == nil {
			log.Printf("  %s (dir: %v)", path, d.IsDir())
		}
		return nil
	})

	// 前端路由处理
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 尝试提供静态文件
		if isStaticFile(path) {
			// 移除开头的 "/"
			filePath := path[1:]
			if data, err := fs.ReadFile(staticFS, filePath); err == nil {
				// 设置正确的 Content-Type
				if strings.HasSuffix(path, ".js.map") || strings.HasSuffix(path, ".css.map") || strings.HasSuffix(path, ".map") {
					c.Header("Content-Type", "application/json")
				} else if strings.HasSuffix(path, ".js") {
					c.Header("Content-Type", "application/javascript")
				} else if strings.HasSuffix(path, ".css") {
					c.Header("Content-Type", "text/css")
				}
				c.Data(http.StatusOK, "", data)
				return
			}
		}

		// 对于所有其他路由，返回 index.html（SPA 路由）
		c.Header("Content-Type", "text/html")
		data, err := fs.ReadFile(staticFS, "index.html")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load frontend"})
			return
		}
		c.Data(http.StatusOK, "text/html", data)
	})
}

// isStaticFile 检查路径是否是静态文件
func isStaticFile(path string) bool {
	staticExtensions := []string{".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".eot", ".map", ".js.map", ".css.map"}
	for _, ext := range staticExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	return false
}
