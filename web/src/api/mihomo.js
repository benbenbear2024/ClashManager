import request from '@/utils/request'

// 获取 Mihomo 状态
export const getMihomoStatus = () => {
  return request({
    url: '/settings/mihomo/status',
    method: 'get'
  })
}

// 启动 Mihomo
export const startMihomo = () => {
  return request({
    url: '/settings/mihomo/start',
    method: 'post'
  })
}

// 停止 Mihomo
export const stopMihomo = () => {
  return request({
    url: '/settings/mihomo/stop',
    method: 'post'
  })
}

// 重载 Mihomo 配置
export const reloadMihomo = () => {
  return request({
    url: '/settings/mihomo/reload',
    method: 'post'
  })
}

// 检查L2TP连接状态
export const checkL2TPConnection = (nodeName) => {
  return request({
    url: '/settings/mihomo/check-connection',
    method: 'post',
    data: { nodeName }
  })
}
