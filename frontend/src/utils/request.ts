import axios, { type AxiosInstance } from 'axios';
import { message } from 'antd';

/**
 * API 响应基础结构
 */
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
}

/**
 * 分页信息
 */
export interface PageInfo {
  page: number;
  pageSize: number;
  total: number;
}

/**
 * 分页数据
 */
export interface PagedData<T> {
  items: T[];
  page: PageInfo;
}

/**
 * 创建 Axios 实例
 */
const request: AxiosInstance = axios.create({
  baseURL: '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

/**
 * 请求拦截器
 */
request.interceptors.request.use(
  (config) => {
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

/**
 * 响应拦截器
 */
request.interceptors.response.use(
  (response) => {
    const res = response.data;

    // 如果响应码不为 0（成功），显示错误消息
    if (res.code !== 0) {
      message.error(res.message || '请求失败');
      return Promise.reject(new Error(res.message || '请求失败'));
    }

    // 返回完整的 response 对象，保持类型一致
    return response;
  },
  (error) => {
    // HTTP 错误处理
    const status = error.response?.status;
    let errorMsg = '网络错误';

    switch (status) {
      case 400:
        errorMsg = '请求参数错误';
        break;
      case 401:
        errorMsg = '未授权，请登录';
        break;
      case 403:
        errorMsg = '拒绝访问';
        break;
      case 404:
        errorMsg = '请求资源不存在';
        break;
      case 500:
        errorMsg = '服务器内部错误';
        break;
      case 502:
        errorMsg = '网关错误';
        break;
      case 503:
        errorMsg = '服务不可用';
        break;
      case 504:
        errorMsg = '网关超时';
        break;
      default:
        errorMsg = error.message || '网络错误';
    }

    message.error(errorMsg);
    return Promise.reject(error);
  }
);

/**
 * 封装 GET 请求
 */
export function get<T = any>(url: string, params?: any): Promise<ApiResponse<T>> {
  return request.get(url, { params }).then((response) => response.data as unknown as ApiResponse<T>);
}

/**
 * 封装 POST 请求
 */
export function post<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
  return request.post(url, data).then((response) => response.data as unknown as ApiResponse<T>);
}

/**
 * 封装 PUT 请求
 */
export function put<T = any>(url: string, data?: any): Promise<ApiResponse<T>> {
  return request.put(url, data).then((response) => response.data as unknown as ApiResponse<T>);
}

/**
 * 封装 DELETE 请求
 */
export function del<T = any>(url: string): Promise<ApiResponse<T>> {
  return request.delete(url).then((response) => response.data as unknown as ApiResponse<T>);
}

export default request;
