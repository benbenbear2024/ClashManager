<template>
  <div class="nodes-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <div class="header-left">
            <el-icon size="20"><svg viewBox="0 0 1024 1024" width="20" height="20"><path fill="currentColor" d="M128 192h768v128H192v640h640v-64h64v128H128V192z"></path><path fill="currentColor" d="M384 384h384v64H384v320h320v-64h64v128H384V384z"></path></svg></el-icon>
            <span>节点列表</span>
          </div>
          <div class="header-right">
            <el-button type="danger" :icon="Delete" @click="handleBatchDelete" :disabled="selectedNodes.length === 0">批量删除</el-button>
            <el-button type="primary" :icon="Upload" @click="showImportDialog">导入节点</el-button>
            <el-button type="success" :icon="Plus" @click="showCreateDialog">新增节点</el-button>
          </div>
        </div>
      </template>

      <el-table :data="pagedNodes" stripe style="width: 100%" height="calc(100vh - 200px)" :stripe="true" :border="false" @selection-change="handleSelectionChange">
        <el-table-column type="selection" width="35" />
        <el-table-column prop="id" label="ID" min-width="40" />
        <el-table-column prop="name" label="名称" min-width="60" />
        <el-table-column label="备注" min-width="120">
          <template #default="{ row }">
            <el-tag size="small" type="info" v-if="row.rename">{{ row.rename }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" min-width="80">
          <template #default="{ row }">
            <el-tag>{{ getTypeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="server" label="服务器" min-width="100" show-overflow-tooltip />
        <el-table-column prop="port" label="端口" min-width="70" />
        <el-table-column prop="network" label="传输" min-width="80" />
        <el-table-column label="来源" min-width="120">
          <template #default="{ row }">
            <el-tag v-if="row.source === 'subscription'" type="warning" size="small">
              {{ row.sourceName || '订阅' }}
            </el-tag>
            <el-tag v-else type="info" size="small">手动</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="状态" min-width="70">
          <template #default="{ row }">
            <el-tag :type="row.tls ? 'success' : 'info'" size="small">
              {{ row.tls ? 'TLS' : '普通' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" min-width="120" fixed="right">
          <template #default="{ row }">
            <el-button type="success" link @click="handleExport(row)">导出</el-button>
            <el-button type="primary" link @click="handleEdit(row)">编辑</el-button>
            <el-button type="danger" link @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div class="pagination-container">
        <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[50, 100, 200]"
          layout="total, prev, pager, next, sizes"
          :total="nodes.length"
          @size-change="handleSizeChange"
          @current-change="handleCurrentChange"
        />
        <div class="jump-container">
          <span>跳转到</span>
          <el-input
            v-model="jumpPage"
            class="jump-input"
            @keyup.enter="handleJump"
          />
          <span>页</span>
          <el-button
            type="primary"
            :disabled="!canJump"
            @click="handleJump"
          >
            确定
          </el-button>
        </div>
      </div>
    </el-card>

    <!-- 导入节点对话框 -->
    <el-dialog v-model="importDialogVisible" title="导入节点" width="800px">
      <el-form :model="importForm">
        <el-form-item label="分享链接">
          <el-input
            v-model="importForm.link"
            type="textarea"
            :rows="15"
            placeholder="请粘贴节点分享链接 (ss://, vmess://, trojan://, vless://, socks5://, hysteria2:// 等)，每个节点一行。无协议头默认视为 SOCKS5 节点 (格式：服务器 | 端口 | 用户名 | 密码 | 名称)"
            style="font-family: monospace; font-size: 13px;"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="importDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleImport">导入</el-button>
      </template>
    </el-dialog>

    <!-- 新增/编辑节点对话框 -->
    <el-dialog v-model="formDialogVisible" :title="isEdit ? '编辑节点' : '新增节点'" width="700px">
      <el-tabs v-model="activeTab" type="border-card" @tab-change="handleTabChange">
        <!-- Shadowsocks -->
        <el-tab-pane label="Shadowsocks" name="ss">
          <template #label>
            <span style="display: flex; align-items: center; gap: 5px;">
              <el-icon><svg viewBox="0 0 1024 1024" width="16" height="16"><path fill="currentColor" d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64z m0 820c-205.4 0-372-166.6-372-372s166.6-372 372-372 372 166.6 372 372-166.6 372-372 372z"></path></svg></el-icon>
              Shadowsocks
            </span>
          </template>
          <el-form :model="nodeForm" label-width="110px" class="node-form">
            <el-form-item label="节点名称">
              <el-input v-model="nodeForm.Name" placeholder="请输入节点名称" :disabled="isNameAutoGenerated" />
              <span v-if="isNameAutoGenerated" class="auto-name-tip" style="color: #909399; font-size: 12px; margin-top: 5px; display: block;">自动生成的节点名称不可修改</span>
            </el-form-item>
            <el-form-item label="备注">
              <el-input v-model="nodeForm.Rename" placeholder="请输入备注信息" />
              <span style="color: #909399; font-size: 12px; margin-top: 5px; display: block;">用于存储原始节点名称或自定义备注</span>
            </el-form-item>
            <el-form-item label="服务器地址">
              <el-input v-model="nodeForm.Server" placeholder="请输入服务器地址" />
            </el-form-item>
            <el-form-item label="端口">
              <el-input-number v-model="nodeForm.Port" :min="1" :max="65535" style="width: 100%" />
            </el-form-item>
            <el-form-item label="加密方式">
              <el-select v-model="nodeForm.Cipher" placeholder="请选择加密方式" style="width: 100%">
                <el-option label="aes-128-gcm" value="aes-128-gcm" />
                <el-option label="aes-192-gcm" value="aes-192-gcm" />
                <el-option label="aes-256-gcm" value="aes-256-gcm" />
                <el-option label="aes-128-cfb" value="aes-128-cfb" />
                <el-option label="aes-192-cfb" value="aes-192-cfb" />
                <el-option label="aes-256-cfb" value="aes-256-cfb" />
                <el-option label="aes-128-ctr" value="aes-128-ctr" />
                <el-option label="aes-192-ctr" value="aes-192-ctr" />
                <el-option label="aes-256-ctr" value="aes-256-ctr" />
                <el-option label="rc4-md5" value="rc4-md5" />
                <el-option label="chacha20-ietf" value="chacha20-ietf" />
                <el-option label="xchacha20" value="xchacha20" />
                <el-option label="chacha20-ietf-poly1305" value="chacha20-ietf-poly1305" />
                <el-option label="xchacha20-ietf-poly1305" value="xchacha20-ietf-poly1305" />
              </el-select>
            </el-form-item>
            <el-form-item label="密码">
              <el-input v-model="nodeForm.Password" placeholder="请输入密码" show-password />
            </el-form-item>
            <el-form-item label="UDP转发">
              <el-switch v-model="nodeForm.UDP" />
              <span style="margin-left: 10px; color: #909399; font-size: 12px;">启用UDP转发</span>
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- VMess -->
        <el-tab-pane label="VMess" name="vmess">
          <template #label>
            <span style="display: flex; align-items: center; gap: 5px;">
              <el-icon><svg viewBox="0 0 1024 1024" width="16" height="16"><path fill="currentColor" d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64z m0 820c-205.4 0-372-166.6-372-372s166.6-372 372-372 372 166.6 372 372-166.6 372-372 372z"></path></svg></el-icon>
              VMess
            </span>
          </template>
          <el-form :model="nodeForm" label-width="110px" class="node-form">
            <el-form-item label="节点名称">
              <el-input v-model="nodeForm.Name" placeholder="请输入节点名称" :disabled="isNameAutoGenerated" />
              <span v-if="isNameAutoGenerated" class="auto-name-tip" style="color: #909399; font-size: 12px; margin-top: 5px; display: block;">自动生成的节点名称不可修改</span>
            </el-form-item>
            <el-form-item label="备注">
              <el-input v-model="nodeForm.Rename" placeholder="请输入备注信息" />
              <span style="color: #909399; font-size: 12px; margin-top: 5px; display: block;">用于存储原始节点名称或自定义备注</span>
            </el-form-item>
            <el-form-item label="服务器地址">
              <el-input v-model="nodeForm.Server" placeholder="请输入服务器地址" />
            </el-form-item>
            <el-form-item label="端口">
              <el-input-number v-model="nodeForm.Port" :min="1" :max="65535" style="width: 100%" />
            </el-form-item>
            <el-form-item label="UUID">
              <el-input v-model="nodeForm.UUID" placeholder="请输入UUID" />
            </el-form-item>
            <el-form-item label="加密方式">
              <el-select v-model="nodeForm.Cipher" placeholder="请选择加密方式" style="width: 100%">
                <el-option label="auto" value="auto" />
                <el-option label="aes-128-gcm" value="aes-128-gcm" />
                <el-option label="chacha20-poly1305" value="chacha20-poly1305" />
                <el-option label="none" value="none" />
              </el-select>
            </el-form-item>
            <el-form-item label="传输协议">
              <el-select v-model="nodeForm.Network" placeholder="请选择传输协议" style="width: 100%">
                <el-option label="TCP" value="" />
                <el-option label="WebSocket" value="ws" />
                <el-option label="gRPC" value="grpc" />
              </el-select>
            </el-form-item>
            <el-form-item label="路径" v-if="nodeForm.Network === 'ws' || nodeForm.Network === 'grpc'">
              <el-input v-model="nodeForm.Path" :placeholder="nodeForm.Network === 'ws' ? '请输入WebSocket路径，如: /ws' : '请输入gRPC服务名'" />
            </el-form-item>
            <el-form-item label="Host/SNI" v-if="nodeForm.Network === 'ws' || nodeForm.Network === 'grpc' || nodeForm.TLS">
              <el-input v-model="nodeForm.Host" placeholder="请输入Host或SNI" />
            </el-form-item>
            <el-form-item label="TLS加密">
              <el-switch v-model="nodeForm.TLS" />
            </el-form-item>
            <el-form-item label="跳过证书验证" v-if="nodeForm.TLS">
              <el-switch v-model="nodeForm.SkipCert" />
            </el-form-item>
            <el-form-item label="UDP转发">
              <el-switch v-model="nodeForm.UDP" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- VLESS -->
        <el-tab-pane label="VLESS" name="vless">
          <template #label>
            <span style="display: flex; align-items: center; gap: 5px;">
              <el-icon><svg viewBox="0 0 1024 1024" width="16" height="16"><path fill="currentColor" d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64z m0 820c-205.4 0-372-166.6-372-372s166.6-372 372-372 372 166.6 372 372-166.6 372-372 372z"></path></svg></el-icon>
              VLESS
            </span>
          </template>
          <el-form :model="nodeForm" label-width="110px" class="node-form">
            <el-form-item label="节点名称">
              <el-input v-model="nodeForm.Name" placeholder="请输入节点名称" :disabled="isNameAutoGenerated" />
              <span v-if="isNameAutoGenerated" class="auto-name-tip" style="color: #909399; font-size: 12px; margin-top: 5px; display: block;">自动生成的节点名称不可修改</span>
            </el-form-item>
            <el-form-item label="备注">
              <el-input v-model="nodeForm.Rename" placeholder="请输入备注信息" />
              <span style="color: #909399; font-size: 12px; margin-top: 5px; display: block;">用于存储原始节点名称或自定义备注</span>
            </el-form-item>
            <el-form-item label="服务器地址">
              <el-input v-model="nodeForm.Server" placeholder="请输入服务器地址" />
            </el-form-item>
            <el-form-item label="端口">
              <el-input-number v-model="nodeForm.Port" :min="1" :max="65535" style="width: 100%" />
            </el-form-item>
            <el-form-item label="UUID">
              <el-input v-model="nodeForm.UUID" placeholder="请输入UUID" />
            </el-form-item>
            <el-form-item label="传输协议">
              <el-select v-model="nodeForm.Network" placeholder="请选择传输协议" style="width: 100%">
                <el-option label="TCP" value="" />
                <el-option label="WebSocket" value="ws" />
                <el-option label="gRPC" value="grpc" />
              </el-select>
            </el-form-item>
            <el-form-item label="路径" v-if="nodeForm.Network === 'ws' || nodeForm.Network === 'grpc'">
              <el-input v-model="nodeForm.Path" :placeholder="nodeForm.Network === 'ws' ? '请输入WebSocket路径，如: /ws' : '请输入gRPC服务名'" />
            </el-form-item>
            <el-form-item label="Host/SNI" v-if="nodeForm.Network === 'ws' || nodeForm.Network === 'grpc' || nodeForm.TLS">
              <el-input v-model="nodeForm.Host" placeholder="请输入Host或SNI" />
            </el-form-item>
            <el-form-item label="TLS加密">
              <el-switch v-model="nodeForm.TLS" />
            </el-form-item>
            <el-form-item label="跳过证书验证" v-if="nodeForm.TLS">
              <el-switch v-model="nodeForm.SkipCert" />
            </el-form-item>
            <el-form-item label="UDP转发">
              <el-switch v-model="nodeForm.UDP" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- Trojan -->
        <el-tab-pane label="Trojan" name="trojan">
          <template #label>
            <span style="display: flex; align-items: center; gap: 5px;">
              <el-icon><svg viewBox="0 0 1024 1024" width="16" height="16"><path fill="currentColor" d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64z m0 820c-205.4 0-372-166.6-372-372s166.6-372 372-372 372 166.6 372 372-166.6 372-372 372z"></path></svg></el-icon>
              Trojan
            </span>
          </template>
          <el-form :model="nodeForm" label-width="110px" class="node-form">
            <el-form-item label="节点名称">
              <el-input v-model="nodeForm.Name" placeholder="请输入节点名称" :disabled="isNameAutoGenerated" />
              <span v-if="isNameAutoGenerated" class="auto-name-tip" style="color: #909399; font-size: 12px; margin-top: 5px; display: block;">自动生成的节点名称不可修改</span>
            </el-form-item>
            <el-form-item label="备注">
              <el-input v-model="nodeForm.Rename" placeholder="请输入备注信息" />
              <span style="color: #909399; font-size: 12px; margin-top: 5px; display: block;">用于存储原始节点名称或自定义备注</span>
            </el-form-item>
            <el-form-item label="服务器地址">
              <el-input v-model="nodeForm.Server" placeholder="请输入服务器地址" />
            </el-form-item>
            <el-form-item label="端口">
              <el-input-number v-model="nodeForm.Port" :min="1" :max="65535" style="width: 100%" />
            </el-form-item>
            <el-form-item label="密码">
              <el-input v-model="nodeForm.Password" placeholder="请输入密码" show-password />
            </el-form-item>
            <el-form-item label="传输协议">
              <el-select v-model="nodeForm.Network" placeholder="请选择传输协议" style="width: 100%">
                <el-option label="TCP" value="" />
                <el-option label="WebSocket" value="ws" />
                <el-option label="gRPC" value="grpc" />
              </el-select>
            </el-form-item>
            <el-form-item label="路径" v-if="nodeForm.Network === 'ws' || nodeForm.Network === 'grpc'">
              <el-input v-model="nodeForm.Path" :placeholder="nodeForm.Network === 'ws' ? '请输入WebSocket路径，如: /ws' : '请输入gRPC服务名'" />
            </el-form-item>
            <el-form-item label="Host/SNI" v-if="nodeForm.Network === 'ws' || nodeForm.Network === 'grpc'">
              <el-input v-model="nodeForm.Host" placeholder="请输入Host或SNI" />
            </el-form-item>
            <el-form-item label="TLS加密">
              <el-switch v-model="nodeForm.TLS" />
              <span style="margin-left: 10px; color: #909399; font-size: 12px;">Trojan必须启用TLS</span>
            </el-form-item>
            <el-form-item label="跳过证书验证" v-if="nodeForm.TLS">
              <el-switch v-model="nodeForm.SkipCert" />
            </el-form-item>
            <el-form-item label="UDP转发">
              <el-switch v-model="nodeForm.UDP" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- Hysteria2 -->
        <el-tab-pane label="Hysteria2" name="hysteria2">
          <template #label>
            <span style="display: flex; align-items: center; gap: 5px;">
              <el-icon><svg viewBox="0 0 1024 1024" width="16" height="16"><path fill="currentColor" d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64z m0 820c-205.4 0-372-166.6-372-372s166.6-372 372-372 372 166.6 372 372-166.6 372-372 372z"></path></svg></el-icon>
              Hysteria2
            </span>
          </template>
          <el-form :model="nodeForm" label-width="110px" class="node-form">
            <el-form-item label="节点名称">
              <el-input v-model="nodeForm.Name" placeholder="请输入节点名称" :disabled="isNameAutoGenerated" />
              <span v-if="isNameAutoGenerated" class="auto-name-tip" style="color: #909399; font-size: 12px; margin-top: 5px; display: block;">自动生成的节点名称不可修改</span>
            </el-form-item>
            <el-form-item label="备注">
              <el-input v-model="nodeForm.Rename" placeholder="请输入备注信息" />
              <span style="color: #909399; font-size: 12px; margin-top: 5px; display: block;">用于存储原始节点名称或自定义备注</span>
            </el-form-item>
            <el-form-item label="服务器地址">
              <el-input v-model="nodeForm.Server" placeholder="请输入服务器地址" />
            </el-form-item>
            <el-form-item label="端口">
              <el-input-number v-model="nodeForm.Port" :min="1" :max="65535" style="width: 100%" />
            </el-form-item>
            <el-form-item label="密码">
              <el-input v-model="nodeForm.Password" placeholder="请输入密码" show-password />
            </el-form-item>
            <el-form-item label="SNI">
              <el-input v-model="nodeForm.Host" placeholder="请输入SNI" />
            </el-form-item>
            <el-form-item label="TLS加密">
              <el-switch v-model="nodeForm.TLS" />
              <span style="margin-left: 10px; color: #909399; font-size: 12px;">Hysteria2必须启用TLS</span>
            </el-form-item>
            <el-form-item label="跳过证书验证" v-if="nodeForm.TLS">
              <el-switch v-model="nodeForm.SkipCert" />
            </el-form-item>
            <el-form-item label="UDP转发">
              <el-switch v-model="nodeForm.UDP" />
            </el-form-item>
          </el-form>
        </el-tab-pane>

        <!-- SOCKS5 -->
        <el-tab-pane label="SOCKS5" name="socks5">
          <template #label>
            <span style="display: flex; align-items: center; gap: 5px;">
              <el-icon><svg viewBox="0 0 1024 1024" width="16" height="16"><path fill="currentColor" d="M512 64C264.6 64 64 264.6 64 512s200.6 448 448 448 448-200.6 448-448S759.4 64 512 64z m0 820c-205.4 0-372-166.6-372-372s166.6-372 372-372 372 166.6 372 372-166.6 372-372 372z"></path></svg></el-icon>
              SOCKS5
            </span>
          </template>
          <el-form :model="nodeForm" label-width="110px" class="node-form">
            <el-form-item label="节点名称">
              <el-input v-model="nodeForm.Name" placeholder="请输入节点名称" :disabled="isNameAutoGenerated" />
              <span v-if="isNameAutoGenerated" class="auto-name-tip" style="color: #909399; font-size: 12px; margin-top: 5px; display: block;">自动生成的节点名称不可修改</span>
            </el-form-item>
            <el-form-item label="备注">
              <el-input v-model="nodeForm.Rename" placeholder="请输入备注信息" />
              <span style="color: #909399; font-size: 12px; margin-top: 5px; display: block;">用于存储原始节点名称或自定义备注</span>
            </el-form-item>
            <el-form-item label="服务器地址">
              <el-input v-model="nodeForm.Server" placeholder="请输入服务器地址" />
            </el-form-item>
            <el-form-item label="端口">
              <el-input-number v-model="nodeForm.Port" :min="1" :max="65535" style="width: 100%" />
            </el-form-item>
            <el-form-item label="用户名">
              <el-input v-model="nodeForm.Username" placeholder="请输入用户名" />
            </el-form-item>
            <el-form-item label="密码">
              <el-input v-model="nodeForm.Password" placeholder="请输入密码" show-password />
            </el-form-item>
            <el-form-item label="传输协议">
              <el-select v-model="nodeForm.Network" placeholder="请选择传输协议" style="width: 100%">
                <el-option label="TCP" value="" />
                <el-option label="WebSocket" value="ws" />
                <el-option label="gRPC" value="grpc" />
              </el-select>
            </el-form-item>
            <el-form-item label="路径" v-if="nodeForm.Network === 'ws' || nodeForm.Network === 'grpc'">
              <el-input v-model="nodeForm.Path" :placeholder="nodeForm.Network === 'ws' ? '请输入WebSocket路径，如: /ws' : '请输入gRPC服务名'" />
            </el-form-item>
            <el-form-item label="Host/SNI" v-if="nodeForm.Network === 'ws' || nodeForm.Network === 'grpc' || nodeForm.TLS">
              <el-input v-model="nodeForm.Host" placeholder="请输入Host或SNI" />
            </el-form-item>
            <el-form-item label="TLS加密">
              <el-switch v-model="nodeForm.TLS" />
            </el-form-item>
            <el-form-item label="跳过证书验证" v-if="nodeForm.TLS">
              <el-switch v-model="nodeForm.SkipCert" />
            </el-form-item>
            <el-form-item label="UDP转发">
              <el-switch v-model="nodeForm.UDP" />
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
      <template #footer>
        <el-button @click="formDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSave">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus, Upload, Delete, ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import { getNodes, createNode, updateNode, deleteNode, importNode, exportNode } from '@/api/nodes'

const nodes = ref([])
const selectedNodes = ref([])
const importDialogVisible = ref(false)
const formDialogVisible = ref(false)
const activeTab = ref('vmess')
const isEdit = ref(false)
const importForm = ref({ link: '' })

// 分页相关
const currentPage = ref(1)
const pageSize = ref(50)
const pagedNodes = ref([])
const jumpPage = ref(1)

// 计算总页数
const totalPages = computed(() => {
  return Math.ceil(nodes.value.length / pageSize.value) || 1
})

// 判断是否可以跳转
const canJump = computed(() => {
  return jumpPage.value >= 1 && jumpPage.value <= totalPages.value && jumpPage.value !== currentPage.value
})

// 判断节点名称是否为自动生成的Name*格式
const isNameAutoGenerated = computed(() => {
  return /^Name\d+$/.test(nodeForm.value.Name)
})

const nodeForm = ref({
  Name: '',
  Type: 'vmess',
  Server: '',
  Port: 443,
  UUID: '',
  Username: '',
  Password: '',
  Cipher: '',
  Network: '',
  Path: '',
  Host: '',
  TLS: true,
  SkipCert: false,
  UDP: true
})

const getTypeLabel = (type) => {
  const labels = {
    ss: 'Shadowsocks',
    vmess: 'VMess',
    vless: 'VLESS',
    trojan: 'Trojan',
    hysteria2: 'Hysteria2',
    socks5: 'SOCKS5'
  }
  return labels[type] || type
}

const loadNodes = async () => {
  nodes.value = await getNodes()
  updatePagedNodes()
}

// 更新分页数据
const updatePagedNodes = () => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  pagedNodes.value = nodes.value.slice(start, end)
}

// 获取页码列表
const getPageList = () => {
  const total = totalPages.value
  const current = currentPage.value
  const pages = []
  
  if (total <= 7) {
    for (let i = 1; i <= total; i++) {
      pages.push(i)
    }
  } else {
    if (current <= 4) {
      for (let i = 1; i <= 5; i++) {
        pages.push(i)
      }
      pages.push('...')
      pages.push(total)
    } else if (current >= total - 3) {
      pages.push(1)
      pages.push('...')
      for (let i = total - 4; i <= total; i++) {
        pages.push(i)
      }
    } else {
      pages.push(1)
      pages.push('...')
      for (let i = current - 1; i <= current + 1; i++) {
        pages.push(i)
      }
      pages.push('...')
      pages.push(total)
    }
  }
  return pages
}

// 处理跳转
const handleJump = () => {
  if (canJump.value) {
    handleCurrentChange(jumpPage.value)
  }
}

// 处理每页大小变化
const handleSizeChange = (size) => {
  pageSize.value = size
  currentPage.value = 1
  updatePagedNodes()
}

// 处理页码变化
const handleCurrentChange = (current) => {
  currentPage.value = current
  updatePagedNodes()
}

const handleSelectionChange = (selection) => {
  selectedNodes.value = selection.map(node => node.id)
}

const handleBatchDelete = async () => {
  if (selectedNodes.value.length === 0) {
    ElMessage.warning('请选择要删除的节点')
    return
  }
  
  try {
    await ElMessageBox.confirm(
      `确定要删除选中的 ${selectedNodes.value.length} 个节点吗？`,
      '批量删除',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    // 批量删除节点
    const deletePromises = selectedNodes.value.map(id => deleteNode(id))
    await Promise.all(deletePromises)
    
    ElMessage.success(`成功删除 ${selectedNodes.value.length} 个节点`)
    selectedNodes.value = []
    loadNodes()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败')
      console.error('批量删除失败:', error)
    }
  }
}

const showImportDialog = () => {
  importForm.value.link = ''
  importDialogVisible.value = true
}

const handleImport = async () => {
  if (!importForm.value.link.trim()) {
    ElMessage.warning('请输入节点链接')
    return
  }
  
  // 按行分割链接，处理不同的换行符
  const links = importForm.value.link.trim().split(/[\r\n]+/)
  let successCount = 0
  let errorCount = 0
  
  console.log('开始导入节点，共', links.length, '个链接')
  
  for (let i = 0; i < links.length; i++) {
    const link = links[i]
    const trimmedLink = link.trim()
    if (trimmedLink) {
      // 如果没有协议头，默认添加 sk5:// 前缀
      let processLink = trimmedLink
      const hasProtocol = /^(ss|vmess|trojan|vless|socks5|hysteria2|hysteria):\/\//i.test(trimmedLink)
      if (!hasProtocol) {
        processLink = 'socks5://' + trimmedLink
        console.log('检测到无协议头的节点，自动添加 socks5:// 前缀:', processLink)
      }
      
      console.log('正在导入第', i + 1, '个节点:', processLink)
      try {
        await importNode(processLink)
        console.log('第', i + 1, '个节点导入成功')
        successCount++
      } catch (error) {
        console.error('第', i + 1, '个节点导入失败:', processLink, error)
        errorCount++
      }
    }
  }
  
  console.log('导入完成，成功:', successCount, '失败:', errorCount)
  
  if (successCount > 0) {
    ElMessage.success(`成功导入 ${successCount} 个节点`)
  }
  if (errorCount > 0) {
    ElMessage.warning(`跳过了 ${errorCount} 个无效节点`)
  }
  
  importDialogVisible.value = false
  loadNodes()
}

const showCreateDialog = () => {
  isEdit.value = false
  activeTab.value = 'vmess'
  resetForm()
  formDialogVisible.value = true
}

const resetForm = () => {
  nodeForm.value = {
    Name: '',
    Type: activeTab.value,
    Server: '',
    Port: 443,
    UUID: '',
    Username: '',
    Password: '',
    Cipher: activeTab.value === 'ss' ? 'aes-256-gcm' : 'auto',
    Network: '',
    Path: '',
    Host: '',
    TLS: true,
    SkipCert: false,
    UDP: true,
    Rename: ''
  }
}

const handleTabChange = (tabName) => {
  nodeForm.value.Type = tabName
  // 切换tab时重置表单，保留已填写的名称和服务器
  const keepFields = { Name: nodeForm.value.Name, Server: nodeForm.value.Server, Port: nodeForm.value.Port, Rename: nodeForm.value.Rename }
  if (isEdit.value) {
    keepFields.ID = nodeForm.value.ID
  }
  resetForm()
  Object.assign(nodeForm.value, keepFields)
}

const handleEdit = (row) => {
  isEdit.value = true
  activeTab.value = row.type
  // 将后端返回的蛇形命名字段转换为表单的驼峰命名
  nodeForm.value = {
    ID: row.id,
    Name: row.name,
    Type: row.type,
    Server: row.server,
    Port: row.port,
    UUID: row.uuid || '',
    Username: row.username || '',
    Password: row.password || '',
    Cipher: row.cipher || '',
    Network: row.network || '',
    Path: row.path || '',
    Host: row.host || '',
    TLS: row.tls || false,
    SkipCert: row.skipCert || false,
    UDP: row.udp || false,
    ALPN: row.alpn || '',
    ExtraConfig: row.extraConfig || '',
    Rename: row.rename || ''
  }
  formDialogVisible.value = true
}

const handleSave = async () => {
  if (!nodeForm.value.Name || !nodeForm.value.Server) {
    ElMessage.warning('请填写节点名称和服务器地址')
    return
  }

  // 根据类型验证必填字段
  const type = nodeForm.value.Type
  if (type === 'vmess' || type === 'vless') {
    if (!nodeForm.value.UUID) {
      ElMessage.warning('请填写UUID')
      return
    }
  } else if (type === 'ss' || type === 'trojan' || type === 'hysteria2') {
    if (!nodeForm.value.Password) {
      ElMessage.warning('请填写密码')
      return
    }
  }
  // SOCKS5节点的用户名和密码是可选的，支持匿名SOCKS5

  // 转换为后端期望的 snake_case 格式
  const data = {
    name: nodeForm.value.Name,
    type: nodeForm.value.Type,
    server: nodeForm.value.Server,
    port: nodeForm.value.Port,
    uuid: nodeForm.value.UUID || '',
    username: nodeForm.value.Username || '',
    password: nodeForm.value.Password || '',
    cipher: nodeForm.value.Cipher || '',
    network: nodeForm.value.Network || '',
    path: nodeForm.value.Path || '',
    host: nodeForm.value.Host || '',
    tls: nodeForm.value.TLS,
    skip_cert: nodeForm.value.SkipCert,
    udp: nodeForm.value.UDP,
    alpn: nodeForm.value.ALPN || '',
    extra_config: nodeForm.value.ExtraConfig || '',
    rename: nodeForm.value.Rename || ''
  }

  if (isEdit.value) {
    console.log('handleSave: Editing node, ID=', nodeForm.value.ID)
    await updateNode(nodeForm.value.ID, data)
    ElMessage.success('更新成功')
  } else {
    console.log('handleSave: Creating new node')
    await createNode(data)
    ElMessage.success('创建成功')
  }
  formDialogVisible.value = false
  loadNodes()
}

const handleDelete = async (row) => {
  await ElMessageBox.confirm('确定删除该节点吗？', '提示', { type: 'warning' })
  await deleteNode(row.id)
  ElMessage.success('删除成功')
  loadNodes()
}

const handleExport = async (row) => {
  try {
    const result = await exportNode(row.id)
    const link = result.link

    // 复制到剪贴板
    await navigator.clipboard.writeText(link)
    ElMessage.success('节点分享链接已复制到剪贴板')

    // 可选：同时显示链接让用户确认
    console.log('导出的节点链接:', link)
  } catch (error) {
    // 如果剪贴板复制失败，显示链接让用户手动复制
    try {
      const result = await exportNode(row.id)
      const link = result.link

      ElMessageBox.alert(
        link,
        '节点分享链接',
        {
          confirmButtonText: '确定',
          type: 'success',
          inputType: 'textarea',
          showInput: false,
          dangerouslyUseHTMLString: false,
          message: link,
          customClass: 'export-link-dialog'
        }
      ).catch(() => {})
    } catch (err) {
      console.error('导出失败:', err)
      ElMessage.error('导出失败：' + (err.response?.data?.error || err.message))
    }
  }
}

onMounted(() => {
  loadNodes()
})
</script>

<style scoped>
.nodes-page {
  height: 100%;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.header-right {
  display: flex;
  gap: 10px;
}

:deep(.el-card__header) {
  padding: 16px 20px;
  border-bottom: 1px solid #f0f0f0;
}

:deep(.el-card__body) {
  padding: 0;
}

.pagination-container {
  padding: 20px;
  display: flex;
  justify-content: center;
  align-items: center;
  flex-wrap: wrap;
  gap: 15px;
  background: #fff;
  border-top: 1px solid #f0f0f0;
}

.pagination-container :deep(.el-pagination) {
  --el-pagination-font-size: 13px;
}

.pagination-container :deep(.el-pagination .el-input__inner) {
  font-size: 13px;
}

.jump-container {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
}

.jump-input {
  width: 50px;
}

.jump-input :deep(.el-input__wrapper) {
  padding: 2px 8px;
}

:deep(.el-table) {
  border: none;
}

:deep(.el-table__header-wrapper) {
  background: #fafafa;
}

:deep(.el-table th) {
  background: #fafafa;
  color: #606266;
  font-weight: 500;
}

.node-form {
  padding: 20px;
  max-height: 500px;
  overflow-y: auto;
}

:deep(.el-tabs__header) {
  margin: 0;
}

:deep(.el-tabs__content) {
  padding: 0;
}

:deep(.el-tabs__item) {
  padding: 0 10px;
}

:deep(.el-tab-pane) {
  background: #fff;
}

/* 导出链接对话框样式 */
:deep(.export-link-dialog .el-message-box__content) {
  padding: 10px 20px;
}

:deep(.export-link-dialog .el-message-box__message) {
  word-break: break-all;
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 12px;
  line-height: 1.5;
  max-height: 200px;
  overflow-y: auto;
}
</style>
