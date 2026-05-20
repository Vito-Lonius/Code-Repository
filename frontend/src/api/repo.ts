import request from '../utils/request'

// 仓库列表查询参数
export interface RepoListParams {
  page?: number
  page_size?: number
  keyword?: string
  owner_id?: number
}

// 创建仓库请求体
export interface CreateRepoData {
  name: string
  description?: string
  is_public: boolean
}

// 获取仓库列表
export const listReposAPI = (params?: RepoListParams) => {
  return request.get('/repos', { params })
}

// 创建仓库
export const createRepoAPI = (data: CreateRepoData) => {
  return request.post('/repos', data)
}

// 获取仓库详情
export const getRepoDetailAPI = (id: number) => {
  return request.get(`/repos/${id}`)
}

// 删除仓库
export const deleteRepoAPI = (id: number) => {
  return request.delete(`/repos/${id}`)
}
