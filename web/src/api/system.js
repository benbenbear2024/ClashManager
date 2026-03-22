import request from '@/utils/request'

export function getSystemInfo() {
  return request({
    url: '/system/info',
    method: 'get'
  })
}
