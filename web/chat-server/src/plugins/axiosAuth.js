import axios from "axios"
import { ensureValidAccessToken, isAuthFreeRequest, refreshAccessToken } from "../utils/auth"

let installed = false

const PUBLIC_ROUTES = new Set(["/login", "/register", "/smsLogin"])

function redirectToLoginIfNeeded(router) {
  const currentPath = router.currentRoute.value.path
  if (!PUBLIC_ROUTES.has(currentPath)) {
    router.push("/login")
  }
}

export function setupAxiosAuth(store, router) {
  if (installed) {
    return
  }
  installed = true

  axios.interceptors.request.use(
    async (config) => {
      // 登录、注册、刷新接口不注入 Authorization 头。
      if (isAuthFreeRequest(config.url)) {
        return config
      }

      const userInfo = store.state.userInfo || {}
      if (!userInfo.token) {
        return config
      }

      let accessToken = ""
      try {
        // 发送请求前确保 access token 有效。
        accessToken = await ensureValidAccessToken(store)
      } catch (error) {
        store.commit("cleanUserInfo")
        redirectToLoginIfNeeded(router)
        return Promise.reject(error)
      }

      config.headers = config.headers || {}
      config.headers.Authorization = "Bearer " + accessToken
      return config
    },
    (error) => Promise.reject(error)
  )

  axios.interceptors.response.use(
    (response) => response,
    async (error) => {
      const originalRequest = error.config || {}
      const status = error.response ? error.response.status : 0

      if (status !== 401) {
        return Promise.reject(error)
      }

      if (originalRequest.__isRetryRequest || isAuthFreeRequest(originalRequest.url)) {
        return Promise.reject(error)
      }

      try {
        // access token 失效时尝试刷新一次并自动重放原请求。
        const refreshed = await refreshAccessToken(store)
        originalRequest.__isRetryRequest = true
        originalRequest.headers = originalRequest.headers || {}
        originalRequest.headers.Authorization = "Bearer " + refreshed.token
        return axios(originalRequest)
      } catch (refreshError) {
        // refresh 失败则视为登录失效。
        store.commit("cleanUserInfo")
        redirectToLoginIfNeeded(router)
        return Promise.reject(refreshError)
      }
    }
  )
}
