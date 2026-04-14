import axios from 'axios'

const api = axios.create({
  baseURL: '/api',
  timeout: 10000
})

export const getServices = () => api.get('/services')
export const createService = (data) => api.post('/services', data)
export const updateService = (id, data) => api.put(`/services/${id}`, data)
export const deleteService = (id) => api.delete(`/services/${id}`)
export const startService = (id) => api.post(`/services/${id}/start`)
export const stopService = (id) => api.post(`/services/${id}/stop`)
export const restartService = (id) => api.post(`/services/${id}/restart`)
export const getLogs = (id, lines = 100) => api.get(`/services/${id}/logs?lines=${lines}`)
export const discover = (dirs) => api.post('/discover', { dirs })

export default api