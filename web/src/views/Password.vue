<template>
  <div class="password-page">
    <el-card shadow="never" class="password-card">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <div class="header-icon">
              <el-icon><Lock /></el-icon>
            </div>
            <span>修改密码</span>
          </div>
        </div>
      </template>

      <el-form :model="passwordForm" :rules="rules" ref="passwordFormRef" label-width="120px" class="password-form">
        <el-form-item label="用户名">
          <el-input :value="userStore.username" disabled>
            <template #prefix>
              <el-icon><User /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="原密码" prop="oldPassword">
          <el-input
            v-model="passwordForm.oldPassword"
            type="password"
            placeholder="请输入原密码"
            show-password
          >
            <template #prefix>
              <el-icon><Lock /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="新密码" prop="newPassword">
          <el-input
            v-model="passwordForm.newPassword"
            type="password"
            placeholder="请输入新密码（至少6位）"
            show-password
          >
            <template #prefix>
              <el-icon><Unlock /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item label="确认新密码" prop="confirmPassword">
          <el-input
            v-model="passwordForm.confirmPassword"
            type="password"
            placeholder="请再次输入新密码"
            show-password
          >
            <template #prefix>
              <el-icon><CircleCheck /></el-icon>
            </template>
          </el-input>
        </el-form-item>

        <el-form-item class="form-actions">
          <el-button type="primary" :loading="loading" @click="handleChangePassword" size="large">
            <el-icon v-if="!loading"><Check /></el-icon>
            <span v-if="!loading">修改密码</span>
            <span v-else>修改中...</span>
          </el-button>
          <el-button @click="resetForm" size="large">
            <el-icon><RefreshLeft /></el-icon>
            重置
          </el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import { Lock, Unlock, User, CircleCheck, Check, RefreshLeft } from '@element-plus/icons-vue'
import { changePassword } from '@/api/auth'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()
const passwordFormRef = ref()
const loading = ref(false)

const passwordForm = ref({
  oldPassword: '',
  newPassword: '',
  confirmPassword: ''
})

const validateConfirmPassword = (rule, value, callback) => {
  if (value === '') {
    callback(new Error('请再次输入新密码'))
  } else if (value !== passwordForm.value.newPassword) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules = {
  oldPassword: [
    { required: true, message: '请输入原密码', trigger: 'blur' }
  ],
  newPassword: [
    { required: true, message: '请输入新密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于 6 位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

const handleChangePassword = async () => {
  const valid = await passwordFormRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await changePassword({
      old_password: passwordForm.value.oldPassword,
      new_password: passwordForm.value.newPassword
    })
    ElMessage.success('密码修改成功，请重新登录')
    resetForm()
  } catch {
    // Error already handled
  } finally {
    loading.value = false
  }
}

const resetForm = () => {
  passwordForm.value = {
    oldPassword: '',
    newPassword: '',
    confirmPassword: ''
  }
  passwordFormRef.value?.resetFields()
}
</script>

<style scoped>
.password-page {
  height: 100%;
  display: flex;
  align-items: flex-start;
  justify-content: center;
  padding-top: 40px;
}

.password-card {
  width: 100%;
  max-width: 600px;
  border-radius: 16px;
  border: none;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

:deep(.password-card .el-card__header) {
  padding: 20px 24px;
  border-bottom: 1px solid #f0f2f5;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.05) 0%, rgba(118, 75, 162, 0.05) 100%);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 18px;
  font-weight: 600;
  color: #303133;
}

.header-icon {
  width: 36px;
  height: 36px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 18px;
}

.password-form {
  padding: 24px;
  max-width: 500px;
  margin: 0 auto;
}

:deep(.password-form .el-form-item__label) {
  font-weight: 500;
  color: #606266;
}

:deep(.password-form .el-input__wrapper) {
  border-radius: 8px;
  padding: 8px 12px;
  transition: all 0.3s ease;
}

:deep(.password-form .el-input__wrapper:hover) {
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.15);
}

:deep(.password-form .el-input__wrapper.is-focus) {
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.3);
}

:deep(.password-form .el-input__prefix) {
  font-size: 16px;
  color: #9ca3af;
}

:deep(.password-form .el-input.is-disabled .el-input__wrapper) {
  background: #f5f7fa;
}

.form-actions {
  margin-top: 32px;
  margin-bottom: 0;
}

:deep(.form-actions .el-button) {
  padding: 10px 24px;
  border-radius: 8px;
  font-weight: 500;
}

:deep(.form-actions .el-button--primary) {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

:deep(.form-actions .el-button--primary:hover) {
  transform: translateY(-2px);
  box-shadow: 0 6px 16px rgba(102, 126, 234, 0.4);
}

:deep(.form-actions .el-button--primary:active) {
  transform: translateY(0);
}

:deep(.form-actions .el-btn) {
  display: flex;
  align-items: center;
  gap: 6px;
}
</style>
