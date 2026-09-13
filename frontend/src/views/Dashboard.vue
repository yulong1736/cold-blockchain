<template>
  <div class="dashboard-container page-content-area" v-loading="statsLoading">
    <!-- Hero / Overview -->
    <div class="dashboard-hero">
      <div class="hero-bg"></div>
      <div class="hero-content">
        <div class="hero-left">
          <h1 class="hero-title">冷链溯源总览</h1>
          <p class="hero-subtitle">
            跨境冷链物流全链路可视化<br>
            按角色展示商品、温控、运输与区块链审计数据
          </p>
          <div class="hero-meta">
            <span class="hero-meta-item">
              <i class="el-icon-time"></i>
              <span>统计周期内数据快照，15 秒自动刷新</span>
            </span>
            <span class="hero-meta-item">
              <i class="el-icon-link"></i>
              <span>
                区块链：
                <span
                  :class="[
                    'hero-badge',
                    stats.blockchainConnected ? 'hero-badge-success' : 'hero-badge-danger'
                  ]"
                >
                  {{ stats.blockchainConnected ? '已连接' : '未连接' }}
                </span>
              </span>
            </span>
          </div>
        </div>
        <div class="hero-right">
          <el-row :gutter="12">
            <el-col :span="12">
              <div class="hero-kpi-card glass">
                <div class="stat-label">商品总数</div>
                <div class="stat-value">{{ stats.totalProducts }}</div>
              </div>
            </el-col>
            <el-col :span="12">
              <div class="hero-kpi-card glass">
                <div class="stat-label">链上记录数</div>
                <div class="stat-value">{{ stats.blockchainRecords }}</div>
              </div>
            </el-col>
          </el-row>
          <el-row :gutter="12" style="margin-top: 8px">
            <el-col :span="12">
              <div class="hero-kpi-card glass">
                <div class="stat-label">异常告警</div>
                <div class="stat-value">{{ stats.abnormalCount }}</div>
              </div>
            </el-col>
            <el-col :span="12">
              <div class="hero-kpi-card glass">
                <div class="stat-label">今日查询</div>
                <div class="stat-value">{{ stats.todayQueries }}</div>
              </div>
            </el-col>
          </el-row>
          <el-row :gutter="12" style="margin-top: 8px">
            <el-col :span="12">
              <div class="hero-kpi-card glass">
                <div class="stat-label">温控记录数</div>
                <div class="stat-value">{{ stats.temperatureCount }}</div>
              </div>
            </el-col>
            <el-col :span="12">
              <div class="hero-kpi-card glass">
                <div class="stat-label">运输节点数</div>
                <div class="stat-value">{{ stats.transportCount }}</div>
              </div>
            </el-col>
          </el-row>
        </div>
      </div>
    </div>

    <!-- Trace selector + charts (data panels) -->
    <div class="data-panels section-gap">
      <el-row :gutter="20" align="middle">
        <el-col :span="24">
          <div class="trace-select-row glass">
            <span class="trace-select-label">选择追溯码：</span>
            <el-select
              v-model="selectedTraceId"
              placeholder="请选择追溯码"
              clearable
              filterable
              style="width: 420px"
              @change="onTraceIdChange"
            >
              <el-option
                v-for="tid in traceIdList"
                :key="tid"
                :label="formatTraceIdOption(tid)"
                :value="tid"
              ></el-option>
            </el-select>
          </div>
        </el-col>
      </el-row>

      <el-row :gutter="20" class="section-gap">
        <el-col :span="12">
          <el-card shadow="hover" class="glass">
            <div slot="header" class="section-title">温度变化趋势</div>
            <div v-if="!selectedTraceId" class="chart-placeholder">请选择追溯码</div>
            <div v-else id="temperatureChart" style="height: 300px"></div>
          </el-card>
        </el-col>
        <el-col :span="12">
          <el-card shadow="hover" class="glass">
            <div slot="header" class="section-title">湿度变化趋势</div>
            <div v-if="!selectedTraceId" class="chart-placeholder">请选择追溯码</div>
            <div v-else id="humidityChart" style="height: 300px"></div>
          </el-card>
        </el-col>
      </el-row>

      <el-row :gutter="20" class="section-gap">
        <el-col :span="24">
          <el-card shadow="hover" class="glass">
            <div slot="header" class="section-title">物流轨迹</div>
            <div v-if="!selectedTraceId" class="chart-placeholder chart-placeholder-wide">
              请选择追溯码
            </div>
            <div v-else id="mapChart" style="height: 320px"></div>
          </el-card>
        </el-col>
      </el-row>
    </div>

    <!-- Feature highlights strip -->
    <el-row :gutter="20" class="section-gap feature-strip">
      <el-col :span="6" v-if="role === 'producer'">
        <div class="feature-card glass" @click="$router.push('/products')">
          <div class="feature-icon">
            <i class="el-icon-goods"></i>
          </div>
          <div class="feature-content">
            <div class="feature-title">商品上架与管理</div>
            <div class="feature-desc">
              通过 POST /api/products 创建商品，追溯码由后端自动生成并展示。
            </div>
          </div>
        </div>
      </el-col>
      <el-col :span="6">
        <div class="feature-card glass" @click="$router.push('/trace')">
          <div class="feature-icon">
            <i class="el-icon-search"></i>
          </div>
          <div class="feature-content">
            <div class="feature-title">追溯码全链路查询</div>
            <div class="feature-desc">
              基于 GET /api/products/trace/:trace_id 展示商品生命周期、温控和运输信息。
            </div>
          </div>
        </div>
      </el-col>
      <el-col :span="6" v-if="role === 'regulator'">
        <div class="feature-card glass" @click="$router.push('/regulator')">
          <div class="feature-icon">
            <i class="el-icon-view"></i>
          </div>
          <div class="feature-content">
            <div class="feature-title">区块链审计日志</div>
            <div class="feature-desc">
              查看 ProductHistory 审计轨迹，含交易哈希、区块号与字段变更详情。
            </div>
          </div>
        </div>
      </el-col>
      <el-col :span="6" v-if="role === 'warehouse' || role === 'logistics' || role === 'regulator'">
        <div
          class="feature-card glass"
          @click="$router.push(role === 'regulator' ? '/regulator-alerts' : '/alerts')"
        >
          <div class="feature-icon">
            <i class="el-icon-warning"></i>
          </div>
          <div class="feature-content">
            <div class="feature-title">告警与监控</div>
            <div class="feature-desc">
              基于 /api/alerts 与 /api/regulator/alerts 的温控与全局告警监控，看板式展示异常。
            </div>
          </div>
        </div>
      </el-col>
    </el-row>

    <!-- Regulator monitoring snippet -->
    <el-row
      v-if="role === 'regulator'"
      :gutter="20"
      class="section-gap regulator-board"
    >
      <el-col :span="24">
        <el-card shadow="hover" class="glass">
          <div slot="header" class="section-title">
            监管监控概览
          </div>
          <div class="regulator-summary">
            <div class="regulator-summary-item">
              <span class="label">当前异常告警：</span>
              <span :class="['value', stats.abnormalCount > 0 ? 'danger' : 'ok']">
                {{ stats.abnormalCount }}
              </span>
            </div>
            <div class="regulator-summary-item">
              <span class="label">区块链连接状态：</span>
              <span
                :class="[
                  'value',
                  stats.blockchainConnected ? 'ok' : 'danger'
                ]"
              >
                {{ stats.blockchainConnected ? '已连接' : '未连接' }}
              </span>
            </div>
            <div class="regulator-actions">
              <el-button size="small" type="primary" @click="$router.push('/regulator')">
                查看全部审计日志
              </el-button>
              <el-button size="small" @click="$router.push('/regulator-alerts')">
                查看全局告警
              </el-button>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import api from '../api'
import * as echarts from 'echarts'
import { formatDateTimeNoTZ } from '../utils/dateFormat'

export default {
  name: 'Dashboard',
  data() {
    return {
      stats: {
        totalProducts: 0,
        blockchainRecords: 0,
        abnormalCount: 0,
        todayQueries: 0,
        temperatureCount: 0,
        transportCount: 0,
        blockchainConnected: false
      },
      statsLoading: false,
      temperatureChart: null,
      humidityChart: null,
      mapChart: null,
      trendData: [],
      transportNodes: [],
      selectedTraceId: '',
      statsPollTimer: null
    }
  },
  computed: {
    role() {
      const user = this.$store.state.user || {}
      return user.role || ''
    },
    traceIdList() {
      const set = new Set()
      this.trendData.forEach(r => {
        const tid = r.trace_id || ''
        if (tid) set.add(tid)
      })
      this.transportNodes.forEach(n => {
        const tid = n.trace_id || ''
        if (tid) set.add(tid)
      })
      return [...set]
    }
  },
  mounted() {
    this.loadStats()
    this.loadChartsData()
    this.statsPollTimer = setInterval(() => this.loadStats(), 15000)
  },
  beforeDestroy() {
    if (this.statsPollTimer) clearInterval(this.statsPollTimer)
    if (this.temperatureChart) this.temperatureChart.dispose()
    if (this.humidityChart) this.humidityChart.dispose()
    if (this.mapChart) this.mapChart.dispose()
  },
  methods: {
    formatChartTimeNoTZ(val) {
      return formatDateTimeNoTZ(val) || ''
    },
    async loadStats() {
      this.statsLoading = true
      try {
        const res = await api.get('/dashboard/stats')
        const d = res.data.data || {}
        this.stats.totalProducts = d.total_products ?? 0
        this.stats.blockchainRecords = d.blockchain_records ?? 0
        this.stats.abnormalCount = d.abnormal_count ?? 0
        this.stats.todayQueries = d.today_queries ?? 0
        this.stats.temperatureCount = d.temperature_count ?? 0
        this.stats.transportCount = d.transport_count ?? 0
        this.stats.blockchainConnected = !!d.blockchain_connected
      } catch (e) {
        this.$message.error(e.response?.data?.error || '加载统计失败')
      } finally {
        this.statsLoading = false
      }
    },
    async loadChartsData() {
      try {
        const [trendRes, mapRes] = await Promise.all([
          api.get('/dashboard/temperature-trend'),
          api.get('/dashboard/transport-map')
        ])
        const trendData = trendRes.data.data || []
        const mapData = mapRes.data.data || []
        this.trendData = trendData
        this.transportNodes = mapData
        this.selectedTraceId = ''
        this.$nextTick(() => {
          this.refreshChartsBySelection()
        })
      } catch (e) {
        this.$message.error(e.response?.data?.error || '加载图表数据失败')
        this.trendData = []
        this.transportNodes = []
        this.selectedTraceId = ''
        this.$nextTick(() => this.refreshChartsBySelection())
      }
    },
    formatTraceIdOption(tid) {
      if (!tid) return '-'
      return tid.length > 36 ? tid.slice(0, 32) + '...' : tid
    },
    onTraceIdChange() {
      this.$nextTick(() => this.refreshChartsBySelection())
    },
    refreshChartsBySelection() {
      if (!this.selectedTraceId) {
        this.showPlaceholderInCharts()
        return
      }
      this.$nextTick(() => {
        const trendFiltered = this.trendData.filter(r => r.trace_id === this.selectedTraceId)
        const nodesFiltered = this.transportNodes.filter(n => n.trace_id === this.selectedTraceId)
        this.initTemperatureChart(trendFiltered)
        this.initHumidityChart(trendFiltered)
        this.initTransportChart(nodesFiltered)
      })
    },
    showPlaceholderInCharts() {
      if (this.temperatureChart) {
        this.temperatureChart.dispose()
        this.temperatureChart = null
      }
      if (this.humidityChart) {
        this.humidityChart.dispose()
        this.humidityChart = null
      }
      if (this.mapChart) {
        this.mapChart.dispose()
        this.mapChart = null
      }
    },
    initTemperatureChart(records) {
      const el = document.getElementById('temperatureChart')
      if (!el) return
      if (this.temperatureChart) this.temperatureChart.dispose()
      this.temperatureChart = echarts.init(el)

      const times = records.map(r => this.formatChartTimeNoTZ(r.record_time))
      const temperatures = records.map(r => {
        const v = Number(r.temperature)
        return typeof v === 'number' && !Number.isNaN(v) ? v : 0
      })
      const option = {
        tooltip: { trigger: 'axis' },
        grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
        xAxis: { type: 'category', data: times, axisLabel: { rotate: 30 } },
        yAxis: { type: 'value', name: '温度(℃)' },
        series: [{ name: '温度', type: 'line', data: temperatures, smooth: true }]
      }
      if (times.length === 0) option.title = { text: '暂无温控记录', left: 'center', top: 'middle' }
      this.temperatureChart.setOption(option, true)
    },
    initHumidityChart(records) {
      const el = document.getElementById('humidityChart')
      if (!el) return
      if (this.humidityChart) this.humidityChart.dispose()
      this.humidityChart = echarts.init(el)

      const times = records.map(r => this.formatChartTimeNoTZ(r.record_time))
      const humidities = records.map(r => {
        const v = Number(r.humidity)
        return typeof v === 'number' && !Number.isNaN(v) ? v : 0
      })
      const option = {
        tooltip: { trigger: 'axis' },
        grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
        xAxis: { type: 'category', data: times, axisLabel: { rotate: 30 } },
        yAxis: { type: 'value', name: '湿度(%)' },
        series: [{ name: '湿度', type: 'line', data: humidities, smooth: true }]
      }
      if (times.length === 0) option.title = { text: '暂无湿度记录', left: 'center', top: 'middle' }
      this.humidityChart.setOption(option, true)
    },
    initTransportChart(nodes) {
      const el = document.getElementById('mapChart')
      if (!el) return
      if (this.mapChart) this.mapChart.dispose()
      this.mapChart = echarts.init(el)

      if (!nodes || nodes.length === 0) {
        const option = {
          title: {
            text: this.selectedTraceId ? '该追溯码暂无运输节点' : '请选择追溯码查看物流轨迹',
            left: 'center',
            top: 'middle'
          }
        }
        this.mapChart.setOption(option, true)
        return
      }

      const list = [...nodes].sort((a, b) => new Date(a.arrival_time) - new Date(b.arrival_time))
      const graphNodes = []
      const links = []
      for (let i = 0; i < list.length; i++) {
        const n = list[i]
        const label = `${n.node_name || '节点'} (${n.location || '-'})`
        const id = `node-${i}-${label}`
        graphNodes.push({ id, name: label, value: 1 })
        if (i < list.length - 1) {
          const next = list[i + 1]
          const nextLabel = `${next.node_name || '节点'} (${next.location || '-'})`
          const nextId = `node-${i + 1}-${nextLabel}`
          links.push({ source: id, target: nextId })
        }
      }

      const step = 100
      const option = {
        tooltip: {},
        series: [{
          type: 'graph',
          layout: 'none',
          symbolSize: 40,
          roam: true,
          label: { show: true },
          edgeSymbol: ['circle', 'arrow'],
          edgeSymbolSize: [4, 8],
          data: graphNodes.map((n, i) => ({
            ...n,
            x: (i % 4) * step + 80,
            y: Math.floor(i / 4) * 80 + 80
          })),
          links: links,
          lineStyle: { opacity: 0.6, curveness: 0.2 }
        }]
      }
      this.mapChart.setOption(option, true)
    }
  }
}
</script>

<style scoped>
.dashboard-container {
  padding: 0;
}

.dashboard-hero {
  position: relative;
  overflow: hidden;
  border-radius: 12px;
  padding: 20px 24px;
  background: linear-gradient(135deg, #e0f2fe 0%, #eff6ff 40%, #dbeafe 100%);
  margin-bottom: var(--section-gap);
}

.hero-bg {
  position: absolute;
  inset: 0;
  opacity: 0.5;
  pointer-events: none;
}

.hero-content {
  position: relative;
  display: flex;
  justify-content: space-between;
  align-items: stretch;
  gap: 24px;
}

.hero-left {
  flex: 1.1;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.hero-right {
  flex: 1;
}

.hero-title {
  font-size: var(--font-size-page-title);
  font-weight: 700;
  color: #0f172a;
  margin-bottom: 8px;
}

.hero-subtitle {
  font-size: 0.95rem;
  color: #475569;
  max-width: 520px;
}

.hero-meta {
  margin-top: 16px;
  display: flex;
  flex-wrap: wrap;
  gap: 12px 24px;
  font-size: 0.85rem;
  color: #334155;
}

.hero-meta-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}

.hero-meta-item i {
  font-size: 14px;
}

.hero-badge {
  display: inline-flex;
  align-items: center;
  padding: 2px 8px;
  border-radius: 999px;
  font-size: 0.8rem;
  margin-left: 4px;
}

.hero-badge-success {
  background: rgba(34, 197, 94, 0.12);
  color: #15803d;
}

.hero-badge-danger {
  background: rgba(239, 68, 68, 0.12);
  color: #b91c1c;
}

.hero-kpi-card {
  padding: 10px 12px;
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.3);
  min-height: 72px;
}

.glass {
  background: var(--bg-card) !important;
  backdrop-filter: blur(10px);
}

.section-title {
  font-size: var(--font-size-section);
}

.stat-item {
  text-align: center;
}

.stat-value {
  font-size: var(--font-size-card-number);
  font-weight: bold;
  color: var(--color-primary);
}

.stat-label {
  margin-top: var(--spacing-unit);
  color: var(--text-secondary);
  font-size: var(--font-size-body);
}

.blockchain-status {
  padding: 12px 16px;
  background: var(--bg-card);
  border-radius: 4px;
  font-size: 0.95rem;
  border: 1px solid #ebeef5;
}
.blockchain-status .label {
  color: var(--text-regular);
}
.blockchain-status .value.connected {
  color: var(--color-success);
  font-weight: 500;
}
.blockchain-status .value.disconnected {
  color: var(--color-danger);
  font-weight: 500;
}

.trace-select-row {
  padding: 12px 16px;
  background: var(--bg-card);
  border-radius: 4px;
  font-size: var(--font-size-body);
  display: flex;
  align-items: center;
  gap: 12px;
  border: 1px solid #ebeef5;
}
.trace-select-label {
  color: var(--text-regular);
  font-weight: 500;
}

.data-panels {
  margin-top: 8px;
}

.chart-placeholder {
  height: 300px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-secondary);
  font-size: var(--font-size-body);
  background: var(--bg-main);
  border-radius: 4px;
}
.chart-placeholder-wide {
  height: 320px;
}

.feature-strip {
  margin-top: var(--section-gap);
}

.feature-card {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 14px 16px;
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.3);
  cursor: pointer;
  transition: box-shadow 0.2s ease, transform 0.1s ease, border-color 0.2s ease;
}

.feature-card:hover {
  box-shadow: 0 10px 24px rgba(15, 23, 42, 0.12);
  transform: translateY(-1px);
  border-color: rgba(59, 130, 246, 0.6);
}

.feature-icon {
  width: 40px;
  height: 40px;
  border-radius: 999px;
  background: linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
  box-shadow: 0 10px 18px rgba(59, 130, 246, 0.35);
}

.feature-icon i {
  font-size: 18px;
}

.feature-content {
  flex: 1;
}

.feature-title {
  font-weight: 600;
  font-size: 0.95rem;
  color: #0f172a;
  margin-bottom: 4px;
}

.feature-desc {
  font-size: 0.85rem;
  color: #6b7280;
  line-height: 1.4;
}

:deep(.app-theme-dark) .dashboard-hero {
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0b1220 100%);
}

:deep(.app-theme-dark) .hero-title {
  color: #e5e7eb;
}

:deep(.app-theme-dark) .hero-subtitle,
:deep(.app-theme-dark) .hero-meta {
  color: #cbd5e1;
}

:deep(.app-theme-dark) .hero-kpi-card {
  border-color: #334155;
  background: #1f2937 !important;
}

:deep(.app-theme-dark) .feature-card {
  border-color: #334155;
  background: #1f2937 !important;
}

:deep(.app-theme-dark) .feature-title {
  color: #e5e7eb;
}

:deep(.app-theme-dark) .feature-desc {
  color: #cbd5e1;
}

:deep(.app-theme-dark) .trace-select-row {
  border-color: #334155;
}

.regulator-board {
  margin-bottom: 8px;
}

.regulator-summary {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px 24px;
  font-size: 0.9rem;
}

.regulator-summary-item .label {
  color: var(--text-regular);
}

.regulator-summary-item .value.ok {
  color: var(--color-success);
  font-weight: 500;
}

.regulator-summary-item .value.danger {
  color: var(--color-danger);
  font-weight: 500;
}

.regulator-actions {
  margin-left: auto;
  display: flex;
  gap: 8px;
}

@media (max-width: 1024px) {
  .hero-content {
    flex-direction: column;
  }
  .hero-right {
    width: 100%;
  }
  .regulator-summary {
    flex-direction: column;
    align-items: flex-start;
  }
  .regulator-actions {
    margin-left: 0;
  }
}
</style>
