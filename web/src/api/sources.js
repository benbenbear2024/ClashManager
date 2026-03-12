import request from '@/utils/request'

export const getSources = async () => {
  return await request.get('/sources')
}

export const createSource = async (data) => {
  return await request.post('/sources', data)
}

export const updateSource = async (id, data) => {
  return await request.put(`/sources/${id}`, data)
}

export const deleteSource = async (id) => {
  return await request.delete(`/sources/${id}`)
}

export const syncSource = async (id) => {
  return await request.post(`/sources/${id}/sync`)
}

export const testSource = async (url) => {
  return await request.post('/sources/test', { url })
}
