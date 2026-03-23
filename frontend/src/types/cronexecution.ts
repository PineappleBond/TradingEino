/**
 * 执行记录相关类型定义
 */

/**
 * 执行状态
 */
export type ExecutionStatus = 'pending' | 'running' | 'success' | 'failed' | 'retried' | 'cancelled';

/**
 * 执行记录详情
 */
export interface CronExecution {
  id: number;
  task_id: number;
  scheduled_at: string;
  started_at?: string;
  completed_at?: string;
  status: ExecutionStatus;
  retry_count: number;
  error: string;
  created_at: string;
  updated_at: string;
}

/**
 * 执行记录列表查询参数
 */
export interface ExecutionListParams {
  page?: number;
  pageSize?: number;
  task_id?: number;
  status?: ExecutionStatus;
  start_time?: string;
  end_time?: string;
}
