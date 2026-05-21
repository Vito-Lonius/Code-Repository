import request from '../utils/request'

// 登录接口
export const loginAPI = (data: any) => {
  return request.post('/login', data) 
}

// 注册接口
export const registerAPI = (data: any) => {
  return request.post('/register', data)
}