/**
 * 按追溯码汇总：将记录列表按 trace_id 分组，得到每类一条的汇总行
 * @param {Array} items - 原始记录（每项需有 trace_id）
 * @param {Object} config - timeField; sortDesc; countAbnormal; withBlockchain: 是否按 trace 汇总上链笔数
 * @returns {Array} 汇总行 [{ trace_id, trace_id_short, count, latest_time 或 latest_arrival, abnormal_count?, blockchain_stats? }]
 */
export function buildTraceSummary(items, config = {}) {
  const {
    timeField = 'record_time',
    sortDesc = true,
    countAbnormal = false,
    withBlockchain = false
  } = config

  const byTrace = {}
  const key = timeField === 'arrival_time' ? 'latest_arrival' : 'latest_time'

  for (const item of items || []) {
    const tid = item.trace_id || ''
    if (!byTrace[tid]) {
      byTrace[tid] = {
        trace_id: tid,
        trace_id_short: tid.length > 20 ? tid.slice(0, 16) + '...' : tid,
        count: 0,
        [key]: null,
        ...(countAbnormal ? { abnormal_count: 0 } : {}),
        ...(withBlockchain ? { blockchain_stats: { total: 0, pending_tx: 0 } } : {})
      }
    }
    const row = byTrace[tid]
    row.count++
    if (countAbnormal && item.is_abnormal) row.abnormal_count++
    if (withBlockchain && row.blockchain_stats) {
      row.blockchain_stats.total += 1
      if (!item.blockchain_tx_hash) row.blockchain_stats.pending_tx += 1
    }
    const t = item[timeField] ? new Date(item[timeField]) : null
    if (t && (!row[key] || t > new Date(row[key]))) {
      row[key] = item[timeField]
    }
  }

  const list = Object.values(byTrace)
  list.sort((a, b) => {
    const va = a[key] || ''
    const vb = b[key] || ''
    return sortDesc ? (vb > va ? 1 : -1) : (va > vb ? 1 : -1)
  })
  return list
}
