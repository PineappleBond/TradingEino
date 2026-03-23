/**
 * 定时任务相关类型定义
 */

/**
 * 任务状态
 */
export type TaskStatus = 'pending' | 'running' | 'completed' | 'stopped' | 'failed';

/**
 * 任务类型
 */
export type TaskType = 'once' | 'recurring';

/**
 * 执行类型
 */
export type ExecType = 'OKXWatcher';

/**
 * 定时任务详情
 */
export interface CronTask {
  id: number;
  name: string;
  spec: string;
  type: TaskType;
  status: TaskStatus;
  exec_type: ExecType;
  raw: string; // JSON 字符串：{"symbol": "ETH-USDT-SWAP"}
  valid_from: string;
  valid_until: string;
  enabled: boolean;
  max_retries: number;
  timeout_seconds: number;
  last_executed_at?: string;
  next_execution_at?: string;
  total_executions: number;
  created_at: string;
  updated_at: string;
}

/**
 * 创建任务请求
 */
export interface CreateTaskRequest {
  name: string;
  spec: string;
  type: TaskType;
  exec_type: ExecType;
  raw: string;
  valid_from: string;
  valid_until: string;
  enabled: boolean;
  max_retries?: number;
  timeout_seconds?: number;
  next_execution_at?: string;
}

/**
 * 更新任务请求
 */
export interface UpdateTaskRequest {
  name?: string;
  spec?: string;
  type?: TaskType;
  exec_type?: ExecType;
  raw?: string;
  valid_from?: string;
  valid_until?: string;
  enabled?: boolean;
  max_retries?: number;
  timeout_seconds?: number;
  next_execution_at?: string;
}

/**
 * 启动任务请求
 */
export interface StartTaskRequest {
  next_execution_time: string;
}

/**
 * 任务列表查询参数
 */
export interface TaskListParams {
  page?: number;
  pageSize?: number;
  status?: TaskStatus;
  type?: TaskType;
  enabled?: boolean;
}
