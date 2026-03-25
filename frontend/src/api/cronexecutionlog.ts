/**
 * 执行日志 API
 */
import { get } from '../utils/request';
import type { CronExecutionLog, ExecutionLogListParams } from '../types/cronexecutionlog';
import type { PagedData } from '../utils/request';

/**
 * 获取执行日志列表
 */
export function getExecutionLogList(params: ExecutionLogListParams) {
  return get<PagedData<CronExecutionLog>>('/cronexecutionlog', params);
}

/**
 * 获取执行日志详情
 */
export function getExecutionLogDetail(id: number) {
  return get<CronExecutionLog>(`/cronexecutionlog/${id}`);
}

/**
 * 按执行 ID 获取执行日志列表
 */
export function getExecutionLogByExecutionId(executionId: number, params?: { page?: number; pageSize?: number }) {
  return get<PagedData<CronExecutionLog>>(`/cronexecutionlog/execution/${executionId}`, params);
}
