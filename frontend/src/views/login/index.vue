<template>
  <div class="min-h-screen flex items-center justify-center bg-gray-100">
    <el-card class="w-full max-w-md shadow-xl rounded-xl border-0">
      <template #header>
        <div class="text-center">
          <h2 class="text-2xl font-bold text-gray-800">
            {{ isLogin ? '登录 Code-Repository' : '注册新账号' }}
          </h2>
          <p class="text-sm text-gray-500 mt-2">
            {{ isLogin ? '欢迎回来，请登录您的账户' : '创建一个账户以托管您的代码' }}
          </p>
        </div>
      </template>

      <el-form 
        ref="formRef" 
        :model="formData" 
        :rules="rules" 
        label-position="top" 
        size="large"
        @keyup.enter="handleSubmit"
      >
        <el-form-item label="邮箱地址" prop="email">
          <el-input 
            v-model="formData.email" 
            placeholder="请输入邮箱" 
            prefix-icon="Message"
          />
        </el-form-item>

        <el-form-item label="密码" prop="password">
          <el-input 
            v-model="formData.password" 
            type="password" 
            placeholder="请输入密码" 
            show-password 
            prefix-icon="Lock"
          />
        </el-form-item>

        <el-form-item v-if="!isLogin" label="确认密码" prop="confirmPassword">
          <el-input 
            v-model="formData.confirmPassword" 
            type="password" 
            placeholder="请再次输入密码" 
            show-password 
            prefix-icon="Lock"
          />
        </el-form-item>

        <el-form-item>
          <el-button type="primary" class="w-full mt-4" @click="handleSubmit">
            {{ isLogin ? '登 录' : '注 册' }}
          </el-button>
        </el-form-item>
      </el-form>

      <div class="text-center mt-2">
        <el-link type="primary" :underline="false" @click="toggleMode">
          {{ isLogin ? '没有账号？点击注册' : '已有账号？点击登录' }}
        </el-link>
      </div>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { useRouter } from 'vue-router'
import { useUserStore } from '../../store/user'
// 引入我们刚刚写的真实 API
import { loginAPI, registerAPI } from '../../api/user'

const router = useRouter()
const userStore = useUserStore()
const isLogin = ref(true)
const formRef = ref<FormInstance>()

const formData = reactive({
  email: '',
  password: '',
  confirmPassword: ''
})

const validatePass2 = (rule: any, value: any, callback: any) => {
  if (value === '') {
    callback(new Error('请再次输入密码'))
  } else if (value !== formData.password) {
    callback(new Error('两次输入密码不一致!'))
  } else {
    callback()
  }
}

const rules = reactive<FormRules>({
  email: [
    { required: true, message: '请输入邮箱地址', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: ['blur', 'change'] }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能小于 6 位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, validator: validatePass2, trigger: 'blur' }
  ]
})

const toggleMode = () => {
  isLogin.value = !isLogin.value
  formRef.value?.resetFields()
}

// 真正的前后端联调提交逻辑
const handleSubmit = () => {
  formRef.value?.validate(async (valid) => {
    if (valid) {
      try {
        if (isLogin.value) {
          // 1. 发起真实登录请求
          const res: any = await loginAPI({
            email: formData.email,
            password: formData.password
          })
          ElMessage.success('登录成功！正在进入大厅...')
          // 2. 保存后端返回的真实 Token
          userStore.setToken(res.token)
          // 3. 跳转
          router.push('/dashboard')
        } else {
          // 发起真实注册请求
          await registerAPI({
            email: formData.email,
            password: formData.password
          })
          ElMessage.success('注册成功！请登录')
          toggleMode() // 注册成功后自动切回登录状态
        }
      } catch (error: any) {
        // 请求失败的错误提示已在 utils/request.ts 中统一处理，这里无需额外写 ElMessage
        console.error('操作失败', error)
      }
    } else {
      ElMessage.error('请检查表单填写是否有误')
      return false
    }
  })
}
</script>