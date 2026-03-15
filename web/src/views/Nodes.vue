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
            <el-button type="warning" :icon="Edit" @click="showBatchEditDialog" :disabled="selectedNodes.length === 0">批量编辑</el-button>
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

    <!-- 批量编辑节点对话框 -->
    <el-dialog v-model="batchEditDialogVisible" title="批量编辑节点" width="800px">
      <div style="margin-bottom: 15px;">
        <el-tag type="info" size="large">已选择 {{ selectedNodes.length }} 条节点</el-tag>
        <el-tag v-if="batchEditInputCount > 0" type="success" size="large" style="margin-left: 10px;">将导入 {{ batchEditInputCount }} 条</el-tag>
      </div>
      <el-form :model="batchEditForm">
        <el-form-item label="节点信息">
          <el-input
            v-model="batchEditForm.nodesText"
            type="textarea"
            :rows="15"
            placeholder="请粘贴节点信息，每行一个节点，格式与导入节点相同。节点数量应与选择的节点数量一致。"
            style="font-family: monospace; font-size: 13px;"
            @input="handleBatchEditInput"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="batchEditDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleBatchEditSave">保存</el-button>
      </template>
    </el-dialog>

    <!-- 新增/编辑节点对话框 -->
    <el-dialog v-model="formDialogVisible" :title="isEdit ? '编辑节点' : '新增节点'" width="700px" class="node-edit-dialog">
      <el-tabs v-model="activeTab" type="border-card" @tab-change="handleTabChange" class="node-type-tabs">
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
            <el-form-item label="地址">
              <el-input v-model="nodeForm.Address" placeholder="请输入地址（可选）" />
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
            <el-form-item label="AlterId">
              <el-input v-model="nodeForm.AlterId" placeholder="请输入AlterId，默认为0" />
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
            <el-form-item label="上行带宽(Mbps)">
              <el-input-number v-model="nodeForm.Up" :min="1" :max="1000" style="width: 100%" />
            </el-form-item>
            <el-form-item label="下行带宽(Mbps)">
              <el-input-number v-model="nodeForm.Down" :min="1" :max="1000" style="width: 100%" />
            </el-form-item>
            <el-form-item label="跳跃间隔(秒)">
              <el-input-number v-model="nodeForm.HopInterval" :min="1" :max="3600" style="width: 100%" />
            </el-form-item>
            <el-form-item label="流控(Flow)">
              <el-input v-model="nodeForm.Flow" placeholder="留空或输入如 xtls-rprx-vision" />
              <span style="color: #909399; font-size: 12px; margin-top: 5px; display: block;">Reality 模式通常留空</span>
            </el-form-item>
            <el-form-item label="服务器名称" v-if="nodeForm.TLS">
              <el-input v-model="nodeForm.ServerName" placeholder="Reality 模式下填写服务器名称" />
            </el-form-item>
            <el-form-item label="Reality 公钥" v-if="nodeForm.TLS">
              <el-input v-model="nodeForm.PublicKey" placeholder="Reality 模式下填写公钥" />
            </el-form-item>
            <el-form-item label="Reality 短ID" v-if="nodeForm.TLS">
              <el-input v-model="nodeForm.ShortID" placeholder="Reality 模式下填写短ID" />
            </el-form-item>
            <el-form-item label="Reality 指纹" v-if="nodeForm.TLS">
              <el-select v-model="nodeForm.Fingerprint" placeholder="请选择 Reality 指纹" style="width: 100%">
                <el-option label="Chrome" value="chrome" />
                <el-option label="Firefox" value="firefox" />
                <el-option label="Safari" value="safari" />
                <el-option label="Edge" value="edge" />
                <el-option label="Random" value="random" />
                <el-option label="Randomized" value="randomized" />
              </el-select>
            </el-form-item>
            <el-form-item label="显示 Reality 详情" v-if="nodeForm.TLS">
              <el-switch v-model="nodeForm.RealityShow" />
              <span style="margin-left: 10px; color: #909399; font-size: 12px;">在日志中显示 REALITY 连接详情（调试用）</span>
            </el-form-item>
            <el-form-item label="开启调试" v-if="nodeForm.TLS">
              <el-switch v-model="nodeForm.RealityDebug" />
              <span style="margin-left: 10px; color: #909399; font-size: 12px;">开启调试信息</span>
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
            <el-form-item label="上行带宽(Mbps)">
              <el-input-number v-model="nodeForm.Up" :min="1" :max="1000" style="width: 100%" />
            </el-form-item>
            <el-form-item label="下行带宽(Mbps)">
              <el-input-number v-model="nodeForm.Down" :min="1" :max="1000" style="width: 100%" />
            </el-form-item>
            <el-form-item label="跳跃间隔(秒)">
              <el-input-number v-model="nodeForm.HopInterval" :min="1" :max="3600" style="width: 100%" />
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
            <el-form-item label="上行带宽(Mbps)">
              <el-input-number v-model="nodeForm.Up" :min="1" :max="1000" style="width: 100%" />
            </el-form-item>
            <el-form-item label="下行带宽(Mbps)">
              <el-input-number v-model="nodeForm.Down" :min="1" :max="1000" style="width: 100%" />
            </el-form-item>
            <el-form-item label="跳跃间隔(秒)">
              <el-input-number v-model="nodeForm.HopInterval" :min="1" :max="3600" style="width: 100%" />
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
import { Plus, Upload, Delete, Edit, ArrowLeft, ArrowRight } from '@element-plus/icons-vue'
import { getNodes, createNode, updateNode, deleteNode, importNode, exportNode } from '@/api/nodes'

const nodes = ref([])
const selectedNodes = ref([])
const importDialogVisible = ref(false)
const formDialogVisible = ref(false)
const batchEditDialogVisible = ref(false)
const activeTab = ref('vmess')
const isEdit = ref(false)
const importForm = ref({ link: '' })

// 批量编辑相关
const batchEditForm = ref({ nodesText: '' })
const batchEditInputCount = ref(0)

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
  UDP: true,
  Up: 30,
  Down: 30,
  HopInterval: 60,
  Flow: '',
  ServerName: '',
  PublicKey: '',
  ShortID: '',
  ClientFingerprint: 'safari',
  Fingerprint: '',
  RealityShow: false,
  RealityDebug: false,
  Address: '',
  AlterId: ''
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
  // 确保 ID 是数字类型
  selectedNodes.value = selection.map(node => Number(node.id))
}

// 显示批量编辑对话框
const showBatchEditDialog = () => {
  if (selectedNodes.value.length === 0) {
    ElMessage.warning('请先选择要编辑的节点')
    return
  }
  batchEditForm.value.nodesText = ''
  batchEditInputCount.value = 0
  batchEditDialogVisible.value = true
}

// 处理批量编辑输入
const handleBatchEditInput = () => {
  const lines = batchEditForm.value.nodesText.trim().split(/[\r\n]+/).filter(line => line.trim())
  batchEditInputCount.value = lines.length
}

// 保存批量编辑
const handleBatchEditSave = async () => {
  if (selectedNodes.value.length === 0) {
    ElMessage.warning('没有选中的节点')
    return
  }

  const lines = batchEditForm.value.nodesText.trim().split(/[\r\n]+/).filter(line => line.trim())
  
  if (lines.length === 0) {
    ElMessage.warning('请输入节点信息')
    return
  }

  // 检查数量是否匹配
  if (lines.length !== selectedNodes.value.length) {
    try {
      await ElMessageBox.confirm(
        `选择的节点数量 (${selectedNodes.value.length}) 与输入的节点数量 (${lines.length}) 不一致，是否继续？`,
        '数量不匹配',
        {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }
      )
    } catch (error) {
      if (error === 'cancel') {
        return
      }
    }
  }

  // 依次更新节点
  let successCount = 0
  let errorCount = 0
  const sortedSelectedIds = [...selectedNodes.value].sort((a, b) => a - b)
  
  for (let i = 0; i < Math.min(lines.length, sortedSelectedIds.length); i++) {
    const line = lines[i]
    const nodeId = sortedSelectedIds[i]
    
    try {
      // 获取原节点信息
      const originalNode = nodes.value.find(n => n.id === nodeId)
      if (!originalNode) {
        errorCount++
        continue
      }
      
      // 解析新节点信息
      const newNode = parseNodeLink(line)
      if (!newNode) {
        errorCount++
        continue
      }
      
      // 保持原节点名称不变，rename 从新输入节点中获取
      const newNodeName = newNode.Name || newNode.name || ''
      newNode.Name = originalNode.name
      newNode.Rename = newNodeName
      
      await updateNode(nodeId, newNode)
      successCount++
    } catch (error) {
      console.error(`更新节点 ${nodeId} 失败:`, error)
      errorCount++
    }
  }

  if (successCount > 0) {
    ElMessage.success(`成功更新 ${successCount} 个节点`)
  }
  if (errorCount > 0) {
    ElMessage.warning(`${errorCount} 个节点更新失败`)
  }

  batchEditDialogVisible.value = false
  selectedNodes.value = []
  loadNodes()
}

// 解析节点链接
const parseNodeLink = (link) => {
  link = link.trim()
  if (!link) return null

  // 处理自定义格式（无协议头）
  if (!link.includes('://')) {
    // 尝试 / 分隔符
    let parts = link.split('/')
    if (parts.length >= 4) {
      return {
        Type: 'socks5',
        Server: parts[0],
        Port: parseInt(parts[1]),
        Username: parts[2],
        Password: parts[3],
        Name: parts[4] || generateAutoName(),
        Rename: parts[4] || '',
        Address: ''
      }
    }
    // 尝试 | 分隔符
    parts = link.split('|')
    if (parts.length >= 4) {
      return {
        Type: 'socks5',
        Server: parts[0],
        Port: parseInt(parts[1]),
        Username: parts[2],
        Password: parts[3],
        Name: parts[4] || generateAutoName(),
        Rename: parts[4] || '',
        Address: ''
      }
    }
    return null
  }

  // 处理各种协议
  try {
    const url = new URL(link)
    const protocol = url.protocol.slice(0, -1)
    const hash = url.hash ? decodeURIComponent(url.hash.slice(1)) : ''
    
    const baseNode = {
      Server: url.hostname,
      Port: parseInt(url.port),
      Name: hash || generateAutoName()
    }

    switch (protocol) {
      case 'ss':
        return parseSSLink(link, baseNode)
      case 'vmess':
        return parseVMessLink(link, baseNode)
      case 'trojan':
        return parseTrojanLink(link, baseNode)
      case 'vless':
        return parseVLESSLink(link, baseNode)
      case 'socks5':
      case 'socks':
        return parseSocksLink(link, baseNode)
      case 'hysteria2':
        return parseHysteria2Link(link, baseNode)
      default:
        return null
    }
  } catch (error) {
    console.error('解析链接失败:', error)
    return null
  }
}

// 生成自动名称
const generateAutoName = () => {
  const now = new Date()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  const random = Math.floor(10000 + Math.random() * 90000)
  return `SK5_${month}${day}_${random}`
}

// 解析 SS 链接
// 格式: ss://base64(method:password)@server:port#name
const parseSSLink = (link, baseNode) => {
  try {
    // 移除 ss:// 前缀
    const content = link.substring(5)
    
    // 分离名称部分
    let serverInfo = content
    let name = ''
    const hashIdx = content.indexOf('#')
    if (hashIdx !== -1) {
      serverInfo = content.substring(0, hashIdx)
      name = decodeURIComponent(content.substring(hashIdx + 1))
    }
    
    // 分离 base64 部分和服务器地址
    let base64Part = serverInfo
    let serverAddr = ''
    const atIdx = serverInfo.indexOf('@')
    if (atIdx !== -1) {
      base64Part = serverInfo.substring(0, atIdx)
      serverAddr = serverInfo.substring(atIdx + 1)
    }
    
    // URL 解码（处理 %3D 等）
    base64Part = decodeURIComponent(base64Part)
    
    // Base64 解码 - 尝试多种方式
    let decoded = ''
    try {
      decoded = atob(base64Part)
    } catch {
      try {
        // URL safe base64
        decoded = atob(base64Part.replace(/-/g, '+').replace(/_/g, '/'))
      } catch {
        // 尝试填充
        const pad = base64Part.length % 4
        if (pad) {
          base64Part += '='.repeat(4 - pad)
        }
        try {
          decoded = atob(base64Part)
        } catch {
          try {
            decoded = atob(base64Part.replace(/-/g, '+').replace(/_/g, '/'))
          } catch {
            return null
          }
        }
      }
    }
    
    // 解析 method:password
    const colonIdx = decoded.indexOf(':')
    if (colonIdx === -1) return null
    
    const method = decoded.substring(0, colonIdx)
    const password = decoded.substring(colonIdx + 1)
    
    // 解析服务器和端口
    let server = ''
    let port = 0
    if (serverAddr) {
      const portIdx = serverAddr.lastIndexOf(':')
      if (portIdx !== -1) {
        server = serverAddr.substring(0, portIdx)
        port = parseInt(serverAddr.substring(portIdx + 1))
      } else {
        server = serverAddr
        port = 8388
      }
    }
    
    return {
      Type: 'ss',
      Server: server,
      Port: port,
      Cipher: method,
      Password: password,
      Address: '',
      Name: name || baseNode.Name
    }
  } catch (error) {
    console.error('解析 SS 链接失败:', error)
    return null
  }
}

// 解析 VMess 链接
// 格式: vmess://base64(json)
const parseVMessLink = (link, baseNode) => {
  try {
    const b64 = link.substring(8) // 移除 'vmess://'
    
    // Base64 解码 - 尝试多种方式
    let decoded = ''
    try {
      decoded = atob(b64)
    } catch {
      try {
        // URL safe base64
        decoded = atob(b64.replace(/-/g, '+').replace(/_/g, '/'))
      } catch {
        // 尝试填充
        let padded = b64
        const pad = b64.length % 4
        if (pad) {
          padded += '='.repeat(4 - pad)
        }
        try {
          decoded = atob(padded)
        } catch {
          try {
            decoded = atob(padded.replace(/-/g, '+').replace(/_/g, '/'))
          } catch {
            return null
          }
        }
      }
    }
    
    const config = JSON.parse(decoded)
    
    return {
      Type: 'vmess',
      Server: config.add || config.host || baseNode.Server,
      Port: parseInt(config.port) || baseNode.Port,
      UUID: config.id || config.uuid,
      AlterId: String(config.aid || config.alterId || 0),
      Cipher: config.scy || config.cipher || 'auto',
      Network: config.net || config.network || 'tcp',
      Path: config.path || '',
      Host: config.host || config.sni || '',
      Address: config.address || '',
      TLS: config.tls === 'tls' || config.security === 'tls' || config.tls === true,
      Name: config.ps || config.remarks || baseNode.Name
    }
  } catch (error) {
    console.error('解析 VMess 链接失败:', error)
    return null
  }
}

// 解析 Trojan 链接
const parseTrojanLink = (link, baseNode) => {
  try {
    const url = new URL(link)
    return {
      Type: 'trojan',
      Server: url.hostname,
      Port: parseInt(url.port) || 443,
      Password: decodeURIComponent(url.username),
      SNI: url.searchParams.get('sni') || url.hostname,
      AllowInsecure: url.searchParams.get('allowInsecure') === '1',
      Address: '',
      Name: baseNode.Name
    }
  } catch (error) {
    console.error('解析 Trojan 链接失败:', error)
    return null
  }
}

// 解析 VLESS 链接
const parseVLESSLink = (link, baseNode) => {
  try {
    const url = new URL(link)
    const security = url.searchParams.get('security') || 'none'
    
    return {
      Type: 'vless',
      Server: url.hostname,
      Port: parseInt(url.port) || 443,
      UUID: decodeURIComponent(url.username),
      Network: url.searchParams.get('type') || 'tcp',
      Security: security,
      Path: url.searchParams.get('path') || '',
      Host: url.searchParams.get('host') || '',
      SNI: url.searchParams.get('sni') || '',
      TLS: security === 'tls' || security === 'reality',
      UDP: true, // VLESS 默认启用 UDP
      Cipher: 'auto', // VLESS 默认加密方式
      AlterId: '0',   // VLESS 默认 alterId
      Flow: url.searchParams.get('flow') || '',
      ServerName: url.searchParams.get('servername') || '',
      PublicKey: url.searchParams.get('pbk') || '',
      ShortID: url.searchParams.get('sid') || '',
      ClientFingerprint: url.searchParams.get('fp') || 'safari', // 默认指纹
      Fingerprint: url.searchParams.get('fp') || 'safari', // Reality 指纹
      RealityShow: false,
      RealityDebug: false,
      Address: '',
      Name: baseNode.Name
    }
  } catch (error) {
    console.error('解析 VLESS 链接失败:', error)
    return null
  }
}

// 解析 Socks5 链接
const parseSocksLink = (link, baseNode) => {
  try {
    const url = new URL(link)
    return {
      Type: 'socks5',
      Server: url.hostname,
      Port: parseInt(url.port) || 1080,
      Username: decodeURIComponent(url.username) || '',
      Password: decodeURIComponent(url.password) || '',
      Address: '',
      Name: baseNode.Name
    }
  } catch (error) {
    console.error('解析 Socks5 链接失败:', error)
    return null
  }
}

// 解析 Hysteria2 链接
const parseHysteria2Link = (link, baseNode) => {
  try {
    const url = new URL(link)
    // Hysteria2 格式: hysteria2://password@server:port?params#name
    // 密码在 username 位置（在 @ 前面）
    const password = decodeURIComponent(url.username) || ''
    return {
      Type: 'hysteria2',
      Server: url.hostname,
      Port: parseInt(url.port) || 443,
      Password: password,
      Host: url.searchParams.get('sni') || '', // SNI 映射到 Host 字段
      SkipCert: url.searchParams.get('insecure') === '1', // insecure 映射到 SkipCert 字段
      Up: parseInt(url.searchParams.get('up')) || 30, // 默认上行带宽 30
      Down: parseInt(url.searchParams.get('down')) || 30, // 默认下行带宽 30
      HopInterval: parseInt(url.searchParams.get('hop-interval')) || 60, // 默认跳跃间隔 60
      Address: '',
      Name: baseNode.Name
    }
  } catch (error) {
    console.error('解析 Hysteria2 链接失败:', error)
    return null
  }
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
    
    // 从大到小排序ID，从后往前删除避免ID变化问题
    const sortedIds = [...selectedNodes.value].sort((a, b) => b - a)
    
    for (const id of sortedIds) {
      await deleteNode(id)
    }
    
    ElMessage.success(`成功删除 ${selectedNodes.value.length} 个节点`)
    selectedNodes.value = []
    loadNodes()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error('批量删除失败: ' + (error.message || error))
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
      // 保留原始链接，让后端处理各种格式（包括自定义格式）
      let processLink = trimmedLink
      console.log('正在处理第', i + 1, '个节点:', processLink)
      
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
    Address: row.address || '',
    AlterId: row.alterId || '',
    ExtraConfig: row.extraConfig || '',
    Rename: row.rename || '',
    Flow: row.flow || '',
    ServerName: row.serverName || '',
    PublicKey: row.publicKey || '',
    ShortID: row.shortId || '',
    ClientFingerprint: row.clientFingerprint || 'safari',
    Fingerprint: row.fingerprint || '',
    RealityShow: row.realityShow || false,
    RealityDebug: row.realityDebug || false
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
    rename: nodeForm.value.Rename || '',
    up: nodeForm.value.Up || 30,
    down: nodeForm.value.Down || 30,
    hop_interval: nodeForm.value.HopInterval || 60,
    flow: nodeForm.value.Flow || '',
    server_name: nodeForm.value.ServerName || '',
    public_key: nodeForm.value.PublicKey || '',
    short_id: nodeForm.value.ShortID || '',
    client_fingerprint: nodeForm.value.ClientFingerprint || 'safari',
    fingerprint: nodeForm.value.Fingerprint || '',
    reality_show: nodeForm.value.RealityShow || false,
    reality_debug: nodeForm.value.RealityDebug || false,
    address: nodeForm.value.Address || '',
    alterId: nodeForm.value.AlterId || ''
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

/* 节点类型标签页美化样式 */
.node-type-tabs {
  :deep(.el-tabs__header) {
    background: linear-gradient(135deg, #f5f7fa 0%, #e4e7ed 100%);
    border-bottom: 2px solid #dcdfe6;
    margin: 0;
  }

  :deep(.el-tabs__nav) {
    border: none;
  }

  :deep(.el-tabs__item) {
    border: none;
    border-right: 1px solid #e4e7ed;
    transition: all 0.3s ease;
    font-weight: 500;
    color: #606266;
    padding: 0 16px;
    height: 42px;
    line-height: 42px;

    &:hover {
      color: #409eff;
      background: rgba(64, 158, 255, 0.1);
    }

    &.is-active {
      color: #fff;
      background: linear-gradient(135deg, #409eff 0%, #66b1ff 100%);
      box-shadow: 0 2px 8px rgba(64, 158, 255, 0.4);
      font-weight: 600;
      transform: translateY(-1px);

      .el-icon {
        color: #fff;
      }
    }

    &:first-child {
      border-radius: 4px 0 0 0;
    }

    &:last-child {
      border-right: none;
      border-radius: 0 4px 0 0;
    }
  }

  :deep(.el-tabs__content) {
    border: 1px solid #e4e7ed;
    border-top: none;
    border-radius: 0 0 4px 4px;
    background: #fff;
  }
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
