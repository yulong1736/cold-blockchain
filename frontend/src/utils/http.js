export function unwrapData(response, defaultValue = null) {
  return response?.data?.data ?? defaultValue
}

export function getErrorMessage(error, fallback = '请求失败，请稍后重试') {
  const msg = error?.response?.data?.error || error?.message || fallback
  return typeof msg === 'string' ? msg : fallback
}
