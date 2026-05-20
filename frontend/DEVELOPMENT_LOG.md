# 前端开发记录

## 已完成

### 1. 仓库 API 模块 (`src/api/repo.ts`)
- `listReposAPI` — GET /api/v1/repos 获取仓库列表（分页+搜索）
- `createRepoAPI` — POST /api/v1/repos 创建仓库
- `getRepoDetailAPI` — GET /api/v1/repos/:id 获取仓库详情
- `deleteRepoAPI` — DELETE /api/v1/repos/:id 删除仓库

### 2. 新建仓库弹窗 (`src/components/CreateRepoDialog.vue`)
- 表单：仓库名（必填）、描述（选填）、可见性（公开/私有）
- 表单校验：名称 1-100 字符，描述 ≤255 字符
- 创建成功后关闭弹窗，触发列表刷新

### 3. Dashboard 页面改造 (`src/views/dashboard/index.vue`)
- `onMounted` 调用 `listReposAPI` 拉取真实数据
- 空状态提示（无仓库时引导创建）
- 加载中状态（Loading 图标）
- 搜索防抖（300ms 后触发请求）
- 分页切换（支持 10/20/30 条每页）
- 时间格式化（刚刚 / N分钟前 / N小时前 / N天前）
- 创建仓库后自动刷新列表
- 仓库卡片可点击 → 跳转 `/repo/:id` 详情页

### 4. 路由守卫 (`src/router/index.ts`)
- `beforeEach` 检查 localStorage 中 token
- 未登录访问需认证页面 → 跳转 `/login`
- 已登录访问登录页 → 跳转 `/dashboard`

### 5. 仓库详情页 (`src/views/repo/index.vue`)
- 顶部导航栏（与 Dashboard 一致）
- Breadcrumb 路径导航
- 仓库基本信息（名称、可见性、分支、描述、星标、克隆地址）
- 左侧目录树（el-tree，支持全部展开/折叠）
- 右侧文件列表（el-table：名称、大小、最后更新）
- README 渲染区
- 目前文件树和 README 使用 Mock 数据，后端文件 API 就绪后替换

---

## 待完成

### 后端（非前端职责）

| 事项 | 位置 | 说明 |
|------|------|------|
| JWT 中间件 | `backend/internal/api/middleware/` | 整个中间件目录不存在，无法从 token 解析 userID |
| 注册仓库列表路由 | `backend/cmd/server/main.go` | 缺少 `GET /api/v1/repos` |
| 新增 List handler | `backend/internal/api/v1/repo_handler.go` | 需新增 `List` 方法 |
| Service 暴露 ListRepos | `backend/internal/service/repo_service.go` | 接口和实现均缺 `ListRepos` |
| 文件上传/下载 API | 后端 | 无任何文件相关 HTTP 端点 |
| Git 文件树/提交历史 API | 后端 | 无文件树查询、无 commit 历史端点 |
| 注册 Profile 路由 | `backend/cmd/server/main.go` | handler 已写但未注册 |
| Create handler userID | `repo_handler.go:32` | 硬编码为 1，需从 JWT 获取 |

### 前端（后续可做）

| 事项 | 说明 |
|------|------|
| 文件上传 | 拖拽/选择器上传，大文件分块，进度条 |
| 文件预览 | 代码高亮、图片缩放、PDF 分页、音视频播放 |
| 分支与提交历史页 | 时间轴列表、Hash 标签 |
| 合并请求 (MR) 页 | Diff 视图、SonarQube 质量门禁联动 |
| 仓库设置页 | 基础信息、协作者、分支保护、删除 |
| 用户个人资料 | 头像、昵称修改 |
| Dashboard 侧边栏动态数据 | 目前仍是写死的 mock 数据 |
| 详情页接入真实文件树 | 替换当前 Mock 数据为后端文件 API |
