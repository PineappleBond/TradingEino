import { createBrowserRouter, Navigate, Outlet } from 'react-router-dom';
import MainLayout from '../layouts/MainLayout';

// 任务管理页面
import { TaskList, TaskDetail, TaskForm } from '../pages/task';
// 执行记录页面
import { ExecutionList, ExecutionDetail } from '../pages/execution';
// 日志页面
import {
  ExecutionLogList,
  ExecutionLogDetail,
  SystemLogList,
  SystemLogDetail,
} from '../pages/log';

/**
 * 路由配置
 */
export const router = createBrowserRouter([
  {
    path: '/',
    element: <MainLayout><Outlet /></MainLayout>,
    children: [
      {
        index: true,
        element: <Navigate to="/task" replace />,
      },
      {
        path: 'task',
        children: [
          {
            index: true,
            element: <TaskList />,
          },
          {
            path: 'create',
            element: <TaskForm />,
          },
          {
            path: ':id',
            element: <TaskDetail />,
          },
          {
            path: ':id/edit',
            element: <TaskForm />,
          },
        ],
      },
      {
        path: 'task/execution',
        children: [
          {
            index: true,
            element: <ExecutionList />,
          },
          {
            path: ':id',
            element: <ExecutionDetail />,
          },
        ],
      },
      {
        path: 'log/execution',
        children: [
          {
            index: true,
            element: <ExecutionLogList />,
          },
          {
            path: ':id',
            element: <ExecutionLogDetail />,
          },
        ],
      },
      {
        path: 'log/system',
        children: [
          {
            index: true,
            element: <SystemLogList />,
          },
          {
            path: ':filename',
            element: <SystemLogDetail />,
          },
        ],
      },
    ],
  },
]);
