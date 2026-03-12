<template>
  <div class="login-page">
    <div class="login-bg">
      <div class="bg-circle circle-1"></div>
      <div class="bg-circle circle-2"></div>
      <div class="bg-circle circle-3"></div>
    </div>
    <div class="login-container">
      <div class="login-box">
        <div class="login-header">
          <div class="logo-wrapper">
            <div class="logo-icon">
              <svg viewBox="0 0 1024 1024" width="36" height="36">
                <path fill="currentColor" d="M765.9 186.2c-119.4-114.7-308.9-111.7-424.4 6.9S227.5 496.7 346.9 611.4c119.4 114.7 308.9 111.7 424.4-6.9s113.9-303.6-5.5-418.3zM393.3 657.6c-91.8-88.1-94.4-233.6-5.8-324.9s235.2-95.3 327 7.1c91.8 102.4 89.2 247.9-2.6 336.2-88.5 85.2-230.4 83.5-318.6-18.4z"/>
                <path fill="currentColor" d="M512 320c-17.7 0-32 14.3-32 32v160c0 17.7 14.3 32 32 32s32-14.3 32-32V352c0-17.7-14.3-32-32-32zm0 280c-22.1 0-40 17.9-40 40s17.9 40 40 40 40-17.9 40-40-17.9-40-40-40z"/>
              </svg>
            </div>
          </div>
          <h1>Clash Manager</h1>
          <p>现代化的代理配置管理系统</p>
        </div>

        <el-form :model="loginForm" :rules="rules" ref="loginFormRef" @submit.prevent="handleLogin">
          <el-form-item prop="username">
            <el-input
              v-model="loginForm.username"
              placeholder="用户名"
              size="large"
            >
              <template #prefix>
                <el-icon><User /></el-icon>
              </template>
            </el-input>
          </el-form-item>

          <el-form-item prop="password">
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="密码"
              size="large"
              show-password
              @keyup.enter="handleLogin"
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
              class="login-btn"
              :loading="loading"
              @click="handleLogin"
            >
              <span v-if="!loading">登录</span>
              <span v-else>登录中...</span>
            </el-button>
          </el-form-item>
        </el-form>

        <div class="login-footer">
          <el-button type="text" @click="checkSetup" class="setup-link">
            <el-icon><Key /></el-icon>
            首次使用？点击初始化
          </el-button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Lock, Key } from '@element-plus/icons-vue'
import { login } from '@/api/auth'
import { useUserStore } from '@/stores/user'

const router = useRouter()
const userStore = useUserStore()
const loginFormRef = ref()
const loading = ref(false)

const loginForm = ref({
  username: '',
  password: ''
})

const rules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' }
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' }
  ]
}

const handleLogin = async () => {
  const valid = await loginFormRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const res = await login(loginForm.value)
    userStore.setToken(res.token)
    userStore.setUsername(res.username)
    ElMessage.success('登录成功')
    router.push('/')
  } catch {
    // Error already handled by request interceptor
  } finally {
    loading.value = false
  }
}

const checkSetup = async () => {
  router.push('/setup')
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1a1f3a 0%, #2d3748 50%, #1a1f3a 100%);
  position: relative;
  overflow: hidden;
}

.login-bg {
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

.login-container {
  width: 100%;
  max-width: 420px;
  padding: 20px;
  position: relative;
  z-index: 1;
}

.login-box {
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(20px);
  border-radius: 20px;
  padding: 48px 40px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3), 0 0 0 1px rgba(255, 255, 255, 0.1);
}

.login-header {
  text-align: center;
  margin-bottom: 40px;
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

.login-header h1 {
  margin: 0 0 8px;
  font-size: 28px;
  font-weight: 700;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.login-header p {
  margin: 0;
  color: #6b7280;
  font-size: 14px;
}

:deep(.el-form-item) {
  margin-bottom: 24px;
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

.login-btn {
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

.login-btn:hover {
  transform: translateY(-2px);
  box-shadow: 0 6px 20px rgba(102, 126, 234, 0.5);
}

.login-btn:active {
  transform: translateY(0);
}

.login-footer {
  text-align: center;
  margin-top: 24px;
}

.setup-link {
  color: #667eea;
  font-weight: 500;
  transition: all 0.3s ease;
}

.setup-link:hover {
  color: #764ba2;
}

.setup-link .el-icon {
  margin-right: 4px;
}
</style>
