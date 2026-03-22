import request from '@/utils/request'

// 获取配置文件内容
export const getConfigContent = () => {
  return request({
    url: '/settings/config/content',
    method: 'get'
  })
}

// 保存配置文件内容
export const saveConfigContent = (content) => {
  return request({
    url: '/settings/config/content',
    method: 'post',
    data: { content }
  })
}
