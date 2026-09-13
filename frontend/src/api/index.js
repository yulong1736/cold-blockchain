import axios from 'axios'
import { Message } from 'element-ui'
import store from '../store'
import { getErrorMessage, unwrapData } from '../utils/http'

const api = axios.create({
  baseURL: process.env.VUE_APP_API_BASE_URL || '/api',
  timeout: 10000
})

// 请求拦截器：统一附带 JWT
api.interceptors.request.use(
  config => {
    const token = store.state.token
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => Promise.reject(error)
)

// 响应拦截器：401 跳转登录；5xx/网络错误时全局提示，4xx 由页面自行提示
api.interceptors.response.use(
  response => response,
  error => {
    if (error.response && error.response.status === 401) {
      const reqUrl = error.config?.url || ''
      // 登录失败也返回 401，不应清 token 并整页跳转，否则登录页无法展示错误信息
      if (reqUrl.includes('/auth/login')) {
        return Promise.reject(error)
      }
      store.dispatch('logout')
      window.location.href = '/login'
      return Promise.reject(error)
    }
    const status = error.response?.status
    const is5xx = status >= 500
    const isNetworkError = !error.response
    if (is5xx || isNetworkError) {
      Message.error(getErrorMessage(error))
    }
    return Promise.reject(error)
  }
)

export const sendForgotPasswordCode = payload => api.post('/auth/forgot-password/send-code', payload)
export const resetForgotPassword = payload => api.post('/auth/forgot-password/reset', payload)

export default api
export { getErrorMessage, unwrapData }
