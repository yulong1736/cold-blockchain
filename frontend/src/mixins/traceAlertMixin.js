import api, { getErrorMessage, unwrapData } from '../api'
import { buildTraceSummary } from '../utils/traceSummary'

export default {
  data() {
    return {
      activeTab: 'summary',
      alerts: [],
      loading: false,
      pollTimer: null,
      detailDrawerVisible: false,
      detailTraceId: '',
      pollIntervalMs: 3000
    }
  },
  computed: {
    refreshIntervalText() {
      const sec = this.pollIntervalMs / 1000
      if (sec >= 1 && Number.isInteger(sec)) return `每${sec}秒刷新一次`
      if (sec < 1) return `每${this.pollIntervalMs}毫秒刷新一次`
      return `每${sec}秒刷新一次`
    },
    traceSummaryList() {
      return buildTraceSummary(this.alerts, { timeField: 'record_time' })
    },
    detailAlerts() {
      if (!this.detailTraceId) return []
      return this.alerts
        .filter(a => a.trace_id === this.detailTraceId)
        .sort((a, b) => new Date(b.record_time) - new Date(a.record_time))
    }
  },
  mounted() {
    const conf = this.getAlertConfig()
    this.pollIntervalMs = conf.pollIntervalMs
    this.loadAlerts(true)
    this.pollTimer = setInterval(() => this.loadAlerts(false), this.pollIntervalMs)
  },
  beforeDestroy() {
    if (this.pollTimer) clearInterval(this.pollTimer)
  },
  methods: {
    getAlertConfig() {
      return {
        endpoint: '/alerts',
        limit: 100,
        pollIntervalMs: 3000,
        silentPollError: true
      }
    },
    openTraceDetail(traceId) {
      this.detailTraceId = traceId
      this.detailDrawerVisible = true
    },
    async loadAlerts(showLoading = true) {
      const conf = this.getAlertConfig()
      if (showLoading) this.loading = true
      try {
        const res = await api.get(conf.endpoint, { params: { limit: conf.limit } })
        this.alerts = unwrapData(res, [])
      } catch (e) {
        const shouldShowError = showLoading || !conf.silentPollError || this.alerts.length === 0
        if (shouldShowError) {
          this.$message.error(getErrorMessage(e, '加载告警失败'))
        }
      } finally {
        if (showLoading) this.loading = false
      }
    }
  }
}
