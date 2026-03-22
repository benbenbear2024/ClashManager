<template>
  <div class="rules-page">
    <el-card shadow="never" class="rules-card">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <div class="header-icon">
              <el-icon><DocumentCopy /></el-icon>
            </div>
            <span>规则列表</span>
          </div>
          <div class="header-actions">
            <el-button :icon="Upload" @click="showImportDialog">导入规则</el-button>
            <el-button type="primary" :icon="Plus" @click="showCreateDialog">新增规则</el-button>
          </div>
        </div>
      </template>

      <!-- 搜索和过滤区域 -->
      <div class="filter-section">
        <div class="filter-left">
          <el-input
            v-model="searchKeyword"
            placeholder="搜索匹配内容或目标"
            clearable
            class="search-input"
            @clear="handleSearchChange"
            @keyup.enter="handleSearchChange"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>

          <el-select
            v-model="filterType"
            placeholder="规则类型"
            clearable
            class="filter-select"
            @change="handleFilterTypeChange"
            @clear="handleFilterTypeClear"
          >
            <el-option label="DOMAIN-SUFFIX" value="DOMAIN-SUFFIX" />
            <el-option label="DOMAIN" value="DOMAIN" />
            <el-option label="DOMAIN-KEYWORD" value="DOMAIN-KEYWORD" />
            <el-option label="IP-CIDR" value="IP-CIDR" />
            <el-option label="SRC-IP-CIDR" value="SRC-IP-CIDR" />
            <el-option label="GEOIP" value="GEOIP" />
            <el-option label="MATCH" value="MATCH" />
          </el-select>

          <el-select
            v-model="filterTarget"
            placeholder="目标类型"
            clearable
            class="filter-select"
            @change="handleFilterTargetChange"
            @clear="handleFilterTargetClear"
          >
            <el-option-group label="固定出口">
              <el-option label="DIRECT - 直连" value="DIRECT">
                <div class="option-content">
                  <el-tag type="success" size="small">DIRECT</el-tag>
                  <span class="option-text">直连</span>
                </div>
              </el-option>
              <el-option label="PROXY - 代理" value="PROXY">
                <div class="option-content">
                  <el-tag type="primary" size="small">PROXY</el-tag>
                  <span class="option-text">代理</span>
                </div>
              </el-option>
              <el-option label="REJECT - 拒绝" value="REJECT">
                <div class="option-content">
                  <el-tag type="danger" size="small">REJECT</el-tag>
                  <span class="option-text">拒绝</span>
                </div>
              </el-option>
            </el-option-group>
            <el-option-group label="代理节点" v-if="nodes.length > 0">
              <el-option
                v-for="node in nodes"
                :key="'node-' + node.id"
                :label="node.name"
                :value="'node:' + node.name"
              >
                <div class="option-content-flex">
                  <el-icon><Connection /></el-icon>
                  <span class="option-name">{{ node.name }}</span>
                  <el-tag size="small" class="option-type-tag">{{ node.type }}</el-tag>
                </div>
              </el-option>
            </el-option-group>
          </el-select>

          <el-button @click="resetFilter" :icon="RefreshLeft">重置</el-button>
        </div>
        <div class="filter-right">
          <span class="total-count">共 {{ total }} 条规则</span>
        </div>
      </div>

      <el-table :data="displayRules" stripe class="rules-table" v-loading="loading">
        <el-table-column prop="id" label="ID" min-width="60" />
        <el-table-column prop="type" label="规则类型" min-width="140">
          <template #default="{ row }">
            <el-tag :type="getTypeTagType(row.type)" size="small">{{ row.type }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="payload" label="匹配内容" min-width="200" show-overflow-tooltip />
        <el-table-column label="拨号状态" min-width="100">
          <template #default="{ row }">
            <div v-if="getRuleDialingStatus(row)">
              <el-tag :type="getRuleDialingStatus(row).type" size="small">
                {{ getRuleDialingStatus(row).text }}
              </el-tag>
            </div>
            <div v-else>
              <el-tag type="info" size="small">-</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="target" label="目标" min-width="140">
          <template #default="{ row }">
            <div class="target-cell">
              <!-- 固定出口 builtin -->
              <div v-if="row.target_type === 'builtin' && row.target === 'PROXY'" class="target-tag proxy">
                <span>PROXY</span>
              </div>
              <div v-else-if="row.target_type === 'builtin' && row.target === 'DIRECT'" class="target-tag direct">
                <span>DIRECT</span>
              </div>
              <div v-else-if="row.target_type === 'builtin' && row.target === 'REJECT'" class="target-tag reject">
                <span>REJECT</span>
              </div>
              <div v-else-if="row.target_type === 'builtin'" class="target-tag builtin">
                <span>{{ row.target }}</span>
              </div>
              <!-- 代理节点 -->
              <div v-else-if="row.target_type === 'node'" class="target-tag node">
                <el-icon><Connection /></el-icon>
                <span>{{ getTargetDisplayName(row) }}</span>
              </div>
              <!-- 代理组 -->
              <div v-else-if="row.target_type === 'group'" class="target-tag group">
                <el-icon><Grid /></el-icon>
                <span>{{ getTargetDisplayName(row) }}</span>
              </div>
              <!-- 其他（兼容旧数据） -->
              <div v-else class="target-tag builtin">
                <span>{{ row.target }}</span>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="备注" min-width="180" show-overflow-tooltip>
          <template #default="{ row }">
            <span v-if="getNodeRemark(row)">{{ getNodeRemark(row) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="150" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link @click="handleEdit(row)">
              <el-icon><Edit /></el-icon>
              编辑
            </el-button>
            <el-button type="danger" link @click="handleDelete(row)">
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-section">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[20, 50, 100, 200]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="handleSizeChange"
          @current-change="handlePageChange"
        />
      </div>
    </el-card>

    <!-- 新增/编辑规则对话框 -->
    <el-dialog v-model="formDialogVisible" :title="isEdit ? '编辑规则' : '新增规则'" width="520px" class="rule-dialog">
      <el-form :model="ruleForm" label-width="110px" class="rule-form">
        <el-form-item label="规则类型">
          <el-select v-model="ruleForm.Type" placeholder="请选择规则类型" class="form-input">
            <el-option label="DOMAIN-SUFFIX - 域名后缀匹配" value="DOMAIN-SUFFIX" />
            <el-option label="DOMAIN - 完整域名匹配" value="DOMAIN" />
            <el-option label="DOMAIN-KEYWORD - 域名关键字匹配" value="DOMAIN-KEYWORD" />
            <el-option label="IP-CIDR - IP段匹配" value="IP-CIDR" />
            <el-option label="SRC-IP-CIDR - 源IP段匹配" value="SRC-IP-CIDR" />
            <el-option label="GEOIP - 地理位置匹配" value="GEOIP" />
            <el-option label="MATCH - 全匹配（默认规则）" value="MATCH" />
          </el-select>
        </el-form-item>
        <el-form-item label="匹配内容">
          <el-input v-model="ruleForm.Payload" placeholder="如: google.com 或 192.168.1.0/24" class="form-input" />
        </el-form-item>
        <el-form-item label="目标">
          <el-select v-model="ruleForm.Target" placeholder="请选择目标" class="form-input" @change="handleTargetChange" filterable>
            <el-option-group label="内置目标">
              <el-option label="PROXY - 代理" value="PROXY">
                <span>PROXY</span>
                <span class="option-desc">代理</span>
              </el-option>
              <el-option label="DIRECT - 直连" value="DIRECT">
                <span>DIRECT</span>
                <span class="option-desc">直连</span>
              </el-option>
              <el-option label="REJECT - 拒绝" value="REJECT">
                <span>REJECT</span>
                <span class="option-desc">拒绝</span>
              </el-option>
            </el-option-group>
            <el-option-group label="代理节点">
              <el-option
                v-for="node in nodes"
                :key="node.id"
                :label="node.name"
                :value="`node:${node.id}:${node.name}`"
              >
                <div class="option-content-flex">
                  <div class="option-left">
                    <el-icon><Connection /></el-icon>
                    <span class="option-name">{{ node.name }}</span>
                    <span v-if="node.rename" class="option-rename">{{ node.rename }}</span>
                  </div>
                  <el-tag size="small" class="option-type-tag">{{ node.type }}</el-tag>
                </div>
              </el-option>
            </el-option-group>
          </el-select>
          <!-- 节点备注显示 -->
          <div v-if="ruleForm.TargetType === 'node' && selectedNodeRemark" class="node-remark">
            <el-tag size="small" type="info">备注: {{ selectedNodeRemark }}</el-tag>
          </div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- 导入规则对话框 -->
    <el-dialog v-model="importDialogVisible" title="导入规则" width="600px" class="import-dialog">
      <el-form label-width="100px" class="import-form">
        <el-form-item label="选择文件">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :show-file-list="true"
            :limit="1"
            accept=".yaml,.yml"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            drag
          >
            <el-icon class="upload-icon"><UploadFilled /></el-icon>
            <div class="upload-text">拖拽文件到此处或点击上传</div>
            <template #tip>
              <div class="upload-tip">支持 .yaml 或 .yml 格式文件，将自动解析 rules 节点</div>
            </template>
          </el-upload>
        </el-form-item>
        <el-form-item label="文件内容" v-if="importContent">
          <el-input
            v-model="importContent"
            type="textarea"
            :rows="10"
            placeholder="文件内容将显示在这里"
            readonly
            class="import-textarea"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="importDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleImport" :loading="importing">导入</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Plus,
  Search,
  Upload,
  RefreshLeft,
  DocumentCopy,
  Connection,
  Grid,
  Edit,
  Delete,
  UploadFilled
} from '@element-plus/icons-vue'
import { getRules, createRule, updateRule, deleteRule, importRules } from '@/api/rules'
import { getNodes } from '@/api/nodes'
import { checkL2TPConnection } from '@/api/mihomo'

const rules = ref([])
const nodes = ref([])
const formDialogVisible = ref(false)
const isEdit = ref(false)
const editId = ref(null)
const loading = ref(false)

// 导入相关
const importDialogVisible = ref(false)
const importContent = ref('')
const importing = ref(false)

// L2TP连接状态缓存
const l2tpConnectionStatus = ref({})

const ruleForm = ref({
  Type: 'DOMAIN-SUFFIX',
  Payload: '',
  Target: 'PROXY',
  TargetType: ''
})

// 当前选中节点的备注信息
const selectedNodeRemark = ref('')

// 解析目标显示名称
const getTargetDisplayName = (row) => {
  // 直接返回目标值
  return row.target
}

// 获取节点备注（rename）
const getNodeRemark = (row) => {
  // 后端返回的是 targetType（驼峰命名）
  const targetType = row.targetType || row.target_type
  if (targetType !== 'node') {
    return ''
  }
  // 查找节点
  const node = nodes.value.find(n => n.name === row.target)
  return node?.rename || ''
}

// 获取规则拨号状态
const getRuleDialingStatus = (row) => {
  const targetType = row.targetType || row.target_type
  const target = row.target
  
  // 只有节点类型才检查拨号状态
  if (targetType !== 'node') {
    return null
  }
  
  // 从缓存中读取状态
  const connected = l2tpConnectionStatus.value[target]
  if (connected === undefined) {
    return null
  }
  
  return {
    type: connected ? 'success' : 'danger',
    text: connected ? '拨号成功' : '未连接'
  }
}

// 批量检查节点拨号状态
const checkAllNodesDialingStatus = async () => {
  // 获取所有唯一的节点名称
  const nodeNames = new Set()
  rules.value.forEach(rule => {
    const targetType = rule.targetType || rule.target_type
    if (targetType === 'node' && rule.target) {
      nodeNames.add(rule.target)
    }
  })
  
  // 批量检查每个节点的连接状态
  for (const nodeName of nodeNames) {
    try {
      const result = await checkL2TPConnection(nodeName)
      l2tpConnectionStatus.value[nodeName] = result.connected
    } catch (error) {
      console.error('Check L2TP connection error:', error)
      l2tpConnectionStatus.value[nodeName] = false
    }
  }
}

// 搜索和过滤
const searchKeyword = ref('')
const filterType = ref('')
const filterTarget = ref('')

// 分页相关
const currentPage = ref(1)
const pageSize = ref(50)
const total = ref(0)

// 计算属性：显示的规则（本地分页）
const displayRules = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return rules.value.slice(start, end)
})

// 获取规则类型标签颜色
const getTypeTagType = (type) => {
  const typeMap = {
    'DOMAIN': 'primary',
    'DOMAIN-SUFFIX': 'success',
    'DOMAIN-KEYWORD': 'warning',
    'IP-CIDR': 'info',
    'SRC-IP-CIDR': 'success',
    'GEOIP': 'primary',
    'MATCH': 'danger'
  }
  return typeMap[type] || ''
}

const loadRules = async () => {
  loading.value = true
  try {
    const params = {
      page: currentPage.value,
      pageSize: pageSize.value
    }

    // 类型过滤
    if (filterType.value) {
      params.type = filterType.value
    }

    // 关键词搜索
    if (searchKeyword.value) {
      params.keyword = searchKeyword.value
    }

    // 目标过滤 - 解析filterTarget
    if (filterTarget.value) {
      if (filterTarget.value.startsWith('node:')) {
        // 具体节点名筛选
        params.target = filterTarget.value.replace('node:', '')
      } else if (filterTarget.value.startsWith('group:')) {
        // 具体代理组名筛选
        params.target = filterTarget.value.replace('group:', '')
      } else {
        // 固定出口 (DIRECT, PROXY, REJECT)
        params.target = filterTarget.value
      }
    }

    const result = await getRules(params)
    rules.value = result.rules || []
    total.value = result.total || 0
    
    // 检查所有节点的拨号状态
    await checkAllNodesDialingStatus()
  } catch (error) {
    console.error('Load rules error:', error)
    ElMessage.error('加载规则失败')
  } finally {
    loading.value = false
  }
}

const loadNodes = async () => {
  try {
    nodes.value = await getNodes()
  } catch (error) {
    console.error('Load nodes error:', error)
    nodes.value = []
  }
}

const handleTargetChange = (value) => {
  // Only update TargetType, keep Target as-is for select display
  if (value && value.startsWith('node:')) {
    ruleForm.value.TargetType = 'node'
    // 提取节点ID，获取备注信息
    const parts = value.split(':')
    if (parts.length >= 2) {
      const nodeId = parseInt(parts[1], 10)
      const node = nodes.value.find(n => n.id === nodeId)
      selectedNodeRemark.value = node?.rename || ''
    } else {
      selectedNodeRemark.value = ''
    }
  } else if (value && value.startsWith('group:')) {
    ruleForm.value.TargetType = 'group'
    selectedNodeRemark.value = ''
  } else {
    // Built-in target (PROXY, DIRECT, REJECT)
    ruleForm.value.TargetType = 'builtin'
    selectedNodeRemark.value = ''
  }
  // Target value is kept as select option value (node:ID:Name or group:ID:Name or builtin value)
  ruleForm.value.Target = value
}

const showCreateDialog = async () => {
  isEdit.value = false
  editId.value = null
  ruleForm.value = {
    Type: 'DOMAIN-SUFFIX',
    Payload: '',
    Target: 'PROXY',
    TargetType: 'builtin'
  }
  await loadNodes()
  formDialogVisible.value = true
}

const handleEdit = async (row) => {
  isEdit.value = true
  editId.value = row.id
  // 先加载节点数据
  await loadNodes()
  // Build Target value for select dropdown
  let targetValue = row.target || 'PROXY'
  let nodeRemark = ''
  // 后端返回的是 targetType（驼峰命名）
  const targetType = row.targetType || row.target_type
  if (targetType === 'node') {
    // 后端只返回节点名称，需要通过名称查找节点
    const node = nodes.value.find(n => n.name === row.target)
    if (node) {
      targetValue = `node:${node.id}:${node.name}`
      nodeRemark = node.rename || ''
    }
  } else if (targetType === 'group') {
    // Find group by ID to get the name
  }
  ruleForm.value = {
    Type: row.type,
    Payload: row.payload,
    Target: targetValue,
    TargetType: targetType || 'builtin' // Default to builtin for backward compatibility
  }
  // 设置节点备注信息
  selectedNodeRemark.value = nodeRemark
  formDialogVisible.value = true
}

const handleSave = async () => {
  if (!ruleForm.value.Payload || !ruleForm.value.Target) {
    ElMessage.warning('请填写完整信息')
    return
  }

  // Parse Target value based on TargetType
  let targetValue = ruleForm.value.Target
  let targetID = 0
  const targetType = ruleForm.value.TargetType || 'builtin'

  if (targetType === 'node' && targetValue.startsWith('node:')) {
    // Extract ID from "node:ID:Name" format
    const parts = targetValue.split(':')
    if (parts.length >= 2) {
      targetID = parseInt(parts[1], 10)
      // Extract name for target field
      if (parts.length >= 3) {
        targetValue = parts.slice(2).join(':') // Use the node name
      } else {
        targetValue = parts[1] // Fallback to ID if name not available
      }
    }
  }
  // For builtin type, keep the original value

  // 转换为后端期望的 snake_case 格式
  const data = {
    type: ruleForm.value.Type,
    payload: ruleForm.value.Payload,
    target: targetValue,
    target_id: targetID,
    target_type: targetType
  }
  if (isEdit.value) {
    await updateRule(editId.value, data)
    ElMessage.success('更新成功')
  } else {
    await createRule(data)
    ElMessage.success('创建成功')
  }
  formDialogVisible.value = false
  // 重新加载节点和策略组数据，确保规则列表能正确显示目标名称和备注
  await loadNodes()
  await loadRules()
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定删除该规则吗？', '提示', { type: 'warning' })
  await deleteRule(row.id)
  ElMessage.success('删除成功')
  // 重新加载节点数据，确保规则列表能正确显示目标名称
  await loadNodes()
  loadRules()
}

const handleSearchChange = () => {
  currentPage.value = 1
  loadRules()
}

const handleFilterTypeChange = () => {
  currentPage.value = 1
  loadRules()
}

const handleFilterTypeClear = () => {
  currentPage.value = 1
  loadRules()
}

const handleFilterTargetChange = () => {
  currentPage.value = 1
  loadRules()
}

const handleFilterTargetClear = () => {
  currentPage.value = 1
  loadRules()
}

const resetFilter = () => {
  searchKeyword.value = ''
  filterType.value = ''
  filterTarget.value = ''
  currentPage.value = 1
  loadRules()
}

const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  loadRules()
}

const handlePageChange = (page) => {
  currentPage.value = page
  loadRules()
}

const showImportDialog = () => {
  importDialogVisible.value = true
}

const handleFileChange = (file) => {
  const reader = new FileReader()
  reader.onload = (e) => {
    importContent.value = e.target.result
  }
  reader.readAsText(file.raw)
}

const handleFileRemove = () => {
  importContent.value = ''
}

const handleImport = async () => {
  if (!importContent.value) {
    ElMessage.warning('请先选择文件')
    return
  }

  importing.value = true
  try {
    const result = await importRules(importContent.value)
    const importCount = result.import_count || 0
    const updateCount = result.update_count || 0
    let message = ''
    if (importCount > 0 && updateCount > 0) {
      message = `成功导入 ${importCount} 条规则，更新 ${updateCount} 条规则`
    } else if (importCount > 0) {
      message = `成功导入 ${importCount} 条规则`
    } else if (updateCount > 0) {
      message = `成功更新 ${updateCount} 条规则`
    } else {
      message = '没有导入任何规则'
    }
    ElMessage.success(message)
    importDialogVisible.value = false
    importContent.value = ''
    // 重新加载节点数据，确保规则列表能正确显示目标名称
    await loadNodes()
    loadRules()
  } catch (error) {
    ElMessage.error('导入失败: ' + (error.message || '未知错误'))
  } finally {
    importing.value = false
  }
}

onMounted(async () => {
  // 先加载节点数据，确保规则列表能正确显示目标名称
  await loadNodes()
  loadRules()
})
</script>

<style scoped>
.rules-page {
  height: 100%;
}

.rules-card {
  border-radius: 16px;
  border: none;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.08);
  height: 100%;
  display: flex;
  flex-direction: column;
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

.header-actions {
  display: flex;
  gap: 10px;
}

.filter-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  background: #f8f9fa;
  border-bottom: 1px solid #e9ecef;
  flex-wrap: wrap;
  gap: 12px;
}

.filter-left {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  align-items: center;
}

.filter-right {
  display: flex;
  align-items: center;
  gap: 10px;
}

.search-input {
  width: 250px;
}

.filter-select {
  width: 150px;
}

.total-count {
  color: #909399;
  font-size: 14px;
}

.rules-table {
  flex: 1;
  height: calc(100vh - 280px);
}

.priority-badge {
  display: inline-block;
  padding: 2px 8px;
  background: #f0f2f5;
  border-radius: 4px;
  font-size: 12px;
  font-weight: 500;
}

.target-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}

.target-tag {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 6px;
  font-size: 13px;
  font-weight: 500;
}

.target-tag.proxy {
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
}

.target-tag.direct {
  background: linear-gradient(135deg, #f093fb 0%, #f5576c 100%);
  color: #fff;
}

.target-tag.reject {
  background: linear-gradient(135deg, #fa709a 0%, #fee140 100%);
  color: #fff;
}

.target-tag.builtin {
  background: #e9ecef;
  color: #495057;
}

.target-tag.node {
  background: linear-gradient(135deg, #4facfe 0%, #00f2fe 100%);
  color: #fff;
}

.target-tag.group {
  background: linear-gradient(135deg, #43e97b 0%, #38f9d7 100%);
  color: #fff;
}

.text-muted {
  color: #909399;
}

.pagination-section {
  display: flex;
  justify-content: center;
  padding: 16px 0;
  border-top: 1px solid #e9ecef;
}

.rule-dialog {
  border-radius: 16px;
}

.rule-form {
  padding: 10px 0;
}

.form-input {
  width: 100%;
}

.node-remark {
  margin-top: 10px;
  padding: 5px 10px;
  background-color: #f5f7fa;
  border-radius: 4px;
  display: inline-block;
}

.form-hint {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.switch-label {
  margin-left: 8px;
  color: #606266;
}

.option-content {
  display: flex;
  align-items: center;
  gap: 8px;
}

.option-content-flex {
  display: flex;
  align-items: center;
  flex: 1;
  justify-content: space-between;
}

.option-left {
  display: flex;
  align-items: center;
  flex: 1;
  gap: 8px;
}

.option-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.option-rename {
  color: #67c23a;
  font-size: 12px;
  margin-left: 10px;
  background-color: #f0f9eb;
  padding: 2px 8px;
  border-radius: 4px;
}

.option-type-tag {
  flex-shrink: 0;
  margin-left: 10px;
}

.option-text {
  color: #909399;
  font-size: 12px;
}

.option-desc {
  color: #909399;
  font-size: 12px;
}

.import-dialog {
  border-radius: 16px;
}

.import-form {
  padding: 10px 0;
}

.import-textarea {
  font-family: 'Courier New', monospace;
  font-size: 12px;
}

.upload-icon {
  font-size: 48px;
  color: #409eff;
  margin-bottom: 16px;
}

.upload-text {
  font-size: 14px;
  color: #606266;
}

.upload-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 8px;
}

:deep(.el-card__body) {
  padding: 0;
  display: flex;
  flex-direction: column;
}

:deep(.el-table) {
  border: none;
}

:deep(.el-table th) {
  background: #fafafa;
  color: #606266;
  font-weight: 500;
}

:deep(.el-dialog__header) {
  padding: 20px 24px;
  border-bottom: 1px solid #e9ecef;
}

:deep(.el-dialog__body) {
  padding: 24px;
}

:deep(.el-dialog__footer) {
  padding: 16px 24px;
  border-top: 1px solid #e9ecef;
}

:deep(.el-upload-dragger) {
  padding: 40px 20px;
  border: 2px dashed #d9d9d9;
  border-radius: 8px;
  background: #fafafa;
}

:deep(.el-upload-dragger:hover) {
  border-color: #409eff;
  background: #f0f7ff;
}
</style>