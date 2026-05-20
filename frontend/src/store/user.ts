import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore('user', () => {
  // 初始化时从本地存储读取 token
  const token = ref(localStorage.getItem('token') || '')

  // 登录成功后保存 token
  const setToken = (newToken: string) => {
    token.value = newToken
    localStorage.setItem('token', newToken)
  }

  // 退出登录时清除 token
  const clearToken = () => {
    token.value = ''
    localStorage.removeItem('token')
  }

  return { token, setToken, clearToken }
})