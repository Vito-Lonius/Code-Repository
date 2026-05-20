import axios from 'axios'
import { ElMessage } from 'element-plus'
import router from '../router'

// 创建 axios 实例
const service = axios.create({
  baseURL: '/api/v1', // 对应后端核心 Endpoint 预览
  timeout: 10000 // 请求超时时间 10 秒
})

// 请求拦截器：自动注入 JWT Token
service.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers['Authorization'] = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器：统一处理错误和 401 过期
service.interceptors.response.use(
  (response) => {
    // 这里可以直接返回 data，剥离外层无用信息
    return response.data
  },
  (error) => {
    if (error.response?.status === 401) {
      ElMessage.error('登录已过期，请重新登录')
      localStorage.removeItem('token')
      router.push('/login') // 自动跳转至登录页
    } else {
      ElMessage.error(error.message || '服务器请求异常')
    }
    return Promise.reject(error)
  }
)

export default service