import type { RequestConfig } from '@umijs/max';
import { notification } from 'antd';

// 运行时配置
export const layout = () => {
  return {
    title: 'q-deploy',
    menu: {
      locale: false,
    },
  };
};

// 后端统一响应格式: { code: number, message: string, data?: any }
export const request: RequestConfig = {
  baseURL: '/api',
  errorConfig: {
    errorThrower: (res: any) => {
      const { code, message } = res;
      if (code !== 0) {
        const error: any = new Error(message);
        error.name = 'BizError';
        error.info = res;
        throw error;
      }
    },
    errorHandler: (error: any) => {
      if (error.name === 'BizError') {
        const errorInfo = error.info;
        notification.error({
          message: '请求失败',
          description: errorInfo.message,
        });
      } else if (error.response) {
        notification.error({
          message: `HTTP ${error.response.status}`,
          description: '网络请求异常',
        });
      } else {
        notification.error({
          message: '网络异常',
          description: '网络连接失败，请检查网络',
        });
      }
    },
  },
};
