/**
 * 系统日志 API
 */
import { get } from '../utils/request';
import type { LogFileInfo, LogEntry, LogFileListParams, LogContentParams, SearchLogsParams, LogStats } from '../types/systemlog';
import type { PagedData } from '../utils/request';

/**
 * 获取日志文件列表
 */
export function getLogFileList(params: LogFileListParams) {
  return get<PagedData<LogFileInfo>>('/systemlog/files', params);
}

/**
 * 获取日志文件内容
 */
export function getLogContent(params: LogContentParams) {
  const { filename, ...rest } = params;
  return get<PagedData<LogEntry>>(`/systemlog/files/${filename}`, rest);
}

/**
 * 搜索日志
 */
export function searchLogs(params: SearchLogsParams) {
  return get<PagedData<LogEntry>>('/systemlog/search', params);
}

/**
 * 获取日志统计信息
 */
export function getLogStats(params?: { filename?: string; start_time?: string; end_time?: string }) {
  return get<LogStats>('/systemlog/stats', params);
}
