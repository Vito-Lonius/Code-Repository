<template>
  <el-container class="min-h-screen bg-gray-50">
    <el-header class="bg-white shadow-sm flex items-center justify-between px-6 h-16">
      <div class="flex items-center gap-3">
        <div class="w-8 h-8 bg-blue-600 rounded-md flex items-center justify-center text-white font-bold">
          <el-icon><Platform /></el-icon>
        </div>
        <span class="text-xl font-bold text-gray-800">Code-Repository</span>
      </div>
      
      <div class="flex items-center gap-6">
        <el-button type="primary" icon="Plus" @click="createDialogRef?.open()">新建仓库</el-button>
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

    <el-container class="max-w-7xl mx-auto w-full mt-6 gap-6 px-4">
      
      <el-main class="bg-white rounded-xl shadow-sm p-6 overflow-hidden !pt-5">
        <div class="flex justify-between items-center mb-6">
          <h2 class="text-lg font-bold text-gray-800">我的仓库</h2>
          <el-input
            v-model="searchQuery"
            placeholder="搜索仓库名称..."
            prefix-icon="Search"
            class="w-64"
            clearable
            @input="handleSearch"
          />
        </div>

        <!-- 加载状态 -->
        <div v-if="loading" class="text-center py-12">
          <el-icon class="is-loading text-3xl text-blue-500"><Loading /></el-icon>
          <p class="text-gray-400 mt-2">正在加载仓库列表...</p>
        </div>

        <!-- 空状态 -->
        <div v-else-if="repos.length === 0" class="text-center py-16">
          <el-icon class="text-5xl text-gray-300"><FolderOpened /></el-icon>
          <p class="text-gray-400 mt-4 text-lg">还没有仓库</p>
          <p class="text-gray-300 text-sm mt-1">点击上方「新建仓库」按钮创建你的第一个仓库</p>
        </div>

        <!-- 仓库列表 -->
        <div v-else class="space-y-4">
          <el-card
            v-for="repo in repos"
            :key="repo.id"
            shadow="hover"
            class="cursor-pointer border-gray-100"
            @click="router.push(`/repo/${repo.id}`)"
          >
            <div class="flex justify-between items-start">
              <div>
                <h3 class="text-blue-600 font-semibold text-lg hover:underline flex items-center gap-2">
                  <el-icon class="text-gray-400"><FolderOpened /></el-icon>
                  {{ repo.owner_nickname || '用户' }}/{{ repo.name }}
                </h3>
                <p class="text-gray-500 text-sm mt-2 line-clamp-1">
                  {{ repo.description || '暂无描述' }}
                </p>
              </div>
              <el-tag size="small" :type="repo.is_public ? 'success' : 'info'">
                {{ repo.is_public ? '公开' : '私有' }}
              </el-tag>
            </div>
            
            <div class="flex items-center gap-5 mt-4 text-xs text-gray-500 font-medium">
              <span class="flex items-center gap-1 hover:text-blue-600">
                <el-icon><Star /></el-icon> {{ repo.star_count }}
              </span>
              <span class="flex items-center gap-1">
                <el-icon><GitBranch /></el-icon> {{ repo.default_branch }}
              </span>
              <span class="flex items-center gap-1">
                <el-icon><Clock /></el-icon> {{ formatTime(repo.updated_at) }}
              </span>
            </div>
          </el-card>
        </div>

        <!-- 分页 -->
        <div v-if="total > 0" class="flex justify-center mt-6">
          <el-pagination
            v-model:current-page="currentPage"
            v-model:page-size="pageSize"
            :total="total"
            :page-sizes="[10, 20, 30]"
            layout="total, sizes, prev, pager, next"
            @size-change="fetchRepos"
            @current-change="fetchRepos"
          />
        </div>
      </el-main>

      <el-aside width="320px" class="bg-white rounded-xl shadow-sm p-6 hidden lg:block">
        <h3 class="font-bold text-gray-800 mb-6 flex items-center gap-2">
          <el-icon class="text-blue-500"><Bell /></el-icon>
          最近动态
        </h3>
        
        <el-timeline>
          <el-timeline-item timestamp="10 分钟前" type="primary" size="large">
            <p class="text-sm text-gray-700">推送了 3 个 commits 到 <span class="text-blue-600 font-medium">main</span> 分支</p>
          </el-timeline-item>
          <el-timeline-item timestamp="2 小时前" type="success">
            <p class="text-sm text-gray-700">SonarQube 代码质量分析 <span class="font-bold text-green-600">通过</span></p>
          </el-timeline-item>
          <el-timeline-item timestamp="昨天" type="warning">
            <p class="text-sm text-gray-700">严晨 发起了合并请求 (MR) #12</p>
          </el-timeline-item>
          <el-timeline-item timestamp="2 天前" type="info">
            <p class="text-sm text-gray-700">齐赫然 下载了打包文件 <span class="text-gray-500 text-xs">release.zip</span></p>
          </el-timeline-item>
        </el-timeline>
      </el-aside>
      
    </el-container>

    <CreateRepoDialog ref="createDialogRef" @success="fetchRepos" />
  </el-container>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useUserStore } from '../../store/user'
import { ElMessage } from 'element-plus'
import { listReposAPI } from '../../api/repo'
import CreateRepoDialog from '../../components/CreateRepoDialog.vue'

const router = useRouter()
const userStore = useUserStore()

const searchQuery = ref('')
const repos = ref<any[]>([])
const loading = ref(false)
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(10)
const createDialogRef = ref<InstanceType<typeof CreateRepoDialog>>()

let searchTimer: ReturnType<typeof setTimeout> | null = null

const fetchRepos = async () => {
  loading.value = true
  try {
    const res: any = await listReposAPI({
      page: currentPage.value,
      page_size: pageSize.value,
      keyword: searchQuery.value
    })
    repos.value = res.items || []
    total.value = res.total || 0
  } catch (error) {
    repos.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  if (searchTimer) clearTimeout(searchTimer)
  searchTimer = setTimeout(() => {
    currentPage.value = 1
    fetchRepos()
  }, 300)
}

const formatTime = (timeStr: string) => {
  if (!timeStr) return ''
  const diff = Date.now() - new Date(timeStr).getTime()
  const minutes = Math.floor(diff / 60000)
  const hours = Math.floor(diff / 3600000)
  const days = Math.floor(diff / 86400000)
  
  if (minutes < 1) return '刚刚'
  if (minutes < 60) return `${minutes} 分钟前`
  if (hours < 24) return `${hours} 小时前`
  if (days < 30) return `${days} 天前`
  return timeStr.slice(0, 10)
}

const handleLogout = () => {
  userStore.clearToken()
  ElMessage.success('已安全退出登录')
  router.push('/login')
}

onMounted(() => {
  fetchRepos()
})
</script>

<style scoped>
.el-main {
  --el-main-padding: 0;
}
</style>
