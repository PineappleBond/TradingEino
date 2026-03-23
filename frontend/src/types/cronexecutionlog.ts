/**
 * 执行日志相关类型定义
 */

/**
 * 日志级别
 */
export type LogLevel = 'info' | 'warn' | 'error' | 'debug';

/**
 * 执行日志详情
 */
export interface CronExecutionLog {
  id: number;
  execution_id: number;
  from: string;
  level: LogLevel;
  message: string;
  created_at: string;
  updated_at: string;
}

/**
 * 执行日志列表查询参数
 */
export interface ExecutionLogListParams {
  page?: number;
  pageSize?: number;
  execution_id?: number;
  level?: LogLevel;
  from?: string;
}
