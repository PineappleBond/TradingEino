package middleware

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/gin-gonic/gin"
)

// responseWriter wraps gin.ResponseWriter to capture response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// Logger returns a middleware that logs HTTP requests using the service logger
// Similar to gormLogger pattern, it captures request/response details
func Logger(serviceContext *svc.ServiceContext) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Read request body if present
		var requestBody string
		if c.Request.Body != nil && c.Request.ContentLength > 0 && c.Request.ContentLength < 1024*1024 {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err == nil {
				requestBody = string(bodyBytes)
				// Restore body for downstream handlers
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}

		// Capture response body
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer

		c.Next()

		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// Build log message with request/response details
		var logParts []string
		logParts = append(logParts, fmt.Sprintf("[HTTP] %d %s %s", statusCode, c.Request.Method, path))
		logParts = append(logParts, fmt.Sprintf("duration:%v", latency))

		if query != "" {
			logParts = append(logParts, fmt.Sprintf("query:%s", query))
		}

		// Log request headers
		if len(c.Request.Header) > 0 {
			headers := make([]string, 0, len(c.Request.Header))
			for k, v := range c.Request.Header {
				if !strings.EqualFold(k, "Authorization") && !strings.EqualFold(k, "Cookie") {
					headers = append(headers, fmt.Sprintf("%s:%v", k, v))
				}
			}
			if len(headers) > 0 {
				logParts = append(logParts, fmt.Sprintf("req_headers:[%s]", strings.Join(headers, ", ")))
			}
		}

		// Log request body if present
		if requestBody != "" {
			logParts = append(logParts, fmt.Sprintf("req_body:%s", truncateString(requestBody, 1024)))
		}

		// Log response body if captured
		responseBody := writer.body.String()
		if responseBody != "" {
			logParts = append(logParts, fmt.Sprintf("res_body:%s", truncateString(responseBody, 1024)))
		}

		msg := strings.Join(logParts, " ")

		// Choose log function based on status code
		switch {
		case statusCode >= 500:
			serviceContext.Logger4Gin.Error(c.Request.Context(), msg, nil)
		case statusCode >= 400:
			serviceContext.Logger4Gin.Warn(c.Request.Context(), msg)
		default:
			serviceContext.Logger4Gin.Info(c.Request.Context(), msg)
		}
	}
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
