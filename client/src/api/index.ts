import axios from 'axios'

// 创建 axios 实例
const api = axios.create({
  baseURL: 'http://localhost:8070', // 发布准备服务端口
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    console.log('发送请求:', config.method?.toUpperCase(), config.url)
    return config
  },
  (error) => {
    console.error('请求错误:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    console.log('收到响应:', response.status, response.config.url)
    return response
  },
  (error) => {
    console.error('响应错误:', error.response?.status, error.message)
    return Promise.reject(error)
  }
)

// API 接口定义
export const apiService = {
  // 验证服务信息
  validateService: (serviceData: any) => {
    return api.post('/validate-service', serviceData)
  },
  
  // 获取服务状态
  getServiceStatus: () => {
    return api.get('/service-status')
  },
  
  // 获取发布计划
  getReleasePlans: () => {
    return api.get('/release-plans')
  },
  
  // 创建发布计划
  createReleasePlan: (planData: any) => {
    return api.post('/release-plans', planData)
  },
  
  // 更新发布计划
  updateReleasePlan: (id: string, planData: any) => {
    return api.put(`/release-plans/${id}`, planData)
  },
  
  // 取消发布计划
  cancelReleasePlan: (id: string) => {
    return api.delete(`/release-plans/${id}`)
  }
}

export default api
