import { createRouter, createWebHistory } from 'vue-router'
import { ElMessage } from 'element-plus'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      redirect: '/login'
    },
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/login/index.vue'),
      meta: { requiresAuth: false }
    },
    {
      path: '/dashboard',
      name: 'Dashboard',
      component: () => import('../views/dashboard/index.vue'),
      meta: { requiresAuth: true }
    },
    {
      path: '/repo/:id',
      name: 'RepoDetail',
      component: () => import('../views/repo/index.vue'),
      meta: { requiresAuth: true }
    }
  ]
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')

  if (to.meta.requiresAuth !== false && !token) {
    ElMessage.warning('请先登录')
    next({ name: 'Login' })
    return
  }

  if (to.name === 'Login' && token) {
    next({ name: 'Dashboard' })
    return
  }

  next()
})

export default router
