/**
 * 系统日志相关类型定义
 */

/**
 * 日志级别
 */
export type SystemLogLevel = 'INFO' | 'WARN' | 'ERROR' | 'DEBUG';

/**
 * 日志文件信息
 */
export interface LogFileInfo {
  filename: string;
  size: number;
  mod_time: string;
  line_count: number;
  first_log_time?: string;
  last_log_time?: string;
}

/**
 * 日志条目
 */
export interface LogEntry {
  time: string;
  level: SystemLogLevel;
  msg: string;
}

/**
 * 日志文件列表查询参数
 */
export interface LogFileListParams {
  page?: number;
  pageSize?: number;
}

/**
 * 日志内容查询参数
 */
export interface LogContentParams {
  filename: string;
  page?: number;
  pageSize?: number;
  level?: SystemLogLevel;
  start_time?: string;
  end_time?: string;
}

/**
 * 日志搜索参数
 */
export interface SearchLogsParams {
  keyword: string;
  filename?: string;
  level?: SystemLogLevel;
  start_time?: string;
  end_time?: string;
  page?: number;
  pageSize?: number;
}

/**
 * 日志统计信息
 */
export interface LogStats {
  total_entries: number;
  level_counts: Record<string, number>;
  hourly_counts: Array<{
    hour: string;
    count: number;
  }>;
}
