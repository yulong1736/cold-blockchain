<template>
  <div class="page-card-wrap">
    <el-card>
      <div slot="header" class="card-header">
        <span>运输节点管理</span>
        <div class="header-right">
          <span v-if="companyName" class="company-badge">
            <i class="el-icon-office-building"></i> {{ companyName }}
          </span>
          <el-button @click="refreshList">刷新</el-button>
          <el-button type="primary" @click="openAddDialog">添加节点</el-button>
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
            <el-table-column prop="count" label="运输节点数" width="120" align="center"></el-table-column>
            <el-table-column prop="latest_arrival" label="最新到达时间" width="240">
              <template slot-scope="scope">{{ scope.row.latest_arrival | formatDateTimeTZ }}</template>
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
          <div v-if="traceSummaryList.length === 0 && !loading" class="empty-tip">暂无运输节点</div>
        </el-tab-pane>
        <el-tab-pane label="全部记录" name="all">
          <el-table :data="nodes" v-loading="loading">
            <el-table-column label="追溯码" width="280">
              <template slot-scope="scope">
                <TraceCodeCell :value="scope.row.trace_id" :short="true" :max-len="20" />
              </template>
            </el-table-column>
            <el-table-column prop="node_name" label="节点名称"></el-table-column>
            <el-table-column prop="location" label="位置"></el-table-column>
            <el-table-column prop="arrival_time" label="到达时间" width="240" :formatter="(row)=>formatDate(row.arrival_time)"></el-table-column>
            <el-table-column prop="departure_time" label="离开时间" width="240" :formatter="(row)=>formatDate(row.departure_time)"></el-table-column>
            <el-table-column prop="status" label="状态">
              <template slot-scope="scope">
                <el-tag :type="statusType(scope.row.status)" :class="statusClass(scope.row.status)">{{ statusLabel(scope.row.status) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="区块链状态" width="100" align="center">
              <template slot-scope="scope">
                <el-tag :type="scope.row.blockchain_tx_hash ? 'success' : 'info'" size="small">
                  {{ scope.row.blockchain_tx_hash ? '已上链' : '待上链' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template slot-scope="scope">
                <el-button type="text" size="small" @click="openTimeDialog(scope.row)">时间</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-tab-pane>
      </el-tabs>

      <TraceDetailDrawer
        :visible.sync="detailDrawerVisible"
        :trace-id="detailTraceId"
        title="该追溯码下运输节点"
        item-label="运输节点"
        :item-count="detailNodes.length"
        summary="集中查看该追溯码经过的运输节点、到离站时间、当前位置与链上状态，便于核对物流轨迹。"
      >
        <el-table :data="detailNodes" border size="small">
          <el-table-column prop="node_name" label="节点名称" width="120"></el-table-column>
          <el-table-column prop="location" label="位置"></el-table-column>
          <el-table-column prop="arrival_time" label="到达时间" width="220" :formatter="(row)=>formatDate(row.arrival_time)"></el-table-column>
          <el-table-column prop="departure_time" label="离开时间" width="220" :formatter="(row)=>formatDate(row.departure_time)"></el-table-column>
          <el-table-column prop="status" label="状态" width="100">
            <template slot-scope="scope">
              <el-tag :type="statusType(scope.row.status)" :class="statusClass(scope.row.status)" size="mini">{{ statusLabel(scope.row.status) }}</el-tag>
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

      <el-dialog title="添加运输节点" :visible.sync="showDialog" width="520px">
        <el-alert
          v-if="companyName"
          :title="`当前账号所属公司：${companyName}，只能为该公司运输的商品添加节点`"
          type="info"
          show-icon
          :closable="false"
          style="margin-bottom: 16px"
        ></el-alert>
        <el-form :model="nodeForm" label-width="100px">
          <el-form-item label="追溯码">
            <el-select
              v-model="nodeForm.trace_id"
              placeholder="请选择商品追溯码"
              filterable
              style="width: 100%"
              :loading="productsLoading"
              no-data-text="暂无该公司关联商品"
            >
              <el-option
                v-for="p in myProducts"
                :key="p.trace_id"
                :label="formatProductLabel(p)"
                :value="p.trace_id"
              ></el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="节点名称">
            <el-input v-model="nodeForm.node_name"></el-input>
          </el-form-item>
          <el-form-item label="位置">
            <el-input v-model="nodeForm.location"></el-input>
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="nodeForm.status">
              <el-option label="运输中" value="in_transit"></el-option>
              <el-option label="已到达" value="arrived"></el-option>
              <el-option label="已交付" value="delivered"></el-option>
            </el-select>
          </el-form-item>
          <el-form-item label="离开时间">
            <el-date-picker
              v-model="nodeForm.departure_time"
              type="datetime"
              placeholder="可选，离开本节点时间"
              style="width: 100%"
              value-format="yyyy-MM-dd HH:mm:ss"
            ></el-date-picker>
          </el-form-item>
        </el-form>
        <div slot="footer">
          <el-button @click="showDialog = false">取消</el-button>
          <el-button type="primary" @click="saveNode">保存</el-button>
        </div>
      </el-dialog>

      <el-dialog title="编辑运输时间" :visible.sync="timeDialogVisible" width="480px" @close="timeEditRow = null">
        <el-form v-if="timeEditRow" label-width="100px">
          <el-form-item label="节点">
            {{ timeEditRow.node_name }}（{{ shortTrace(timeEditRow.trace_id) }}）
          </el-form-item>
          <el-form-item label="到达时间" required>
            <el-date-picker
              v-model="timeForm.arrival_time"
              type="datetime"
              placeholder="选择到达时间"
              style="width: 100%"
              value-format="yyyy-MM-dd HH:mm:ss"
            ></el-date-picker>
          </el-form-item>
          <el-form-item label="离开时间" required>
            <el-date-picker
              v-model="timeForm.departure_time"
              type="datetime"
              placeholder="选择离开时间"
              style="width: 100%"
              value-format="yyyy-MM-dd HH:mm:ss"
            ></el-date-picker>
          </el-form-item>
        </el-form>
        <div slot="footer">
          <el-button @click="timeDialogVisible = false">取消</el-button>
          <el-button type="primary" :loading="timeSaving" @click="saveTimes">保存</el-button>
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
import { mapGetters } from 'vuex'
import { formatDateTimeWithTZ } from '../utils/dateFormat'

const STATUS_MAP = {
  in_transit: { type: 'primary', label: '运输中' },
  arrived: { type: 'success', label: '已到达' },
  delivered: { type: '', label: '已交付', class: 'status-delivered' }
}

export default {
  name: 'Transport',
  components: { TraceDetailDrawer, TraceCodeCell },
  data() {
    return {
      activeTab: 'summary',
      nodes: [],
      loading: false,
      showDialog: false,
      detailDrawerVisible: false,
      detailTraceId: '',
      myProducts: [],
      productsLoading: false,
      nodeForm: {
        trace_id: '',
        node_name: '',
        location: '',
        status: 'in_transit',
        departure_time: null
      },
      blockchainPollTimer: null,
      blockchainDisconnectedAlertShown: false,
      lastBlockchainConnected: null,
      chainWatchJobs: {},
      chainTxNotified: {},
      timeDialogVisible: false,
      timeEditRow: null,
      timeForm: { arrival_time: null, departure_time: null },
      timeSaving: false
    }
  },
  computed: {
    ...mapGetters(['companyName']),
    traceSummaryList() {
      return buildTraceSummary(this.nodes, { timeField: 'arrival_time', withBlockchain: true })
    },
    detailNodes() {
      if (!this.detailTraceId) return []
      return this.nodes
        .filter(n => n.trace_id === this.detailTraceId)
        .sort((a, b) => new Date(a.arrival_time) - new Date(b.arrival_time))
    }
  },
  mounted() {
    this.loadNodes()
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
    startWatchTransportRow(nodeId) {
      if (!nodeId) return
      const key = String(nodeId)
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
              await api.post(`/transport/${nodeId}/sync-blockchain`)
            } catch (_) {
              /* ignore */
            }
          }
          const prev = (this.nodes || []).map(n => ({
            id: n.id,
            blockchain_tx_hash: n.blockchain_tx_hash
          }))
          const res = await api.get('/transport')
          const next = unwrapData(res, [])
          this.applyChainTxNotifications(prev, next, 'transport')
          this.nodes = next
          const row = next.find(n => String(n.id) === key)
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
        if (now && !had && !this.chainTxNotified[`t-${row.id}`]) {
          this.$set(this.chainTxNotified, `t-${row.id}`, true)
          if (kind === 'transport') {
            const name = row.node_name || '节点'
            this.$message.success(`运输节点「${name}」（追溯码 ${this.shortTrace(row.trace_id)}）已上链`)
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
              '当前无法连接区块链节点，运输节点将保持「待上链」；节点恢复后请刷新列表查看交易哈希。',
              '区块链未连接',
              { type: 'warning', confirmButtonText: '知道了' }
            )
          }
        } else {
          this.blockchainDisconnectedAlertShown = false
          const hasPending = (this.nodes || []).some(n => !n.blockchain_tx_hash)
          if (wasDown || hasPending) {
            await this.loadNodes(true)
          }
        }
      } catch (_) {
        /* 未登录或网络异常 */
      }
    },
    formatDate(val) {
      return formatDateTimeWithTZ(val) || '-'
    },
    statusType(s) {
      return (STATUS_MAP[s] && STATUS_MAP[s].type) || ''
    },
    statusLabel(s) {
      return (STATUS_MAP[s] && STATUS_MAP[s].label) || s
    },
    statusClass(s) {
      return (STATUS_MAP[s] && STATUS_MAP[s].class) || ''
    },
    openTraceDetail(traceId) {
      this.detailTraceId = traceId
      this.detailDrawerVisible = true
    },
    formatProductLabel(p) {
      const short = p.trace_id ? p.trace_id.slice(0, 16) + '...' : '-'
      return `${short}（${p.product_id || ''} ${p.origin || ''}→${p.destination || ''}）`
    },
    openTimeDialog(row) {
      this.timeEditRow = row
      this.timeForm.arrival_time = row.arrival_time ? this.formatDatePickerValue(row.arrival_time) : null
      this.timeForm.departure_time = row.departure_time ? this.formatDatePickerValue(row.departure_time) : null
      this.timeDialogVisible = true
    },
    formatDatePickerValue(val) {
      if (!val) return null
      const d = new Date(val)
      if (Number.isNaN(d.getTime())) return null
      const pad = n => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
    },
    async saveTimes() {
      if (!this.timeEditRow || !this.timeForm.arrival_time || !this.timeForm.departure_time) {
        this.$message.warning('请选择到达和离开时间')
        return
      }
      if (new Date(this.timeForm.departure_time) < new Date(this.timeForm.arrival_time)) {
        this.$message.warning('离开时间不能早于到达时间')
        return
      }
      const rowId = this.timeEditRow.id
      this.timeSaving = true
      try {
        await api.put(`/transport/${rowId}`, {
          arrival_time: `${this.timeForm.arrival_time.replace(' ', 'T')}+08:00`,
          departure_time: `${this.timeForm.departure_time.replace(' ', 'T')}+08:00`
        })
        this.$message.success('运输时间已更新，存证将重新上链')
        this.timeDialogVisible = false
        await this.loadNodes()
        this.startWatchTransportRow(rowId)
      } catch (e) {
        this.$message.error(getErrorMessage(e, '保存失败'))
      } finally {
        this.timeSaving = false
      }
    },
    async openAddDialog() {
      this.showDialog = true
      this.nodeForm = {
        trace_id: '',
        node_name: '',
        location: '',
        status: 'in_transit',
        departure_time: null
      }
      this.productsLoading = true
      try {
        const res = await api.get('/transport/my-products')
        this.myProducts = unwrapData(res, [])
      } catch (e) {
        this.$message.error(getErrorMessage(e, '获取关联商品失败'))
        this.myProducts = []
      } finally {
        this.productsLoading = false
      }
    },
    async loadNodes(silent = false) {
      const prev =
        silent && (this.nodes || []).length
          ? (this.nodes || []).map(n => ({
              id: n.id,
              blockchain_tx_hash: n.blockchain_tx_hash
            }))
          : null
      if (!silent) this.loading = true
      try {
        const res = await api.get('/transport')
        const next = unwrapData(res, [])
        if (silent && prev) {
          this.applyChainTxNotifications(prev, next, 'transport')
        }
        this.nodes = next
      } catch (e) {
        if (!silent) this.$message.error(getErrorMessage(e, '加载运输节点失败'))
      } finally {
        if (!silent) this.loading = false
      }
    },
    async refreshList() {
      await this.loadNodes()
      this.$message.success('列表已刷新')
    },
    async saveNode() {
      if (!this.nodeForm.trace_id) {
        this.$message.warning('请选择商品追溯码')
        return
      }
      if (!this.nodeForm.node_name) {
        this.$message.warning('请填写节点名称')
        return
      }
      try {
        const payload = {
          trace_id: this.nodeForm.trace_id,
          node_name: this.nodeForm.node_name,
          location: this.nodeForm.location,
          status: this.nodeForm.status
        }
        if (this.nodeForm.departure_time) {
          payload.departure_time = `${this.nodeForm.departure_time.replace(' ', 'T')}+08:00`
        }
        const res = await api.post('/transport', payload)
        this.$message.success('保存成功。链上存证异步进行中，列表将自动刷新状态')
        const newId = res?.data?.data?.id
        this.showDialog = false
        this.nodeForm = {
          trace_id: '',
          node_name: '',
          location: '',
          status: 'in_transit',
          departure_time: null
        }
        await this.loadNodes()
        if (newId) {
          this.startWatchTransportRow(newId)
        }
      } catch (e) {
        this.$message.error(getErrorMessage(e, '保存失败'))
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
.header-right {
  display: flex;
  align-items: center;
  gap: 12px;
}
.company-badge {
  font-size: 13px;
  color: #409eff;
  background: #ecf5ff;
  border: 1px solid #b3d8ff;
  padding: 2px 10px;
  border-radius: 12px;
}
.empty-tip {
  text-align: center;
  color: #909399;
  padding: 20px;
}
.status-delivered {
  background-color: #9c27b0;
  border-color: #9c27b0;
  color: #fff;
}
</style>
