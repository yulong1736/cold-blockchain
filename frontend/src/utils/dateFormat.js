/**
 * 格式：年-月-日 时-分-秒 时区（如 2025-02-11 14:30:00 +08:00）
 * 支持字符串、Date、或后端返回的对象；零值时间显示为「未记录」
 */
export function formatDateTimeWithTZ(val) {
  if (val == null || val === '') return '-'
  // 后端可能返回对象（如 { Time: "..." }）或嵌套结构
  let str = val
  if (typeof val === 'object' && val !== null) {
    if (typeof val.Time === 'string') str = val.Time
    else if (typeof val.time === 'string') str = val.time
    else if (val instanceof Date) str = val.toISOString()
    else str = String(val)
  } else if (typeof val !== 'string') {
    str = String(val)
  }
  const d = new Date(str)
  if (isNaN(d.getTime())) return '-'
  // Go 零值时间 0001-01-01 视为未记录
  if (d.getFullYear() <= 1) return '未记录'
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const h = String(d.getHours()).padStart(2, '0')
  const min = String(d.getMinutes()).padStart(2, '0')
  const s = String(d.getSeconds()).padStart(2, '0')
  const offset = -d.getTimezoneOffset()
  const sign = offset >= 0 ? '+' : '-'
  const absM = Math.abs(offset)
  const tzHour = Math.floor(absM / 60)
  const tzMin = absM % 60
  const tz = `${sign}${String(tzHour).padStart(2, '0')}:${String(tzMin).padStart(2, '0')}`
  return `${y}-${m}-${day} ${h}:${min}:${s} ${tz}`
}

/**
 * 格式：年-月-日 时-分-秒（不含时区，用于图表横坐标等）
 */
export function formatDateTimeNoTZ(val) {
  if (val == null || val === '') return ''
  let str = val
  if (typeof val === 'object' && val !== null) {
    if (typeof val.Time === 'string') str = val.Time
    else if (typeof val.time === 'string') str = val.time
    else if (val instanceof Date) str = val.toISOString()
    else str = String(val)
  } else if (typeof val !== 'string') {
    str = String(val)
  }
  const d = new Date(str)
  if (isNaN(d.getTime())) return ''
  if (d.getFullYear() <= 1) return '' // 零值时间不显示
  const y = d.getFullYear()
  const m = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const h = String(d.getHours()).padStart(2, '0')
  const min = String(d.getMinutes()).padStart(2, '0')
  const s = String(d.getSeconds()).padStart(2, '0')
  return `${y}-${m}-${day} ${h}:${min}:${s}`
}
