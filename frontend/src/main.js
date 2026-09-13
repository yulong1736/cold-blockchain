import Vue from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import ElementUI from 'element-ui'
import 'element-ui/lib/theme-chalk/index.css'
import './assets/styles/main.css'
import { formatDateTimeWithTZ } from './utils/dateFormat'

Vue.config.productionTip = false
Vue.use(ElementUI)

// 全局时间格式化：年-月-日 时:分:秒 时区
Vue.filter('formatDateTimeTZ', formatDateTimeWithTZ)

function applyTheme(theme) {
  const isDark = theme === 'dark'
  const themeClass = isDark ? 'app-theme-dark' : 'app-theme-light'
  const nodes = [document.documentElement, document.body, document.getElementById('app')].filter(Boolean)

  // 先清理再设置，避免多次快速点击后 class 残留
  nodes.forEach((node) => {
    node.classList.remove('app-theme-dark', 'app-theme-light')
    node.classList.add(themeClass)
  })

  document.documentElement.setAttribute('data-theme', isDark ? 'dark' : 'light')
}

// 首次加载应用时应用主题
applyTheme(store.state.theme)

// 在每次主题 mutation 后同步主题，确保切换原子生效
store.subscribe((mutation, state) => {
  if (mutation.type === 'SET_THEME' || mutation.type === 'TOGGLE_THEME') {
    applyTheme(state.theme)
  }
})

new Vue({
  router,
  store,
  render: h => h(App)
}).$mount('#app')
