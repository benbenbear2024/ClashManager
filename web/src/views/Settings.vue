<template>
  <div class="settings-page">
    <!-- DNS 设置卡片 -->
    <el-card shadow="never" class="settings-card">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <div class="header-icon">
              <el-icon><Setting /></el-icon>
            </div>
            <span>DNS 设置</span>
          </div>
          <el-button type="primary" :icon="Check" @click="handleSave" :loading="saving" size="large">
            保存配置
          </el-button>
        </div>
      </template>

      <el-form :model="dnsForm" label-width="140px" v-loading="loading" class="settings-form">
        <!-- 基础设置部分 -->
        <div class="section-wrapper">
          <div class="section-header">
            <div class="section-icon">
              <el-icon><Operation /></el-icon>
            </div>
            <span class="section-title">基础设置</span>
          </div>

          <div class="section-content">
            <div class="setting-item">
              <div class="setting-label">
                <el-icon><Switch /></el-icon>
                <span>启用 DNS</span>
              </div>
              <div class="setting-control">
                <el-switch v-model="dnsForm.enable" size="large" />
                <span class="setting-hint">关闭后将使用系统DNS</span>
              </div>
            </div>

            <div class="setting-item">
              <div class="setting-label">
                <el-icon><Location /></el-icon>
                <span>监听地址</span>
              </div>
              <div class="setting-control">
                <el-input v-model="dnsForm.listen" placeholder="0.0.0.0:53" class="setting-input" />
                <span class="setting-hint">DNS服务器监听地址</span>
              </div>
            </div>

            <div class="setting-item">
              <div class="setting-label">
                <el-icon><MagicStick /></el-icon>
                <span>增强模式</span>
              </div>
              <div class="setting-control">
                <el-select v-model="dnsForm.enhancedMode" class="setting-select" size="large">
                  <el-option label="Fake-IP" value="fake-ip">
                    <div class="option-content">
                      <span class="option-label">Fake-IP</span>
                      <span class="option-desc">返回假的IP地址，速度快</span>
                    </div>
                  </el-option>
                  <el-option label="Redir-Host" value="redir-host">
                    <div class="option-content">
                      <span class="option-label">Redir-Host</span>
                      <span class="option-desc">返回真实IP，兼容性好</span>
                    </div>
                  </el-option>
                  <el-option label="Mapping" value="mapping">
                    <div class="option-content">
                      <span class="option-label">Mapping</span>
                      <span class="option-desc">使用映射表，介于两者之间</span>
                    </div>
                  </el-option>
                </el-select>
              </div>
            </div>
          </div>
        </div>

        <!-- 主DNS服务器 -->
        <div class="section-wrapper">
          <div class="section-header">
            <div class="section-icon primary">
              <el-icon><Connection /></el-icon>
            </div>
            <span class="section-title">主DNS服务器</span>
            <el-tag size="small" type="primary" effect="plain">必填</el-tag>
          </div>

          <div class="section-content">
            <div class="dns-add-area">
              <el-input
                v-model="nameserverInput"
                placeholder="输入DNS地址（如：223.5.5.5）"
                @keyup.enter="addNameserver"
                clearable
                class="dns-input"
                size="large"
              >
                <template #prefix>
                  <el-icon><Search /></el-icon>
                </template>
                <template #append>
                  <el-button @click="addNameserver" :icon="Plus">添加</el-button>
                </template>
              </el-input>

              <div class="quick-add-buttons">
                <span class="quick-label">快速添加</span>
                <el-button size="small" @click="quickAddNameserver('223.5.5.5')" class="quick-btn aliyun">
                  <el-icon><Connection /></el-icon>
                  阿里DNS
                </el-button>
                <el-button size="small" @click="quickAddNameserver('119.29.29.29')" class="quick-btn tencent">
                  <el-icon><Platform /></el-icon>
                  腾讯DNS
                </el-button>
                <el-button size="small" @click="quickAddNameserver('180.76.76.76')" class="quick-btn baidu">
                  <el-icon><Promotion /></el-icon>
                  百度DNS
                </el-button>
              </div>
            </div>

            <div class="dns-list" v-if="dnsForm.nameserver.length > 0">
              <div v-for="(dns, index) in dnsForm.nameserver" :key="index" class="dns-item">
                <div class="dns-item-left">
                  <div class="dns-icon">
                    <el-icon><Connection /></el-icon>
                  </div>
                  <div class="dns-info">
                    <span class="dns-ip">{{ dns }}</span>
                    <span v-if="isCommonDNS(dns)" class="dns-name">{{ getDNSName(dns) }}</span>
                  </div>
                </div>
                <el-button link type="danger" @click="removeNameserver(index)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>

            <div v-else class="empty-state">
              <el-icon><Connection /></el-icon>
              <p>暂无配置，请添加主DNS服务器</p>
            </div>
          </div>
        </div>

        <!-- 备用DNS服务器 -->
        <div class="section-wrapper">
          <div class="section-header">
            <div class="section-icon success">
              <el-icon><Connection /></el-icon>
            </div>
            <span class="section-title">备用DNS服务器</span>
            <el-tag size="small" type="success" effect="plain">可选</el-tag>
          </div>

          <div class="section-content">
            <div class="dns-add-area">
              <el-input
                v-model="fallbackInput"
                placeholder="输入备用DNS地址"
                @keyup.enter="addFallback"
                clearable
                class="dns-input"
                size="large"
              >
                <template #prefix>
                  <el-icon><Search /></el-icon>
                </template>
                <template #append>
                  <el-button @click="addFallback" :icon="Plus">添加</el-button>
                </template>
              </el-input>

              <div class="quick-add-buttons">
                <span class="quick-label">快速添加</span>
                <el-button size="small" @click="quickAddFallback('8.8.8.8')" class="quick-btn google">
                  <el-icon><Compass /></el-icon>
                  Google
                </el-button>
                <el-button size="small" @click="quickAddFallback('1.1.1.1')" class="quick-btn cloudflare">
                  <el-icon><Share /></el-icon>
                  Cloudflare
                </el-button>
                <el-button size="small" @click="quickAddFallback('208.67.222.222')" class="quick-btn opendns">
                  <el-icon><Guide /></el-icon>
                  OpenDNS
                </el-button>
              </div>
            </div>

            <div class="dns-list" v-if="dnsForm.fallback.length > 0">
              <div v-for="(dns, index) in dnsForm.fallback" :key="index" class="dns-item success">
                <div class="dns-item-left">
                  <div class="dns-icon success">
                    <el-icon><Connection /></el-icon>
                  </div>
                  <div class="dns-info">
                    <span class="dns-ip">{{ dns }}</span>
                    <span v-if="isCommonDNS(dns)" class="dns-name">{{ getDNSName(dns) }}</span>
                  </div>
                </div>
                <el-button link type="danger" @click="removeFallback(index)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>

            <div v-else class="empty-state">
              <el-icon><Connection /></el-icon>
              <p>暂无配置</p>
            </div>
          </div>
        </div>

        <!-- 默认DNS服务器 -->
        <div class="section-wrapper">
          <div class="section-header">
            <div class="section-icon warning">
              <el-icon><Connection /></el-icon>
            </div>
            <span class="section-title">默认DNS服务器</span>
            <el-tooltip content="用于解析其他DNS服务器的地址" placement="top">
              <el-icon class="info-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </div>

          <div class="section-content">
            <div class="dns-add-area">
              <el-input
                v-model="defaultNameserverInput"
                placeholder="用于解析DNS服务器的DNS"
                @keyup.enter="addDefaultNameserver"
                clearable
                class="dns-input"
                size="large"
              >
                <template #prefix>
                  <el-icon><Search /></el-icon>
                </template>
                <template #append>
                  <el-button @click="addDefaultNameserver" :icon="Plus">添加</el-button>
                </template>
              </el-input>

              <div class="quick-add-buttons">
                <span class="quick-label">快速添加</span>
                <el-button size="small" @click="quickAddDefaultNameserver('114.114.114.114')" class="quick-btn">
                  <el-icon><Compass /></el-icon>
                  114DNS
                </el-button>
              </div>
            </div>

            <div class="dns-list" v-if="dnsForm.defaultNameserver.length > 0">
              <div v-for="(dns, index) in dnsForm.defaultNameserver" :key="index" class="dns-item warning">
                <div class="dns-item-left">
                  <div class="dns-icon warning">
                    <el-icon><Connection /></el-icon>
                  </div>
                  <div class="dns-info">
                    <span class="dns-ip">{{ dns }}</span>
                  </div>
                </div>
                <el-button link type="danger" @click="removeDefaultNameserver(index)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>

            <div v-else class="empty-state">
              <el-icon><Connection /></el-icon>
              <p>暂无配置</p>
            </div>
          </div>
        </div>

        <!-- Fake-IP 过滤 -->
        <div class="section-wrapper">
          <div class="section-header">
            <div class="section-icon info">
              <el-icon><Filter /></el-icon>
            </div>
            <span class="section-title">Fake-IP 过滤</span>
            <el-tooltip content="这些域名不会被 Fake-IP 处理" placement="top">
              <el-icon class="info-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </div>

          <div class="section-content">
            <div class="dns-add-area">
              <el-input
                v-model="fakeIPFilterInput"
                placeholder="输入域名（如：*.lan）"
                @keyup.enter="addFakeIPFilter"
                clearable
                class="dns-input"
                size="large"
              >
                <template #prefix>
                  <el-icon><Search /></el-icon>
                </template>
                <template #append>
                  <el-button @click="addFakeIPFilter" :icon="Plus">添加</el-button>
                </template>
              </el-input>
            </div>

            <div class="dns-list" v-if="dnsForm.fakeIPFilter.length > 0">
              <div v-for="(filter, index) in dnsForm.fakeIPFilter" :key="index" class="dns-item info">
                <div class="dns-item-left">
                  <div class="dns-icon info">
                    <el-icon><Filter /></el-icon>
                  </div>
                  <div class="dns-info">
                    <span class="dns-ip">{{ filter }}</span>
                  </div>
                </div>
                <el-button link type="danger" @click="removeFakeIPFilter(index)">
                  <el-icon><Delete /></el-icon>
                </el-button>
              </div>
            </div>

            <div v-else class="empty-state">
              <el-icon><Filter /></el-icon>
              <p>暂无配置</p>
            </div>
          </div>
        </div>
      </el-form>
    </el-card>

    <!-- 配置说明卡片 -->
    <el-card shadow="never" class="info-card">
      <template #header>
        <div class="info-header">
          <el-icon class="info-icon"><InfoFilled /></el-icon>
          <span>配置说明</span>
        </div>
      </template>

      <div class="info-content">
        <div class="info-section">
          <h4><el-icon><Connection /></el-icon> DNS 服务器说明</h4>
          <ul>
            <li><strong>主DNS服务器</strong>：用于常规DNS查询，推荐使用国内DNS如阿里、腾讯</li>
            <li><strong>备用DNS服务器</strong>：当主DNS查询失败时使用，推荐使用国外DNS</li>
            <li><strong>默认DNS服务器</strong>：用于解析其他DNS服务器的地址，通常使用运营商DNS</li>
          </ul>
        </div>

        <div class="info-section">
          <h4><el-icon><MagicStick /></el-icon> 增强模式说明</h4>
          <ul>
            <li><strong>Fake-IP</strong>：返回假的IP地址，速度快，但部分应用可能不兼容</li>
            <li><strong>Redir-Host</strong>：返回真实IP地址，兼容性好但稍慢</li>
            <li><strong>Mapping</strong>：使用映射表，介于两者之间</li>
          </ul>
        </div>
      </div>
    </el-card>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Check,
  Setting,
  Operation,
  Switch,
  Location,
  MagicStick,
  Connection,
  Search,
  Plus,
  Delete,
  Filter,
  QuestionFilled,
  InfoFilled,
  Platform,
  Promotion,
  Compass,
  Share,
  Guide
} from '@element-plus/icons-vue'
import { getDNS, saveDNS } from '@/api/settings'

const loading = ref(false)
const saving = ref(false)

// 输入框
const nameserverInput = ref('')
const fallbackInput = ref('')
const defaultNameserverInput = ref('')
const fakeIPFilterInput = ref('')

const dnsForm = ref({
  enable: true,
  listen: '0.0.0.0:53',
  enhancedMode: 'fake-ip',
  nameserver: [],
  fallback: [],
  defaultNameserver: [],
  fakeIPFilter: []
})

// 常用DNS映射
const dnsMap = {
  '223.5.5.5': '阿里 DNS',
  '119.29.29.29': '腾讯 DNS',
  '180.76.76.76': '百度 DNS',
  '8.8.8.8': 'Google DNS',
  '1.1.1.1': 'Cloudflare DNS',
  '208.67.222.222': 'OpenDNS',
  '114.114.114.114': '114DNS'
}

const isCommonDNS = (ip) => dnsMap[ip] !== undefined

const getDNSName = (ip) => dnsMap[ip] || ip

// 主DNS操作
const addNameserver = () => {
  const dns = nameserverInput.value.trim()
  if (!dns) {
    ElMessage.warning('请输入DNS地址')
    return
  }
  if (dnsForm.value.nameserver.includes(dns)) {
    ElMessage.warning('该DNS已存在')
    return
  }
  dnsForm.value.nameserver.push(dns)
  nameserverInput.value = ''
}

const removeNameserver = (index) => {
  dnsForm.value.nameserver.splice(index, 1)
}

const quickAddNameserver = (dns) => {
  if (dnsForm.value.nameserver.includes(dns)) {
    ElMessage.warning('该DNS已存在')
    return
  }
  dnsForm.value.nameserver.push(dns)
}

// 备用DNS操作
const addFallback = () => {
  const dns = fallbackInput.value.trim()
  if (!dns) {
    ElMessage.warning('请输入DNS地址')
    return
  }
  if (dnsForm.value.fallback.includes(dns)) {
    ElMessage.warning('该DNS已存在')
    return
  }
  dnsForm.value.fallback.push(dns)
  fallbackInput.value = ''
}

const removeFallback = (index) => {
  dnsForm.value.fallback.splice(index, 1)
}

const quickAddFallback = (dns) => {
  if (dnsForm.value.fallback.includes(dns)) {
    ElMessage.warning('该DNS已存在')
    return
  }
  dnsForm.value.fallback.push(dns)
}

// 默认DNS操作
const addDefaultNameserver = () => {
  const dns = defaultNameserverInput.value.trim()
  if (!dns) {
    ElMessage.warning('请输入DNS地址')
    return
  }
  if (dnsForm.value.defaultNameserver.includes(dns)) {
    ElMessage.warning('该DNS已存在')
    return
  }
  dnsForm.value.defaultNameserver.push(dns)
  defaultNameserverInput.value = ''
}

const removeDefaultNameserver = (index) => {
  dnsForm.value.defaultNameserver.splice(index, 1)
}

const quickAddDefaultNameserver = (dns) => {
  if (dnsForm.value.defaultNameserver.includes(dns)) {
    ElMessage.warning('该DNS已存在')
    return
  }
  dnsForm.value.defaultNameserver.push(dns)
}

// FakeIP过滤操作
const addFakeIPFilter = () => {
  const filter = fakeIPFilterInput.value.trim()
  if (!filter) {
    ElMessage.warning('请输入域名')
    return
  }
  if (dnsForm.value.fakeIPFilter.includes(filter)) {
    ElMessage.warning('该过滤规则已存在')
    return
  }
  dnsForm.value.fakeIPFilter.push(filter)
  fakeIPFilterInput.value = ''
}

const removeFakeIPFilter = (index) => {
  dnsForm.value.fakeIPFilter.splice(index, 1)
}

const loadDNS = async () => {
  loading.value = true
  try {
    const data = await getDNS()
    dnsForm.value = {
      enable: data.enable ?? true,
      listen: data.listen || '0.0.0.0:53',
      enhancedMode: data.enhancedMode || 'fake-ip',
      nameserver: data.nameserver || [],
      fallback: data.fallback || [],
      defaultNameserver: data.defaultNameserver || [],
      fakeIPFilter: data.fakeIPFilter || []
    }
  } catch {
    ElMessage.error('加载DNS配置失败')
  } finally {
    loading.value = false
  }
}

const handleSave = async () => {
  if (dnsForm.value.nameserver.length === 0) {
    ElMessage.warning('请至少配置一个主DNS服务器')
    return
  }

  saving.value = true
  try {
    const data = {
      enable: dnsForm.value.enable,
      listen: dnsForm.value.listen,
      enhancedMode: dnsForm.value.enhancedMode,
      nameserver: dnsForm.value.nameserver,
      fallback: dnsForm.value.fallback,
      defaultNameserver: dnsForm.value.defaultNameserver,
      fakeIPFilter: dnsForm.value.fakeIPFilter
    }
    await saveDNS(data)
    ElMessage.success('保存成功')
  } catch {
    ElMessage.error('保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(() => {
  loadDNS()
})
</script>

<style scoped>
.settings-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
  padding-bottom: 20px;
}

.settings-card {
  border-radius: 16px;
  border: none;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

:deep(.settings-card .el-card__header) {
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

.settings-form {
  padding: 24px;
}

/* 分区样式 */
.section-wrapper {
  margin-bottom: 32px;
}

.section-wrapper:last-child {
  margin-bottom: 0;
}

.section-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 20px;
}

.section-icon {
  width: 32px;
  height: 32px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 16px;
}

.section-icon.primary {
  background: linear-gradient(135deg, #409eff 0%, #66b1ff 100%);
}

.section-icon.success {
  background: linear-gradient(135deg, #67c23a 0%, #85ce61 100%);
}

.section-icon.warning {
  background: linear-gradient(135deg, #e6a23c 0%, #f0c78a 100%);
}

.section-icon.info {
  background: linear-gradient(135deg, #909399 0%, #b1b3b8 100%);
}

.section-title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.info-icon {
  color: #909399;
  font-size: 16px;
  cursor: help;
}

.section-content {
  padding-left: 44px;
}

/* 设置项样式 */
.setting-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  background: #fafbfc;
  border-radius: 12px;
  margin-bottom: 12px;
}

.setting-item:last-child {
  margin-bottom: 0;
}

.setting-label {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 14px;
  font-weight: 500;
  color: #606266;
}

.setting-label .el-icon {
  font-size: 18px;
  color: #909399;
}

.setting-control {
  display: flex;
  align-items: center;
  gap: 16px;
}

.setting-input,
.setting-select {
  width: 240px;
}

.setting-hint {
  font-size: 12px;
  color: #909399;
}

/* 下拉选项样式 */
.option-content {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.option-label {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.option-desc {
  font-size: 12px;
  color: #909399;
}

/* DNS添加区域 */
.dns-add-area {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 16px;
}

.dns-input {
  width: 100%;
}

.quick-add-buttons {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.quick-label {
  font-size: 13px;
  color: #909399;
}

.quick-btn {
  display: flex;
  align-items: center;
  gap: 6px;
}

.quick-btn.aliyun {
  color: #ff6a00;
  border-color: #ff6a00;
}

.quick-btn.tencent {
  color: #00a4ff;
  border-color: #00a4ff;
}

.quick-btn.baidu {
  color: #2932e1;
  border-color: #2932e1;
}

.quick-btn.google {
  color: #4285f4;
  border-color: #4285f4;
}

.quick-btn.cloudflare {
  color: #f38020;
  border-color: #f38020;
}

.quick-btn.opendns {
  color: #005bc4;
  border-color: #005bc4;
}

/* DNS列表 */
.dns-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.dns-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 16px;
  background: #f8f9fa;
  border-radius: 10px;
  border-left: 3px solid #667eea;
  transition: all 0.3s ease;
}

.dns-item:hover {
  background: #f0f2f5;
}

.dns-item.success {
  border-left-color: #67c23a;
}

.dns-item.warning {
  border-left-color: #e6a23c;
}

.dns-item.info {
  border-left-color: #909399;
}

.dns-item-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.dns-icon {
  width: 32px;
  height: 32px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  font-size: 14px;
}

.dns-icon.success {
  background: linear-gradient(135deg, #67c23a 0%, #85ce61 100%);
}

.dns-icon.warning {
  background: linear-gradient(135deg, #e6a23c 0%, #f0c78a 100%);
}

.dns-icon.info {
  background: linear-gradient(135deg, #909399 0%, #b1b3b8 100%);
}

.dns-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.dns-ip {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.dns-name {
  font-size: 12px;
  color: #909399;
}

/* 空状态 */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px;
  color: #909399;
}

.empty-state .el-icon {
  font-size: 48px;
  color: #dcdfe6;
  margin-bottom: 12px;
}

.empty-state p {
  font-size: 14px;
  margin: 0;
}

/* 说明卡片 */
.info-card {
  border-radius: 16px;
  border: none;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
}

:deep(.info-card .el-card__header) {
  padding: 18px 24px;
  border-bottom: 1px solid #f0f2f5;
}

.info-header {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.info-header .info-icon {
  color: #409eff;
  font-size: 18px;
}

.info-content {
  padding: 24px;
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 32px;
}

@media (max-width: 1200px) {
  .info-content {
    grid-template-columns: 1fr;
  }
}

.info-section h4 {
  display: flex;
  align-items: center;
  gap: 8px;
  margin: 0 0 16px 0;
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.info-section h4 .el-icon {
  color: #667eea;
}

.info-section ul {
  margin: 0;
  padding-left: 20px;
  list-style: none;
}

.info-section li {
  position: relative;
  padding-left: 16px;
  margin-bottom: 12px;
  font-size: 14px;
  color: #606266;
  line-height: 1.8;
}

.info-section li::before {
  content: '•';
  position: absolute;
  left: 0;
  color: #667eea;
  font-weight: bold;
}
</style>
