/**
 * 执行记录 API
 */
import { get } from '../utils/request';
import type { CronExecution, ExecutionListParams } from '../types/cronexecution';
import type { PagedData } from '../utils/request';

/**
 * 获取执行记录列表
 */
export function getExecutionList(params: ExecutionListParams) {
  return get<PagedData<CronExecution>>('/cronexecution', params);
}

/**
 * 获取执行记录详情
 */
export function getExecutionDetail(id: number) {
  return get<CronExecution>(`/cronexecution/${id}`);
}

/**
 * 按任务 ID 获取执行记录列表
 */
export function getExecutionByTaskId(taskId: number, params?: { page?: number; pageSize?: number }) {
  return get<PagedData<CronExecution>>(`/cronexecution/task/${taskId}`, params);
}
