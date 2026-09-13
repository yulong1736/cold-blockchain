import Vue from 'vue'
import Vuex from 'vuex'
import api, { unwrapData } from '../api'

Vue.use(Vuex)

export default new Vuex.Store({
  state: {
    token: localStorage.getItem('token') || '',
    user: JSON.parse(localStorage.getItem('user') || 'null'),
    theme: localStorage.getItem('theme') || 'light',
    sidebarCollapsed: localStorage.getItem('sidebarCollapsed') === 'true'
  },
  getters: {
    isAuthenticated: state => !!state.token,
    isDarkTheme: state => state.theme === 'dark',
    sidebarCollapsed: state => state.sidebarCollapsed,
    username: state => state.user ? state.user.username : '',
    userRole: state => state.user ? state.user.role : '',
    userId: state => state.user ? state.user.id : null,
    companyName: state => state.user ? (state.user.company_name || '') : ''
  },
  mutations: {
    SET_TOKEN(state, token) {
      state.token = token
      localStorage.setItem('token', token)
    },
    SET_USER(state, user) {
      state.user = user
      localStorage.setItem('user', JSON.stringify(user))
    },
    CLEAR_AUTH(state) {
      state.token = ''
      state.user = null
      localStorage.removeItem('token')
      localStorage.removeItem('user')
    },
    SET_THEME(state, theme) {
      const normalizedTheme = theme === 'dark' ? 'dark' : 'light'
      state.theme = normalizedTheme
      localStorage.setItem('theme', normalizedTheme)
    },
    TOGGLE_THEME(state) {
      const nextTheme = state.theme === 'dark' ? 'light' : 'dark'
      state.theme = nextTheme
      localStorage.setItem('theme', nextTheme)
    },
    SET_SIDEBAR_COLLAPSED(state, collapsed) {
      state.sidebarCollapsed = collapsed
      localStorage.setItem('sidebarCollapsed', String(collapsed))
    }
  },
  actions: {
    async login({ commit }, credentials) {
      const response = await api.post('/auth/login', credentials)
      const payload = unwrapData(response, {})
      commit('SET_TOKEN', payload.token || '')
      commit('SET_USER', payload.user || null)
      return response
    },
    async register({ commit }, userData) {
      return api.post('/auth/register', userData)
    },
    logout({ commit }) {
      commit('CLEAR_AUTH')
    }
  }
})
