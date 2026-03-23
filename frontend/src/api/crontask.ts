/**
 * 定时任务 API
 */
import { get, post, put, del } from '../utils/request';
import type { CronTask, CreateTaskRequest, UpdateTaskRequest, StartTaskRequest, TaskListParams } from '../types/crontask';
import type { PagedData } from '../utils/request';

/**
 * 获取任务列表
 */
export function getTaskList(params: TaskListParams) {
  return get<PagedData<CronTask>>('/crontask', params);
}

/**
 * 获取任务详情
 */
export function getTaskDetail(id: number) {
  return get<CronTask>(`/crontask/${id}`);
}

/**
 * 创建任务
 */
export function createTask(data: CreateTaskRequest) {
  return post<CronTask>('/crontask', data);
}

/**
 * 更新任务
 */
export function updateTask(id: number, data: UpdateTaskRequest) {
  return put<CronTask>(`/crontask/${id}`, data);
}

/**
 * 删除任务
 */
export function deleteTask(id: number) {
  return del<string>(`/crontask/${id}`);
}

/**
 * 启用任务
 */
export function enableTask(id: number) {
  return post<string>(`/crontask/${id}/enable`);
}

/**
 * 禁用任务
 */
export function disableTask(id: number) {
  return post<string>(`/crontask/${id}/disable`);
}

/**
 * 启动任务
 */
export function startTask(id: number, data: StartTaskRequest) {
  return post<string>(`/crontask/${id}/start`, data);
}

/**
 * 停止任务
 */
export function stopTask(id: number) {
  return post<string>(`/crontask/${id}/stop`);
}
