package response

import "time"

// LogFileInfo represents information about a log file
type LogFileInfo struct {
	Filename     string    `json:"filename"`
	Size         int64     `json:"size"`
	ModTime      time.Time `json:"mod_time"`
	LineCount    int       `json:"line_count"`
	FirstLogTime *time.Time `json:"first_log_time,omitempty"`
	LastLogTime  *time.Time `json:"last_log_time,omitempty"`
}

// LogEntry represents a parsed JSON log entry
type LogEntry struct {
	Time    time.Time `json:"time"`
	Level   string    `json:"level"`
	Message string    `json:"msg"`
}

// LogStatsResponse 日志统计响应
type LogStatsResponse struct {
	TotalEntries int64            `json:"total_entries"`
	LevelCounts  map[string]int64 `json:"level_counts"`
	HourlyCounts []HourlyCount    `json:"hourly_counts"`
}

// HourlyCount represents log count per hour
type HourlyCount struct {
	Hour  string `json:"hour"`
	Count int64  `json:"count"`
}
