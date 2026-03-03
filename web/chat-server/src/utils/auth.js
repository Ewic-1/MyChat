import axios from "axios"
import { isTokenExpired } from "./jwt"

const refreshClient = axios.create()
let refreshingPromise = null

const AUTH_FREE_PATHS = [
  "/login",
  "/register",
  "/user/sendSmsCode",
  "/user/smsLogin",
  "/auth/refresh",
]

function normalizeUrl(url) {
  if (!url) {
    return ""
  }
  return String(url).split("?")[0]
}

export function isAuthFreeRequest(url) {
  const path = normalizeUrl(url)
  return AUTH_FREE_PATHS.some((item) => path.endsWith(item))
}

export async function refreshAccessToken(store) {
  // 并发请求共享同一个 refresh 过程，避免重复刷新。
  if (refreshingPromise) {
    return refreshingPromise
  }

  const currentUser = store.state.userInfo || {}
  const refreshToken = currentUser.refresh_token
  if (!refreshToken) {
    throw new Error("missing refresh token")
  }
  if (isTokenExpired(refreshToken)) {
    throw new Error("refresh token expired")
  }

  refreshingPromise = (async () => {
    const rsp = await refreshClient.post(
      store.state.backendUrl + "/auth/refresh",
      { refresh_token: refreshToken }
    )
    if (rsp.data.code !== 200 || !rsp.data.data || !rsp.data.data.token) {
      throw new Error("refresh failed")
    }
    // 刷新成功后立即回写本地 token。
    store.commit("setAuthTokens", rsp.data.data)
    return rsp.data.data
  })()

  try {
    return await refreshingPromise
  } finally {
    refreshingPromise = null
  }
}

export async function ensureValidAccessToken(store) {
  const currentUser = store.state.userInfo || {}
  const accessToken = currentUser.token
  if (!accessToken) {
    throw new Error("missing access token")
  }
  if (!isTokenExpired(accessToken)) {
    return accessToken
  }
  // access token 过期时自动走 refresh 换新。
  const refreshed = await refreshAccessToken(store)
  return refreshed.token
}
