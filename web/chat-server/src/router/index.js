import { createRouter, createWebHistory } from 'vue-router'
import store from '../store/index.js'
import { ensureValidAccessToken } from '../utils/auth'

const routes = [
  {
    path: '/',
    redirect: { name: 'Login' }
  },
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/access/Login.vue')
  },
  {
    path: '/smsLogin',
    name: 'smsLogin',
    component: () => import('../views/access/SmsLogin.vue')
  },
  {
    path: '/register',
    name: 'Register',
    component: () => import('../views/access/Register.vue')
  },
  {
    path: '/chat/owninfo',
    name: 'OwnInfo',
    component: () => import('../views/chat/user/OwnInfo.vue')
  },
  {
    path: '/chat/contactlist',
    name: 'ContactList',
    component: () => import('../views/chat/contact/ContactList.vue')
  },
  {
    path: '/chat/:id',
    name: 'ContactChat',
    component: () => import('../views/chat/contact/ContactChat.vue')
  },
  {
    path: '/chat/sessionList',
    name: 'SessionList',
    component: () => import('../views/chat/session/SessionList.vue')
  },
  {
    path: '/manager',
    name: 'Manager',
    component: () => import('../views/manager/Manager.vue')
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

const publicPaths = new Set(['/login', '/register', '/smsLogin'])

router.beforeEach(async (to, from, next) => {
  // 公开页不做鉴权拦截。
  if (publicPaths.has(to.path)) {
    next()
    return
  }

  const token = (store.state.userInfo || {}).token
  if (!token) {
    store.commit('cleanUserInfo')
    next('/login')
    return
  }

  try {
    // 进入受保护页面前，确保 access token 可用（必要时自动 refresh）。
    await ensureValidAccessToken(store)
    next()
  } catch (error) {
    // token 无法恢复时清理登录态并回到登录页。
    store.commit('cleanUserInfo')
    next('/login')
  }
})

export default router
