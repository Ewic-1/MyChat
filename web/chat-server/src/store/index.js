import { createStore } from 'vuex'

const backendUrl = process.env.VUE_APP_BACKEND_URL || 'http://127.0.0.1:8000'
const wsUrl = process.env.VUE_APP_WS_URL || 'ws://127.0.0.1:8000'

export default createStore({
  state: {
    // Default to local backend for development.
    // Override with VUE_APP_BACKEND_URL / VUE_APP_WS_URL if needed.
    backendUrl,
    wsUrl,
    userInfo: (sessionStorage.getItem('userInfo') && JSON.parse(sessionStorage.getItem('userInfo'))) || {},
    socket: null,
  },
  getters: {
  },
  mutations: {
    setUserInfo(state, userInfo) {
      // 保留已有 token 字段，避免部分接口只返回用户资料时覆盖鉴权信息。
      const mergedUserInfo = { ...state.userInfo, ...userInfo }
      state.userInfo = mergedUserInfo
      sessionStorage.setItem('userInfo', JSON.stringify(mergedUserInfo))
    },
    setAuthTokens(state, tokenData) {
      // 仅更新 token 相关字段，配合 refresh 续期流程使用。
      const mergedUserInfo = {
        ...state.userInfo,
        token: tokenData.token,
        refresh_token: tokenData.refresh_token,
        access_expires_at: tokenData.access_expires_at,
        refresh_expires_at: tokenData.refresh_expires_at,
      }
      state.userInfo = mergedUserInfo
      sessionStorage.setItem('userInfo', JSON.stringify(mergedUserInfo))
    },
    cleanUserInfo(state) {
      state.userInfo = {}
      sessionStorage.removeItem('userInfo')
    }
  },
  actions: {
  },
  modules: {
  }
})
