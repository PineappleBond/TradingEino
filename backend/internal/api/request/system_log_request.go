package request

import "time"

// ListLogFilesRequest 获取日志文件列表请求
type ListLogFilesRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"pageSize"`
}

// GetLogContentRequest 获取日志内容请求
type GetLogContentRequest struct {
	Filename  string     `uri:"filename" binding:"required"`
	Page      int        `form:"page"`
	PageSize  int        `form:"pageSize"`
	Level     *string    `form:"level"`
	StartTime *time.Time `form:"start_time"`
	EndTime   *time.Time `form:"end_time"`
}

// SearchLogsRequest 搜索日志请求
type SearchLogsRequest struct {
	Keyword   string     `form:"keyword" binding:"required"`
	Filename  *string    `form:"filename"`
	Level     *string    `form:"level"`
	StartTime *time.Time `form:"start_time"`
	EndTime   *time.Time `form:"end_time"`
	Page      int        `form:"page"`
	PageSize  int        `form:"pageSize"`
}

// GetLogStatsRequest 获取日志统计请求
type GetLogStatsRequest struct {
	Filename  *string    `form:"filename"`
	StartTime *time.Time `form:"start_time"`
	EndTime   *time.Time `form:"end_time"`
}
