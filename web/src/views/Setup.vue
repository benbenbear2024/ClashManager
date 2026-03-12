<template>
  <div class="setup-page">
    <div class="setup-bg">
      <div class="bg-circle circle-1"></div>
      <div class="bg-circle circle-2"></div>
      <div class="bg-circle circle-3"></div>
    </div>
    <div class="setup-container">
      <div class="setup-box">
        <div class="setup-header">
          <div class="logo-wrapper">
            <div class="logo-icon">
              <svg viewBox="0 0 1024 1024" width="36" height="36">
                <path fill="currentColor" d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64z" opacity="0.2"/>
                <path fill="currentColor" d="M765.9 186.2c-119.4-114.7-308.9-111.7-424.4 6.9S227.5 496.7 346.9 611.4c119.4 114.7 308.9 111.7 424.4-6.9s113.9-303.6-5.5-418.3zM393.3 657.6c-91.8-88.1-94.4-233.6-5.8-324.9s235.2-95.3 327 7.1c91.8 102.4 89.2 247.9-2.6 336.2-88.5 85.2-230.4 83.5-318.6-18.4z"/>
                <path fill="currentColor" d="M512 320c-17.7 0-32 14.3-32 32v160c0 17.7 14.3 32 32 32s32-14.3 32-32V352c0-17.7-14.3-32-32-32z"/>
              </svg>
            </div>
          </div>
          <h1>系统初始化</h1>
          <p>创建管理员账号</p>
        </div>

        <div class="welcome-card">
          <div class="welcome-icon">
            <el-icon><InfoFilled /></el-icon>
          </div>
          <div class="welcome-content">
            <h3>欢迎使用 Clash Manager</h3>
            <p>这是您第一次使用系统，请创建管理员账号开始使用。</p>
          </div>
        </div>

        <el-form :model="setupForm" :rules="rules" ref="setupFormRef">
          <el-form-item prop="username">
            <el-input
              v-model="setupForm.username"
              placeholder="管理员用户名"
              size="large"
            >
              <template #prefix>
                <el-icon><User /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item prop="password">
            <el-input
              v-model="setupForm.password"
              type="password"
              placeholder="密码（至少6位）"
              size="large"
              show-password
            >
              <template #prefix>
                <el-icon><Lock /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item prop="confirmPassword">
            <el-input
              v-model="setupForm.confirmPassword"
              type="password"
              placeholder="确认密码"
              size="large"
              show-password
              @keyup.enter="handleSetup"
            >
              <template #prefix>
                <el-icon><Lock /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              size="large"
              class="setup-btn"
              :loading="loading"
              @click="handleSetup"
            >
              <span v-if="!loading">创建管理员账号</span>
              <span v-else>创建中...</span>
            </el-button>
          </el-form-item>

          <el-form-item>
            <el-button
              size="large"
              class="back-btn"
              @click="goToLogin"
            >
              返回登录
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, InfoFilled } from '@element-plus/icons-vue'
import { setup } from '@/api/auth'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const setupFormRef = ref()
const loading = ref(false)

const setupForm = ref({
  username: '',
  password: '',
  confirmPassword: ''
})

const validateConfirmPassword = (rule, value, callback) => {
  if (value === '') {
    callback(new Error('请再次输入密码'))
  } else if (value !== setupForm.value.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 20, message: '用户名长度在 3 到 20 个字符', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 6, message: '密码长度不能少于 6 位', trigger: 'blur' }
  ],
  confirmPassword: [
    { required: true, validator: validateConfirmPassword, trigger: 'blur' }
  ]
}

const handleSetup = async () => {
  const valid = await setupFormRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    await setup({
      username: setupForm.value.username,
      password: setupForm.value.password
    })
    ElMessage.success('管理员账号创建成功，请登录')
    router.push('/login')
  } catch {
    // Error already handled by request interceptor
  } finally {
    loading.value = false
  }
}

const goToLogin = () => {
  router.push('/login')
}
</script>

<style scoped>
.setup-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1a1f3a 0%, #2d3748 50%, #1a1f3a 100%);
  position: relative;
  overflow: hidden;
}

.setup-bg {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  overflow: hidden;
  z-index: 0;
}

.bg-circle {
  position: absolute;
  border-radius: 50%;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.3), rgba(118, 75, 162, 0.3));
  animation: float 20s ease-in-out infinite;
}

.circle-1 {
  width: 500px;
  height: 500px;
  top: -200px;
  left: -100px;
  animation-delay: 0s;
}

.circle-2 {
  width: 400px;
  height: 400px;
  bottom: -150px;
  right: -50px;
  animation-delay: 5s;
}

.circle-3 {
  width: 300px;
  height: 300px;
  top: 50%;
  right: 20%;
  animation-delay: 10s;
}

@keyframes float {
  0%, 100% {
    transform: translateY(0) rotate(0deg);
    opacity: 0.5;
  }
  50% {
    transform: translateY(-30px) rotate(180deg);
    opacity: 0.8;
  }
}

.setup-container {
  width: 100%;
  max-width: 440px;
  padding: 20px;
  position: relative;
  z-index: 1;
}

.setup-box {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 20px;
  padding: 44px 40px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3), 0 0 0 1px rgba(255, 255, 255, 0.1);
}

.setup-header {
  text-align: center;
  margin-bottom: 32px;
}

.logo-wrapper {
  display: flex;
  justify-content: center;
  margin-bottom: 20px;
}

.logo-icon {
  width: 64px;
  height: 64px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  box-shadow: 0 8px 24px rgba(102, 126, 234, 0.4);
}

.setup-header h1 {
  margin: 0 0 8px;
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.setup-header p {
  margin: 0;
  color: #6b7280;
  font-size: 14px;
}

.welcome-card {
  display: flex;
  gap: 16px;
  padding: 16px;
  background: linear-gradient(135deg, rgba(102, 126, 234, 0.1) 0%, rgba(118, 75, 162, 0.1) 100%);
  border-radius: 12px;
  margin-bottom: 28px;
  border: 1px solid rgba(102, 126, 234, 0.2);
}

.welcome-icon {
  width: 40px;
  height: 40px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 20px;
  flex-shrink: 0;
}

.welcome-content h3 {
  margin: 0 0 4px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.welcome-content p {
  margin: 0;
  font-size: 13px;
  color: #6b7280;
  line-height: 1.5;
}

:deep(.el-form-item) {
  margin-bottom: 20px;
}

:deep(.el-input--large .el-input__wrapper) {
  padding: 10px 15px;
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.04);
  transition: all 0.3s ease;
}

:deep(.el-input--large .el-input__wrapper:hover) {
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.15);
}

:deep(.el-input--large .el-input__wrapper.is-focus) {
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
}

:deep(.el-input__prefix) {
  font-size: 18px;
  color: #9ca3af;
}

.setup-btn {
  width: 100%;
  height: 46px;
  border-radius: 10px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  font-size: 16px;
  font-weight: 600;
  box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
  transition: all 0.3s ease;
}

.setup-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.5);
}

.setup-btn:active {
  transform: translateY(0);
}

.back-btn {
  width: 100%;
  height: 46px;
  border-radius: 10px;
  background: #f5f7fa;
  border: 1px solid #e4e7ed;
  color: #606266;
  font-size: 15px;
  font-weight: 500;
  transition: all 0.3s ease;
}

.back-btn:hover {
  background: #e8eaed;
  border-color: #dcdfe6;
}
</style>
