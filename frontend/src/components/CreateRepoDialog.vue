<template>
  <el-dialog
    v-model="visible"
    title="新建仓库"
    width="500px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="rules"
      label-position="top"
      size="large"
    >
      <el-form-item label="仓库名称" prop="name">
        <el-input
          v-model="formData.name"
          placeholder="请输入仓库名称（如 my-awesome-project）"
          maxlength="100"
          show-word-limit
        />
      </el-form-item>

      <el-form-item label="描述（选填）" prop="description">
        <el-input
          v-model="formData.description"
          type="textarea"
          placeholder="简要描述您的仓库..."
          maxlength="255"
          show-word-limit
          :rows="3"
        />
      </el-form-item>

      <el-form-item label="可见性" prop="is_public">
        <el-radio-group v-model="formData.is_public">
          <el-radio :value="true">
            <span class="flex items-center gap-1">
              <el-icon><View /></el-icon> 公开
            </span>
          </el-radio>
          <el-radio :value="false">
            <span class="flex items-center gap-1">
              <el-icon><Lock /></el-icon> 私有
            </span>
          </el-radio>
        </el-radio-group>
        <p class="text-xs text-gray-400 mt-1">
          {{ formData.is_public ? '所有人可见，适合开源项目' : '仅你与指定协作者可见' }}
        </p>
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleSubmit">
        创建仓库
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { createRepoAPI } from '../../api/repo'

const emit = defineEmits<{
  success: []
}>()

const visible = ref(false)
const loading = ref(false)
const formRef = ref<FormInstance>()

const formData = reactive({
  name: '',
  description: '',
  is_public: true
})

const rules: FormRules = {
  name: [
    { required: true, message: '请输入仓库名称', trigger: 'blur' },
    { min: 1, max: 100, message: '仓库名称长度在 1 到 100 个字符之间', trigger: 'blur' }
  ],
  description: [
    { max: 255, message: '描述不能超过 255 个字符', trigger: 'blur' }
  ]
}

const open = () => {
  formData.name = ''
  formData.description = ''
  formData.is_public = true
  formRef.value?.resetFields()
  visible.value = true
}

const handleClose = () => {
  formRef.value?.resetFields()
}

const handleSubmit = () => {
  formRef.value?.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      await createRepoAPI({
        name: formData.name,
        description: formData.description,
        is_public: formData.is_public
      })
      ElMessage.success('仓库创建成功！')
      visible.value = false
      emit('success')
    } catch (error) {
      console.error('创建仓库失败', error)
    } finally {
      loading.value = false
    }
  })
}

defineExpose({ open })
</script>
