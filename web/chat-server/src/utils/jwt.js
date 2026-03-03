function decodeBase64Url(value) {
  if (!value) {
    return ""
  }
  const normalized = value.replace(/-/g, "+").replace(/_/g, "/")
  const padded = normalized + "=".repeat((4 - (normalized.length % 4)) % 4)
  return atob(padded)
}

// 只做前端展示/过期判断用途，不作为安全校验依据。
export function parseJwtPayload(token) {
  if (!token || typeof token !== "string") {
    return null
  }
  const parts = token.split(".")
  if (parts.length !== 3) {
    return null
  }
  try {
    const decoded = decodeBase64Url(parts[1])
    return JSON.parse(decoded)
  } catch (error) {
    return null
  }
}

export function getTokenExpireAt(token) {
  const payload = parseJwtPayload(token)
  if (!payload || typeof payload.exp !== "number") {
    return 0
  }
  return payload.exp
}

// skewSeconds 预留少量时钟偏差，避免边界时刻误判。
export function isTokenExpired(token, skewSeconds = 5) {
  const exp = getTokenExpireAt(token)
  if (!exp) {
    return true
  }
  const now = Math.floor(Date.now() / 1000)
  return exp <= now + skewSeconds
}
