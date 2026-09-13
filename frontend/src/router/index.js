import Vue from 'vue'
import VueRouter from 'vue-router'
import store from '../store'

Vue.use(VueRouter)

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/Register.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('../views/Dashboard.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/products',
    name: 'Products',
    component: () => import('../views/Products.vue'),
    meta: { requiresAuth: true, roles: ['producer'] }
  },
  {
    path: '/trace',
    name: 'Trace',
    component: () => import('../views/Trace.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/temperature',
    name: 'Temperature',
    component: () => import('../views/Temperature.vue'),
    meta: { requiresAuth: true, roles: ['warehouse'] }
  },
  {
    path: '/transport',
    name: 'Transport',
    component: () => import('../views/Transport.vue'),
    meta: { requiresAuth: true, roles: ['logistics'] }
  },
  {
    path: '/alerts',
    name: 'Alert',
    component: () => import('../views/Alert.vue'),
    meta: { requiresAuth: true, roles: ['warehouse', 'logistics'] }
  },
  {
    path: '/regulator',
    name: 'Regulator',
    component: () => import('../views/Regulator.vue'),
    meta: { requiresAuth: true, roles: ['regulator'] }
  },
  {
    path: '/regulator-alerts',
    name: 'RegulatorAlert',
    component: () => import('../views/RegulatorAlert.vue'),
    meta: { requiresAuth: true, roles: ['regulator'] }
  },
  {
    path: '/account',
    name: 'Account',
    component: () => import('../views/Account.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/',
    redirect: '/dashboard'
  }
]

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
})

router.beforeEach((to, from, next) => {
  const requiresAuth = to.matched.some(record => record.meta.requiresAuth)
  const isAuthenticated = store.getters.isAuthenticated
  const allowedRoles = Array.isArray(to.meta.roles) ? to.meta.roles : null

  if (requiresAuth && !isAuthenticated) {
    next('/login')
  } else if (allowedRoles && !allowedRoles.includes(store.getters.userRole)) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router
