package handler

import (
	"bufio"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/PineappleBond/TradingEino/backend/internal/api/request"
	"github.com/PineappleBond/TradingEino/backend/internal/api/response"
	"github.com/PineappleBond/TradingEino/backend/internal/svc"
	"github.com/gin-gonic/gin"
)

// SystemLogHandler handles system log file operations
type SystemLogHandler struct {
	svcCtx  *svc.ServiceContext
	logDir  string
}

// NewSystemLogHandler creates a new SystemLogHandler
func NewSystemLogHandler(svcCtx *svc.ServiceContext) *SystemLogHandler {
	// Get log directory from config
	logDir := svcCtx.Config.Logger.FilePath
	if logDir == "" {
		logDir = "./backend/logs"
	} else {
		logDir = filepath.Dir(logDir)
	}
	return &SystemLogHandler{
		svcCtx: svcCtx,
		logDir: logDir,
	}
}

// ListLogFiles 分页获取日志文件列表
// @Summary 分页获取日志文件列表
// @Tags systemlog
// @Accept json
// @Produce json
// @Param page query int false "页码 (默认 1)"
// @Param pageSize query int false "每页数量 (默认 10)"
// @Success 200 {object} response.Response[response.PagedData[response.LogFileInfo]]
// @Router /api/systemlog/files [get]
func (h *SystemLogHandler) ListLogFiles(ctx *gin.Context) {
	var req request.ListLogFilesRequest
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

	// Find all .jsonl files in log directory
	var files []response.LogFileInfo
	err := filepath.WalkDir(h.logDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".jsonl") {
			info, err := d.Info()
			if err != nil {
				return nil
			}
			// Get line count and time range
			lineCount, firstTime, lastTime := analyzeLogFile(path)
			files = append(files, response.LogFileInfo{
				Filename:     d.Name(),
				Size:         info.Size(),
				ModTime:      info.ModTime(),
				LineCount:    lineCount,
				FirstLogTime: firstTime,
				LastLogTime:  lastTime,
			})
		}
		return nil
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
		return
	}

	// Sort by modification time (newest first)
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime.After(files[j].ModTime)
	})

	// Apply pagination
	total := int64(len(files))
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if start >= len(files) {
		files = []response.LogFileInfo{}
	} else {
		if end > len(files) {
			end = len(files)
		}
		files = files[start:end]
	}

	ctx.JSON(http.StatusOK, response.Success(response.PagedData[response.LogFileInfo]{
		Items: files,
		Page: response.PageInfo{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}))
}

// GetLogContent 分页获取日志文件内容
// @Summary 分页获取日志文件内容
// @Tags systemlog
// @Accept json
// @Produce json
// @Param filename path string true "日志文件名"
// @Param page query int false "页码 (默认 1)"
// @Param pageSize query int false "每页数量 (默认 10)"
// @Param level query string false "日志级别 (INFO/WARN/ERROR/DEBUG)"
// @Param start_time query string false "开始时间 (2006-01-02T15:04:05Z07:00)"
// @Param end_time query string false "结束时间 (2006-01-02T15:04:05Z07:00)"
// @Success 200 {object} response.Response[response.PagedData[response.LogEntry]]
// @Failure 400 {object} response.Response[any]
// @Failure 404 {object} response.Response[any]
// @Router /api/systemlog/files/{filename} [get]
func (h *SystemLogHandler) GetLogContent(ctx *gin.Context) {
	var req request.GetLogContentRequest
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

	filePath := filepath.Join(h.logDir, req.Filename)

	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		ctx.JSON(http.StatusNotFound, response.Error[any](response.CodeResourceNotFound, "log file not found"))
		return
	}
	if info.IsDir() {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, "not a file"))
		return
	}

	// Read and parse log file
	entries, total, err := h.readLogfileWithPagination(filePath, req.Page, req.PageSize, req.Level, req.StartTime, req.EndTime)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, response.Success(response.PagedData[response.LogEntry]{
		Items: entries,
		Page: response.PageInfo{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}))
}

// readLogfileWithPagination reads a JSONL log file with pagination and filtering
func (h *SystemLogHandler) readLogfileWithPagination(filePath string, page, pageSize int, level *string, startTime, endTime *time.Time) ([]response.LogEntry, int64, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	var allEntries []response.LogEntry
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		var entry response.LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue // Skip invalid JSON lines
		}

		// Apply filters
		if level != nil && !strings.EqualFold(entry.Level, *level) {
			continue
		}
		if startTime != nil && entry.Time.Before(*startTime) {
			continue
		}
		if endTime != nil && entry.Time.After(*endTime) {
			continue
		}

		allEntries = append(allEntries, entry)
	}

	if err := scanner.Err(); err != nil {
		return nil, 0, err
	}

	// Sort by time (newest first)
	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].Time.After(allEntries[j].Time)
	})

	// Apply pagination
	total := int64(len(allEntries))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(allEntries) {
		return []response.LogEntry{}, total, nil
	}
	if end > len(allEntries) {
		end = len(allEntries)
	}

	return allEntries[start:end], total, nil
}

// analyzeLogFile analyzes a log file and returns line count and time range
func analyzeLogFile(filePath string) (lineCount int, firstTime, lastTime *time.Time) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, nil, nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var times []time.Time

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		lineCount++

		var entry response.LogEntry
		if err := json.Unmarshal([]byte(line), &entry); err == nil {
			times = append(times, entry.Time)
		}
	}

	if len(times) > 0 {
		// Find first and last times
		first := times[0]
		last := times[0]
		for _, t := range times[1:] {
			if t.Before(first) {
				first = t
			}
			if t.After(last) {
				last = t
			}
		}
		firstTime = &first
		lastTime = &last
	}

	return lineCount, firstTime, lastTime
}

// SearchLogs 搜索日志内容
// @Summary 搜索日志内容
// @Tags systemlog
// @Accept json
// @Produce json
// @Param keyword query string true "搜索关键词"
// @Param filename query string false "日志文件名"
// @Param level query string false "日志级别"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Param page query int false "页码"
// @Param pageSize query int false "每页数量"
// @Success 200 {object} response.Response[response.PagedData[response.LogEntry]]
// @Failure 400 {object} response.Response[any]
// @Router /api/systemlog/search [get]
func (h *SystemLogHandler) SearchLogs(ctx *gin.Context) {
	var req request.SearchLogsRequest
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

	var allEntries []response.LogEntry

	// Determine which files to search
	var filesToSearch []string
	if req.Filename != nil {
		filesToSearch = []string{filepath.Join(h.logDir, *req.Filename)}
	} else {
		// Search all .jsonl files
		err := filepath.WalkDir(h.logDir, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".jsonl") {
				filesToSearch = append(filesToSearch, path)
			}
			return nil
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
			return
		}
	}

	// Search in each file
	for _, filePath := range filesToSearch {
		entries, _, err := h.readLogfileWithPagination(filePath, 1, 10000, req.Level, req.StartTime, req.EndTime)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if strings.Contains(entry.Message, req.Keyword) ||
				strings.Contains(entry.Level, req.Keyword) {
				allEntries = append(allEntries, entry)
			}
		}
	}

	// Sort by time (newest first)
	sort.Slice(allEntries, func(i, j int) bool {
		return allEntries[i].Time.After(allEntries[j].Time)
	})

	// Apply pagination
	total := int64(len(allEntries))
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize

	if start >= len(allEntries) {
		allEntries = []response.LogEntry{}
	} else {
		if end > len(allEntries) {
			end = len(allEntries)
		}
		allEntries = allEntries[start:end]
	}

	ctx.JSON(http.StatusOK, response.Success(response.PagedData[response.LogEntry]{
		Items: allEntries,
		Page: response.PageInfo{
			Page:     req.Page,
			PageSize: req.PageSize,
			Total:    total,
		},
	}))
}

// GetLogStats 获取日志统计信息
// @Summary 获取日志统计信息
// @Tags systemlog
// @Accept json
// @Produce json
// @Param filename query string false "日志文件名"
// @Param start_time query string false "开始时间"
// @Param end_time query string false "结束时间"
// @Success 200 {object} response.Response[response.LogStatsResponse]
// @Router /api/systemlog/stats [get]
func (h *SystemLogHandler) GetLogStats(ctx *gin.Context) {
	var req request.GetLogStatsRequest
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, response.Error[any](response.CodeParameterFormatError, err.Error()))
		return
	}

	levelCounts := make(map[string]int64)
	hourlyCounts := make(map[string]int64)
	var totalEntries int64

	// Determine which files to analyze
	var filesToSearch []string
	if req.Filename != nil {
		filesToSearch = []string{filepath.Join(h.logDir, *req.Filename)}
	} else {
		err := filepath.WalkDir(h.logDir, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".jsonl") {
				filesToSearch = append(filesToSearch, path)
			}
			return nil
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, response.Error[any](response.CodeDatabaseError, err.Error()))
			return
		}
	}

	// Analyze each file
	for _, filePath := range filesToSearch {
		file, err := os.Open(filePath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.TrimSpace(line) == "" {
				continue
			}

			var entry response.LogEntry
			if err := json.Unmarshal([]byte(line), &entry); err != nil {
				continue
			}

			// Apply time filters
			if req.StartTime != nil && entry.Time.Before(*req.StartTime) {
				continue
			}
			if req.EndTime != nil && entry.Time.After(*req.EndTime) {
				continue
			}

			totalEntries++
			levelCounts[entry.Level]++

			// Group by hour
			hourKey := entry.Time.Format("2006-01-02 15:00")
			hourlyCounts[hourKey]++
		}
		file.Close()
	}

	// Convert hourly counts to sorted slice
	hourlySlice := []response.HourlyCount{}
	for hour, count := range hourlyCounts {
		hourlySlice = append(hourlySlice, response.HourlyCount{
			Hour:  hour,
			Count: count,
		})
	}
	sort.Slice(hourlySlice, func(i, j int) bool {
		return hourlySlice[i].Hour < hourlySlice[j].Hour
	})

	// Limit to last 24 hours
	if len(hourlySlice) > 24 {
		hourlySlice = hourlySlice[len(hourlySlice)-24:]
	}

	ctx.JSON(http.StatusOK, response.Success(response.LogStatsResponse{
		TotalEntries: totalEntries,
		LevelCounts:  levelCounts,
		HourlyCounts: hourlySlice,
	}))
}
