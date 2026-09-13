<template>
  <div class="page-card-wrap">
    <el-card>
      <div slot="header" class="card-header">
        <span>温控监控管理</span>
        <div class="header-actions">
          <el-button @click="refreshList">刷新</el-button>
          <el-button type="primary" @click="showDialog = true">添加记录</el-button>
        </div>
      </div>

      <el-tabs v-model="activeTab">
        <el-tab-pane label="按追溯码汇总" name="summary">
          <el-table :data="traceSummaryList" v-loading="loading" border>
            <el-table-column prop="trace_id_short" label="追溯码" width="220">
              <template slot-scope="scope">
                <TraceCodeCell :value="scope.row.trace_id" :short="true" :max-len="16" />
              </template>
            </el-table-column>
            <el-table-column prop="count" label="温控记录数" width="120" align="center"></el-table-column>
            <el-table-column prop="abnormal_count" label="异常次数" width="100" align="center">
              <template slot-scope="scope">
                <el-tag v-if="scope.row.abnormal_count > 0" type="danger" size="small">{{ scope.row.abnormal_count }}</el-tag>
                <span v-else>0</span>
              </template>
            </el-table-column>
            <el-table-column prop="latest_time" label="最新记录时间" width="240">
              <template slot-scope="scope">{{ scope.row.latest_time | formatDateTimeTZ }}</template>
            </el-table-column>
            <el-table-column label="区块链状态" width="110" align="center">
              <template slot-scope="scope">
                <el-tag :type="blockchainSummaryTagType(scope.row)" size="small">
                  {{ blockchainSummaryLabel(scope.row) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template slot-scope="scope">
                <el-button type="text" size="small" @click="openTraceDetail(scope.row.trace_id)">查看明细</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div v-if="traceSummaryList.length === 0 && !loading" class="empty-tip">暂无温控记录</div>
        </el-tab-pane>
        <el-tab-pane label="全部记录" name="all">
          <el-form :inline="true" class="filter-form">
            <el-form-item label="记录时间">
              <el-date-picker
                v-model="dateRange"
                type="daterange"
                range-separator="至"
                start-placeholder="开始日期"
                end-placeholder="结束日期"
                value-format="yyyy-MM-dd"
                style="width: 240px"
              ></el-date-picker>
            </el-form-item>
            <el-form-item>
              <el-button size="small" @click="dateRange = null">清空</el-button>
            </el-form-item>
          </el-form>
          <el-table :data="filteredRecords" v-loading="loading" :row-class-name="rowClassName">
            <el-table-column label="追溯码" width="280">
              <template slot-scope="scope">
                <TraceCodeCell :value="scope.row.trace_id" :short="true" :max-len="20" />
              </template>
            </el-table-column>
            <el-table-column prop="record_time" label="记录时间" width="240" :formatter="(row)=>formatDate(row.record_time)"></el-table-column>
            <el-table-column prop="temperature" label="温度(℃)"></el-table-column>
            <el-table-column prop="humidity" label="湿度(%)"></el-table-column>
            <el-table-column prop="location" label="位置"></el-table-column>
            <el-table-column prop="is_abnormal" label="状态">
              <template slot-scope="scope">
                <span v-if="scope.row.is_abnormal"><i class="el-icon-warning"></i> </span>
                <el-tag :type="scope.row.is_abnormal ? 'danger' : 'success'">
                  {{ scope.row.is_abnormal ? '异常' : '正常' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="区块链状态" width="100" align="center">
              <template slot-scope="scope">
                <el-tag :type="scope.row.blockchain_tx_hash ? 'success' : 'info'" size="small">
                  {{ scope.row.blockchain_tx_hash ? '已上链' : '待上链' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>

      <TraceDetailDrawer
        :visible.sync="detailDrawerVisible"
        :trace-id="detailTraceId"
        title="该追溯码下温控记录"
        item-label="温控记录"
        :item-count="detailRecords.length"
        summary="按时间倒序查看温度、湿度、异常标记与链上同步状态，便于快速判断冷链是否稳定。"
      >
        <el-table :data="detailRecords" border size="small" :row-class-name="rowClassName">
          <el-table-column prop="record_time" label="记录时间" width="220" :formatter="(row)=>formatDate(row.record_time)"></el-table-column>
          <el-table-column prop="temperature" label="温度(℃)" width="100"></el-table-column>
          <el-table-column prop="humidity" label="湿度(%)" width="100"></el-table-column>
          <el-table-column prop="location" label="位置"></el-table-column>
          <el-table-column prop="is_abnormal" label="状态" width="90">
            <template slot-scope="scope">
              <span v-if="scope.row.is_abnormal"><i class="el-icon-warning"></i> </span>
              <el-tag :type="scope.row.is_abnormal ? 'danger' : 'success'" size="mini">
                {{ scope.row.is_abnormal ? '异常' : '正常' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="链上" width="88" align="center">
            <template slot-scope="scope">
              <el-tag :type="scope.row.blockchain_tx_hash ? 'success' : 'info'" size="mini">
                {{ scope.row.blockchain_tx_hash ? '已上链' : '待上链' }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
      </TraceDetailDrawer>

      <el-dialog title="添加温控记录" :visible.sync="showDialog" width="500px">
        <el-form :model="recordForm" label-width="100px">
          <el-form-item label="追溯码">
            <el-input v-model="recordForm.trace_id" placeholder="请输入商品追溯码"></el-input>
          </el-form-item>
          <el-form-item label="温度(℃)">
            <el-input-number v-model="recordForm.temperature" :precision="1" :step="0.5"></el-input-number>
          </el-form-item>
          <el-form-item label="湿度(%)">
            <el-input-number v-model="recordForm.humidity" :min="0" :max="100" :precision="1"></el-input-number>
          </el-form-item>
          <el-form-item label="位置">
            <el-input v-model="recordForm.location" placeholder="如：仓库A / 冷库3号"></el-input>
          </el-form-item>
        </el-form>
        <div slot="footer">
          <el-button @click="showDialog = false">取消</el-button>
          <el-button type="primary" @click="saveRecord" :loading="saving">保存</el-button>
        </div>
      </el-dialog>
    </el-card>
  </div>
</template>

<script>
import api, { getErrorMessage, unwrapData } from '../api'
import TraceDetailDrawer from '../components/TraceDetailDrawer.vue'
import TraceCodeCell from '../components/TraceCodeCell.vue'
import { buildTraceSummary } from '../utils/traceSummary'
import { formatDateTimeWithTZ } from '../utils/dateFormat'

export default {
  name: 'Temperature',
  components: { TraceDetailDrawer, TraceCodeCell },
  data() {
    return {
      activeTab: 'summary',
      records: [],
      loading: false,
      saving: false,
      showDialog: false,
      detailDrawerVisible: false,
      detailTraceId: '',
      recordForm: {
        trace_id: '',
        temperature: 0,
        humidity: 0,
        location: ''
      },
      dateRange: null,
      blockchainPollTimer: null,
      blockchainDisconnectedAlertShown: false,
      lastBlockchainConnected: null,
      chainWatchJobs: {},
      chainTxNotified: {}
    }
  },
  computed: {
    filteredRecords() {
      let list = this.records || []
      if (this.dateRange && this.dateRange.length === 2) {
        const [start, end] = this.dateRange
        const startT = new Date(start).getTime()
        const endT = new Date(end + 'T23:59:59').getTime()
        list = list.filter(r => {
          const t = new Date(r.record_time).getTime()
          return t >= startT && t <= endT
        })
      }
      return list
    },
    traceSummaryList() {
      return buildTraceSummary(this.records, {
        timeField: 'record_time',
        countAbnormal: true,
        withBlockchain: true
      })
    },
    detailRecords() {
      if (!this.detailTraceId) return []
      return this.records
        .filter(r => r.trace_id === this.detailTraceId)
        .sort((a, b) => new Date(b.record_time) - new Date(a.record_time))
    }
  },
  mounted() {
    this.loadRecords()
    this.checkBlockchainStatus()
    this.blockchainPollTimer = setInterval(() => this.checkBlockchainStatus(), 8000)
  },
  beforeDestroy() {
    if (this.blockchainPollTimer) {
      clearInterval(this.blockchainPollTimer)
      this.blockchainPollTimer = null
    }
    Object.keys(this.chainWatchJobs).forEach(k => this.stopWatchOnChainRow(k))
  },
  methods: {
    shortTrace(traceId) {
      const t = traceId || ''
      return t.length > 16 ? `${t.slice(0, 16)}…` : t
    },
    stopWatchOnChainRow(key) {
      const k = String(key)
      if (this.chainWatchJobs[k]) {
        clearInterval(this.chainWatchJobs[k])
        delete this.chainWatchJobs[k]
      }
    },
    startWatchTemperatureRow(recordId) {
      if (!recordId) return
      const key = String(recordId)
      this.stopWatchOnChainRow(key)
      let attempts = 0
      let syncTried = false
      let inFlight = false
      const tick = async () => {
        if (inFlight) return
        inFlight = true
        attempts += 1
        try {
          if (attempts === 10 && !syncTried) {
            syncTried = true
            try {
              await api.post(`/temperature/${recordId}/sync-blockchain`)
            } catch (_) {
              /* 链未通或权限等，忽略 */
            }
          }
          const prev = (this.records || []).map(r => ({
            id: r.id,
            blockchain_tx_hash: r.blockchain_tx_hash
          }))
          const res = await api.get('/temperature')
          const next = unwrapData(res, [])
          this.applyChainTxNotifications(prev, next, 'temperature')
          this.records = next
          const row = next.find(r => String(r.id) === key)
          if (row && row.blockchain_tx_hash) {
            this.stopWatchOnChainRow(key)
            return
          }
          if (attempts >= 20) {
            this.stopWatchOnChainRow(key)
            this.$message.info('上链仍在处理中，请稍后手动刷新查看')
          }
        } finally {
          inFlight = false
        }
      }
      this.chainWatchJobs[key] = setInterval(tick, 3000)
      tick()
    },
    applyChainTxNotifications(prevSnap, list, kind) {
      const prevMap = {}
      ;(prevSnap || []).forEach(x => {
        prevMap[x.id] = x.blockchain_tx_hash
      })
      for (const row of list || []) {
        const had = prevMap[row.id]
        const now = row.blockchain_tx_hash
        if (now && !had && !this.chainTxNotified[`w-${row.id}`]) {
          this.$set(this.chainTxNotified, `w-${row.id}`, true)
          if (kind === 'temperature') {
            this.$message.success(`温控记录（追溯码 ${this.shortTrace(row.trace_id)}）已上链`)
          }
        }
      }
    },
    blockchainSummaryTagType(row) {
      const s = row.blockchain_stats
      if (!s || !s.total) return 'info'
      if (s.pending_tx === 0) return 'success'
      if (s.pending_tx === s.total) return 'info'
      return 'warning'
    },
    blockchainSummaryLabel(row) {
      const s = row.blockchain_stats
      if (!s || !s.total) return '—'
      if (s.pending_tx === 0) return '已上链'
      if (s.pending_tx === s.total) return '待上链'
      return '部分待上链'
    },
    async checkBlockchainStatus() {
      try {
        const res = await api.get('/blockchain/status')
        const payload = res.data?.data || {}
        const ok = !!payload.connected
        const wasDown = this.lastBlockchainConnected === false
        this.lastBlockchainConnected = ok
        if (!ok) {
          if (!this.blockchainDisconnectedAlertShown) {
            this.blockchainDisconnectedAlertShown = true
            await this.$alert(
              '当前无法连接区块链节点，温控记录将保持「待上链」；节点恢复后请刷新列表查看交易哈希。',
              '区块链未连接',
              { type: 'warning', confirmButtonText: '知道了' }
            )
          }
        } else {
          this.blockchainDisconnectedAlertShown = false
          const hasPending = (this.records || []).some(r => !r.blockchain_tx_hash)
          if (wasDown || hasPending) {
            await this.loadRecords(true)
          }
        }
      } catch (_) {
        /* 未登录或网络异常 */
      }
    },
    rowClassName({ row }) {
      return row.is_abnormal ? 'row-abnormal' : ''
    },
    formatDate(val) {
      return formatDateTimeWithTZ(val) || '-'
    },
    openTraceDetail(traceId) {
      this.detailTraceId = traceId
      this.detailDrawerVisible = true
    },
    async loadRecords(silent = false) {
      const prev =
        silent && (this.records || []).length
          ? (this.records || []).map(r => ({
              id: r.id,
              blockchain_tx_hash: r.blockchain_tx_hash
            }))
          : null
      if (!silent) this.loading = true
      try {
        const res = await api.get('/temperature')
        const next = unwrapData(res, [])
        if (silent && prev) {
          this.applyChainTxNotifications(prev, next, 'temperature')
        }
        this.records = next
      } catch (e) {
        if (!silent) this.$message.error(getErrorMessage(e, '加载温控记录失败'))
      } finally {
        if (!silent) this.loading = false
      }
    },
    async refreshList() {
      await this.loadRecords()
      this.$message.success('列表已刷新')
    },
    async saveRecord() {
      if (!this.recordForm.trace_id) {
        this.$message.warning('请输入追溯码')
        return
      }
      this.saving = true
      try {
        const res = await api.post('/temperature', this.recordForm)
        this.$message.success('保存成功。链上存证异步进行中，列表将自动刷新状态')
        const newId = res?.data?.data?.id
        this.showDialog = false
        this.recordForm = { trace_id: '', temperature: 0, humidity: 0, location: '' }
        await this.loadRecords()
        if (newId) {
          this.startWatchTemperatureRow(newId)
        }
      } catch (e) {
        this.$message.error(getErrorMessage(e, '保存失败'))
      } finally {
        this.saving = false
      }
    }
  }
}
</script>

<style scoped>
.page-card-wrap {
  padding: 20px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 8px;
}
.filter-form {
  margin-bottom: 12px;
}
.empty-tip {
  text-align: center;
  color: #909399;
  padding: 20px;
}
</style>
