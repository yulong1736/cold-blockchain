<template>
  <div id="app" :class="{ 'app-theme-dark': isDarkTheme }">
    <el-container class="app-layout">
      <el-header v-if="!isAuthStandalone" class="app-header">
        <div class="header-content">
          <div class="header-left">
            <el-button
              v-if="isAuthenticated"
              type="text"
              class="sidebar-toggle"
              @click="toggleSidebar"
              aria-label="折叠/展开侧栏"
            >
              <i :class="sidebarCollapsed ? 'el-icon-d-arrow-right' : 'el-icon-d-arrow-left'"></i>
            </el-button>
            <h1 class="page-title">跨境冷链物流溯源与监管系统</h1>
          </div>
          <div class="user-info" v-if="isAuthenticated">
            <el-breadcrumb separator="/" class="breadcrumb">
              <el-breadcrumb-item :to="{ path: '/dashboard' }">首页</el-breadcrumb-item>
              <el-breadcrumb-item v-if="breadcrumbName">{{ breadcrumbName }}</el-breadcrumb-item>
            </el-breadcrumb>
            <el-button type="text" class="theme-toggle" @click="toggleTheme" :title="isDarkTheme ? '切换浅色' : '切换深色'">
              <i :class="isDarkTheme ? 'el-icon-sunny' : 'el-icon-moon'"></i>
            </el-button>
            <router-link to="/account" class="username-link">{{ username }}</router-link>
            <el-button type="text" @click="logout">退出</el-button>
          </div>
        </div>
      </el-header>
      <el-container>
        <el-aside :width="sidebarWidth" class="app-aside" :class="{ 'is-collapsed': sidebarCollapsed }" v-if="isAuthenticated">
          <el-menu
            :default-active="activeMenu"
            router
            class="sidebar-menu"
            background-color="#1F2D3D"
            text-color="#bfcbd9"
            active-text-color="#2C7BE5"
          >
            <el-menu-item index="/dashboard">
              <i class="el-icon-s-home"></i>
              <span>仪表盘</span>
            </el-menu-item>
            <el-menu-item index="/products" v-if="canManageProducts">
              <i class="el-icon-goods"></i>
              <span>商品管理</span>
            </el-menu-item>
            <el-menu-item index="/trace">
              <i class="el-icon-search"></i>
              <span>追溯查询</span>
            </el-menu-item>
            <el-menu-item index="/temperature" v-if="canManageTemperature">
              <i class="el-icon-odometer"></i>
              <span>温控监控</span>
            </el-menu-item>
            <el-menu-item index="/transport" v-if="canManageTransport">
              <i class="el-icon-truck"></i>
              <span>运输管理</span>
            </el-menu-item>
            <el-menu-item index="/alerts" v-if="canViewAlerts">
              <i class="el-icon-warning"></i>
              <span>告警</span>
            </el-menu-item>
            <el-menu-item index="/regulator" v-if="isRegulator">
              <i class="el-icon-view"></i>
              <span>监管审计</span>
            </el-menu-item>
            <el-menu-item index="/regulator-alerts" v-if="isRegulator">
              <i class="el-icon-warning-outline"></i>
              <span>全局告警</span>
            </el-menu-item>
          </el-menu>
        </el-aside>
        <el-main class="app-main" :class="{ 'app-main--standalone-auth': isAuthStandalone }">
          <router-view />
        </el-main>
      </el-container>
    </el-container>
  </div>
</template>

<script>
import { mapGetters, mapMutations } from 'vuex'

const routeTitleMap = {
  '/dashboard': '仪表盘',
  '/products': '商品管理',
  '/trace': '追溯查询',
  '/temperature': '温控监控',
  '/transport': '运输管理',
  '/alerts': '告警',
  '/regulator': '监管审计',
  '/regulator-alerts': '全局告警',
  '/account': '账号与安全'
}

const IDLE_TIMEOUT = 30 * 1000 // 30 秒空闲超时

export default {
  name: 'App',
  data() {
    return {
      idleTimer: null,
      idleWarningVisible: false
    }
  },
  computed: {
    ...mapGetters(['isAuthenticated', 'username', 'userRole', 'isDarkTheme', 'sidebarCollapsed']),
    sidebarWidth() {
      return this.sidebarCollapsed ? '64px' : '200px'
    },
    activeMenu() {
      return this.$route.path
    },
    breadcrumbName() {
      return routeTitleMap[this.$route.path] || ''
    },
    canManageProducts() {
      return this.userRole === 'producer'
    },
    canManageTemperature() {
      return this.userRole === 'warehouse'
    },
    canManageTransport() {
      return this.userRole === 'logistics'
    },
    canViewAlerts() {
      return this.userRole === 'warehouse' || this.userRole === 'logistics'
    },
    isRegulator() {
      return this.userRole === 'regulator'
    },
    /** 登录/注册页使用全屏自有布局，不显示全局顶栏标题 */
    isAuthStandalone() {
      const p = this.$route.path
      return p === '/login' || p === '/register'
    }
  },
  watch: {
    isAuthenticated(val) {
      if (val) {
        this.startIdleTimer()
      } else {
        this.stopIdleTimer()
      }
    }
  },
  mounted() {
    if (this.isAuthenticated) {
      this.startIdleTimer()
    }
  },
  beforeDestroy() {
    this.stopIdleTimer()
  },
  methods: {
    ...mapMutations(['TOGGLE_THEME', 'SET_SIDEBAR_COLLAPSED']),
    toggleSidebar() {
      this.SET_SIDEBAR_COLLAPSED(!this.sidebarCollapsed)
    },
    toggleTheme() {
      this.TOGGLE_THEME()
    },
    logout() {
      this.stopIdleTimer()
      this.$store.dispatch('logout')
      this.$router.push('/login')
    },
    /** 重置空闲计时器 —— 用户有操作时调用 */
    resetIdleTimer() {
      if (!this.isAuthenticated) return
      this.clearIdleTimer()
      this.idleTimer = setTimeout(() => {
        this.handleIdleTimeout()
      }, IDLE_TIMEOUT)
    },
    clearIdleTimer() {
      if (this.idleTimer) {
        clearTimeout(this.idleTimer)
        this.idleTimer = null
      }
    },
    removeActivityListeners() {
      const events = ['mousedown', 'mousemove', 'keydown', 'scroll', 'touchstart']
      events.forEach(event => {
        document.removeEventListener(event, this.resetIdleTimer)
      })
    },
    startIdleTimer() {
      this.resetIdleTimer()
      const events = ['mousedown', 'mousemove', 'keydown', 'scroll', 'touchstart']
      events.forEach(event => {
        document.addEventListener(event, this.resetIdleTimer, { passive: true })
      })
    },
    stopIdleTimer() {
      this.removeActivityListeners()
      this.clearIdleTimer()
    },
    handleIdleTimeout() {
      if (!this.isAuthenticated) return
      this.stopIdleTimer()
      this.$message.warning('用户登录超时，即将自动退出登录')
      this.$store.dispatch('logout')
      this.$router.push('/login')
    }
  }
}
</script>

<style scoped>
#app {
  min-height: 100vh;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

.app-layout {
  min-height: 100vh;
}

.app-header {
  height: 56px;
  padding: 0 var(--content-padding);
  background: var(--bg-header);
  border-bottom: 1px solid #e4e7ed;
  display: flex;
  align-items: center;
}
.app-theme-dark .app-header {
  border-bottom-color: #3a3a3a;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  height: 100%;
}

.header-left {
  display: flex;
  align-items: center;
  gap: var(--spacing-unit);
}

.sidebar-toggle {
  padding: 8px;
  min-width: 44px;
  min-height: 44px;
  color: var(--text-regular);
}
.sidebar-toggle:hover {
  color: var(--color-primary);
}

.page-title {
  margin: 0;
  font-size: var(--font-size-page-title);
  font-weight: bold;
  color: var(--color-primary);
}
.app-theme-dark .page-title {
  color: #60a5fa;
}

.user-info {
  display: flex;
  align-items: center;
  gap: 16px;
}

.breadcrumb {
  font-size: 0.875rem;
  margin-right: 8px;
}
.app-theme-dark .breadcrumb ::v-deep .el-breadcrumb__inner,
.app-theme-dark .breadcrumb ::v-deep .el-breadcrumb__inner a,
.app-theme-dark .breadcrumb ::v-deep .el-breadcrumb__item:last-child .el-breadcrumb__inner {
  color: #d1d5db !important;
}
.app-theme-dark .breadcrumb ::v-deep .el-breadcrumb__separator {
  color: #9ca3af !important;
}

.theme-toggle {
  padding: 8px;
  min-width: 44px;
  min-height: 44px;
  color: var(--text-regular);
}
.theme-toggle:hover {
  color: var(--color-primary);
}

.username-link {
  color: var(--color-primary);
  text-decoration: none;
  font-size: 0.95rem;
  padding: 8px 0;
}
.username-link:hover {
  text-decoration: underline;
}

.app-aside {
  background: #1F2D3D;
  overflow-x: hidden;
  transition: width 0.2s ease;
}
.sidebar-menu {
  height: 100%;
  border-right: none;
}
.sidebar-menu .el-menu-item {
  min-height: 44px;
  line-height: 44px;
  color: #bfcbd9;
}
.sidebar-menu .el-menu-item:hover {
  background-color: rgba(0, 0, 0, 0.15) !important;
  color: #fff;
}
/* 选中项：浅色文字 + 蓝色背景，保证可见 */
.sidebar-menu .el-menu-item.is-active {
  background-color: #2C7BE5 !important;
  color: #fff !important;
}
.sidebar-menu .el-menu-item.is-active i {
  color: #fff !important;
}
.app-aside.is-collapsed .sidebar-menu .el-menu-item span {
  display: none;
}
.app-aside.is-collapsed .sidebar-menu .el-menu-item {
  text-align: center;
  padding-left: 0 !important;
}
.app-aside.is-collapsed .sidebar-menu .el-menu-item i {
  margin-right: 0;
}

.app-main {
  background: var(--bg-main);
  overflow-y: auto;
  padding: var(--content-padding);
  min-height: calc(100vh - 56px);
}

.app-main--standalone-auth {
  padding: 0;
  min-height: 100vh;
}
</style>
