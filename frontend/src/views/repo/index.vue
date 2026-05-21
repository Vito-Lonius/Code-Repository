<template>
  <el-container class="min-h-screen bg-gray-50">
    <!-- 顶部导航 -->
    <el-header class="bg-white shadow-sm flex items-center justify-between px-6 h-16">
      <div class="flex items-center gap-3">
        <div class="w-8 h-8 bg-blue-600 rounded-md flex items-center justify-center text-white font-bold cursor-pointer" @click="$router.push('/dashboard')">
          <el-icon><Platform /></el-icon>
        </div>
        <span class="text-xl font-bold text-gray-800">Code-Repository</span>
      </div>
      
      <div class="flex items-center gap-6">
        <el-dropdown>
          <span class="cursor-pointer flex items-center gap-2 text-gray-700 hover:text-blue-600 transition-colors">
            <el-avatar :size="32" src="https://cube.elemecdn.com/3/7c/3ea6beec64369c2642b92c6726f1epng.png" />
            <span class="font-medium">用户</span>
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item icon="User">个人资料</el-dropdown-item>
              <el-dropdown-item icon="Setting">系统设置</el-dropdown-item>
              <el-dropdown-item icon="SwitchButton" divided @click="handleLogout">
                <span class="text-red-500">退出登录</span>
              </el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </el-header>

    <!-- 仓库信息栏 -->
    <div v-if="repo" class="bg-white border-b px-6 py-3">
      <div class="max-w-7xl mx-auto">
        <!-- Breadcrumb -->
        <el-breadcrumb class="mb-2" separator="/">
          <el-breadcrumb-item :to="{ name: 'Dashboard' }">
            <el-icon><FolderOpened /></el-icon> 仓库
          </el-breadcrumb-item>
          <el-breadcrumb-item>{{ repo.owner_nickname || '用户' }}</el-breadcrumb-item>
          <el-breadcrumb-item class="font-semibold">{{ repo.name }}</el-breadcrumb-item>
        </el-breadcrumb>

        <div class="flex items-center gap-4">
          <h1 class="text-xl font-bold text-gray-800 flex items-center gap-2">
            <el-icon class="text-blue-600"><FolderOpened /></el-icon>
            {{ repo.name }}
          </h1>
          <el-tag size="small" :type="repo.is_public ? 'success' : 'info'">
            {{ repo.is_public ? '公开' : '私有' }}
          </el-tag>
          <span class="text-xs text-gray-400">{{ repo.default_branch }}</span>
        </div>
        <p v-if="repo.description" class="text-sm text-gray-500 mt-1">{{ repo.description }}</p>
      </div>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="flex-1 flex items-center justify-center">
      <el-icon class="is-loading text-3xl text-blue-500"><Loading /></el-icon>
      <span class="ml-2 text-gray-400">加载仓库信息...</span>
    </div>

    <!-- 主体 -->
    <el-container v-else-if="repo" class="max-w-7xl mx-auto w-full mt-4 gap-4 px-4">
      <!-- 左侧目录树 -->
      <el-aside width="260px" class="bg-white rounded-xl shadow-sm p-4">
        <div class="flex items-center justify-between mb-3">
          <h3 class="font-bold text-gray-800 text-sm flex items-center gap-1">
            <el-icon><List /></el-icon> 目录结构
          </h3>
          <el-button text size="small" @click="toggleAllFolders">
            {{ allExpanded ? '全部折叠' : '全部展开' }}
          </el-button>
        </div>
        <el-tree
          :data="fileTree"
          :props="{ children: 'children', label: 'name' }"
          node-key="name"
          :default-expand-all="allExpanded"
          highlight-current
          @node-click="handleFileClick"
        >
          <template #default="{ data }">
            <span class="flex items-center gap-1 text-sm">
              <el-icon v-if="data.type === 'dir'" class="text-amber-500"><Folder /></el-icon>
              <el-icon v-else class="text-gray-400"><Document /></el-icon>
              <span>{{ data.name }}</span>
            </span>
          </template>
        </el-tree>
      </el-aside>

      <!-- 右侧主体区 -->
      <el-main class="bg-white rounded-xl shadow-sm p-6 flex-1 overflow-hidden !pt-5">
        <!-- 文件列表区 -->
        <div class="flex items-center justify-between mb-4 pb-3 border-b">
          <div class="flex items-center gap-4 text-xs text-gray-500">
            <span class="flex items-center gap-1">
              <el-icon><Star /></el-icon> {{ repo.star_count }} 星标
            </span>
            <span>{{ repo.clone_url }}</span>
          </div>
          <div class="flex items-center gap-2">
            <el-button size="small" icon="Upload">上传文件</el-button>
            <el-button size="small" icon="Plus">新建文件</el-button>
          </div>
        </div>

        <!-- 文件列表 -->
        <el-table :data="currentFiles" style="width: 100%" size="small" highlight-current-row>
          <el-table-column label="文件" min-width="300">
            <template #default="{ row }">
              <div class="flex items-center gap-2 cursor-pointer hover:text-blue-600" @click="handleFileClick(row)">
                <el-icon v-if="row.type === 'dir'" class="text-amber-500"><Folder /></el-icon>
                <el-icon v-else class="text-gray-400"><Document /></el-icon>
                <span>{{ row.name }}</span>
              </div>
            </template>
          </el-table-column>
          <el-table-column label="大小" width="120" align="right">
            <template #default="{ row }">
              <span v-if="row.type !== 'dir'" class="text-xs text-gray-400">{{ row.size }}</span>
            </template>
          </el-table-column>
          <el-table-column label="最后更新" width="180" align="right">
            <template #default="{ row }">
              <span class="text-xs text-gray-400">{{ row.updated }}</span>
            </template>
          </el-table-column>
        </el-table>

        <!-- README 预览区 -->
        <div v-if="readmeContent" class="mt-8 border rounded-lg">
          <div class="bg-gray-50 px-4 py-2 border-b flex items-center gap-2">
            <el-icon><Document /></el-icon>
            <span class="font-semibold text-sm">README.md</span>
          </div>
          <div class="p-6 prose max-w-none text-sm" v-html="readmeContent"></div>
        </div>
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useUserStore } from '../../store/user'
import { ElMessage } from 'element-plus'
import { getRepoDetailAPI } from '../../api/repo'

interface FileNode {
  name: string
  type: 'file' | 'dir'
  size?: string
  updated?: string
  children?: FileNode[]
}

const route = useRoute()
const router = useRouter()
const userStore = useUserStore()

const repo = ref<any>(null)
const loading = ref(true)
const allExpanded = ref(true)

// Mock 文件树数据（后端文件 API 就绪后替换）
const mockFiles: FileNode[] = [
  {
    name: 'src', type: 'dir', children: [
      { name: 'main.ts', type: 'file', size: '2.1 KB', updated: '3 天前' },
      { name: 'App.vue', type: 'file', size: '156 B', updated: '昨天' },
      {
        name: 'components', type: 'dir', children: [
          { name: 'HelloWorld.vue', type: 'file', size: '1.8 KB', updated: '上周' },
        ]
      },
      {
        name: 'views', type: 'dir', children: [
          { name: 'dashboard.vue', type: 'file', size: '5.4 KB', updated: '2 小时前' },
          { name: 'login.vue', type: 'file', size: '3.2 KB', updated: '昨天' },
        ]
      },
    ]
  },
  { name: 'public', type: 'dir', children: [
    { name: 'favicon.ico', type: 'file', size: '4.2 KB', updated: '1 个月前' },
  ]},
  { name: 'package.json', type: 'file', size: '1.1 KB', updated: '昨天' },
  { name: 'README.md', type: 'file', size: '3.5 KB', updated: '5 天前' },
  { name: 'tsconfig.json', type: 'file', size: '682 B', updated: '1 个月前' },
]

const fileTree = ref<FileNode[]>(mockFiles)

// 当前目录下的文件（简化处理，展示根目录）
const currentFiles = computed(() => fileTree.value)

// Mock README 内容
const readmeContent = `# Code-Repository

在线代码与文件托管平台，支持仓库管理、文件上传、在线预览和代码质量分析。

## 技术栈

- **前端**: Vue 3 + TypeScript + Vite + Element Plus + Tailwind CSS
- **后端**: Go + Gin + GORM + PostgreSQL
- **存储**: MinIO (S3 兼容)

## 快速开始

\`\`\`bash
# 前端
cd frontend && pnpm install && pnpm dev

# 后端
cd backend && go run cmd/server/main.go
\`\`\`
`

const fetchRepo = async () => {
  loading.value = true
  try {
    const id = Number(route.params.id)
    const res: any = await getRepoDetailAPI(id)
    repo.value = res
  } catch (error) {
    ElMessage.error('加载仓库信息失败')
  } finally {
    loading.value = false
  }
}

const toggleAllFolders = () => {
  allExpanded.value = !allExpanded.value
}

const handleFileClick = (data: FileNode) => {
  if (data.type === 'dir') return
  // 文件点击 → 后续接入文件预览页
  ElMessage.info(`预览文件: ${data.name}（文件预览功能开发中）`)
}

const handleLogout = () => {
  userStore.clearToken()
  ElMessage.success('已安全退出登录')
  router.push('/login')
}

onMounted(() => {
  fetchRepo()
})
</script>

<style scoped>
.el-main {
  --el-main-padding: 0;
}
</style>
