<template>
  <div class="regulator-container">
    <el-row :gutter="20" class="stats-row" v-loading="statsLoading">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value">{{ stats.totalProducts }}</div>
            <div class="stat-label">商品总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value">{{ stats.blockchainRecords }}</div>
            <div class="stat-label">链上记录数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value">{{ stats.abnormalCount }}</div>
            <div class="stat-label">异常告警</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value">{{ stats.todayQueries }}</div>
            <div class="stat-label">今日查询</div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="20" class="section-gap stats-row stat-card-hover">
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value">{{ stats.temperatureCount }}</div>
            <div class="stat-label">温控记录数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover">
          <div class="stat-item">
            <div class="stat-value">{{ stats.transportCount }}</div>
            <div class="stat-label">运输节点数</div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="20" class="section-gap">
      <el-col :span="24">
        <div class="blockchain-status">
          <span class="label">区块链状态：</span>
          <span :class="['value', stats.blockchainConnected ? 'connected' : 'disconnected']">
            {{ stats.blockchainConnected ? '已连接' : '未连接' }}
          </span>
        </div>
      </el-col>
    </el-row>

    <el-card class="section-gap">
      <div slot="header">
        <span>监管审计</span>
      </div>

      <el-form :inline="true">
        <el-form-item label="追溯码">
          <el-input v-model="traceID" placeholder="可选，按追溯码筛选" clearable style="width: 280px"></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="loadAuditLogs">查询</el-button>
        </el-form-item>
      </el-form>

      <el-divider></el-divider>

      <h3>审计日志</h3>
      <p v-if="loading" class="loading-tip">加载中…</p>
      <el-table v-else :data="auditLogs" border stripe row-key="id" :default-expand-all="false">
        <el-table-column type="expand">
          <template slot-scope="props">
            <div class="expand-detail">
              <p><strong>区块链交易哈希：</strong>{{ props.row.blockchain_tx_hash || '—' }}</p>
              <p><strong>区块链哈希：</strong>{{ props.row.blockchain_hash || '—' }}</p>
              <p><strong>旧哈希：</strong>{{ props.row.old_hash || '—' }}</p>
              <p><strong>新哈希：</strong>{{ props.row.new_hash || '—' }}</p>
              <p><strong>变更详情：</strong>{{ props.row.change_details || '—' }}</p>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="operator_name" label="操作人" width="120"></el-table-column>
        <el-table-column prop="operation_label" label="操作" width="90"></el-table-column>
        <el-table-column label="商品追溯码" width="220">
          <template slot-scope="scope">
            <TraceCodeCell :value="scope.row.trace_id" :short="true" :max-len="16" />
          </template>
        </el-table-column>
        <el-table-column label="修改字段" width="160">
          <template slot-scope="scope">
            <span>{{ formatFieldNames(scope.row.changed_fields) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="具体变化" min-width="240" show-overflow-tooltip>
          <template slot-scope="scope">
            <span class="changed-fields-cell">{{ scope.row.changed_fields || '—' }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="时间" width="170">
          <template slot-scope="scope">{{ scope.row.created_at | formatDateTimeTZ }}</template>
        </el-table-column>
      </el-table>
      <p v-if="!loading && auditLogs.length === 0" class="empty-tip">暂无审计记录，可按追溯码筛选或留空查看全部。</p>
    </el-card>
  </div>
</template>

<script>
import api from '../api'
import TraceCodeCell from '../components/TraceCodeCell.vue'

export default {
  name: 'Regulator',
  components: { TraceCodeCell },
  data() {
    return {
      traceID: '',
      auditLogs: [],
      loading: false,
      statsLoading: false,
      stats: {
        totalProducts: 0,
        blockchainRecords: 0,
        abnormalCount: 0,
        todayQueries: 0,
        temperatureCount: 0,
        transportCount: 0,
        blockchainConnected: false
      },
      pollTimer: null
    }
  },
  mounted() {
    this.loadStats()
    this.loadAuditLogs()
    this.pollTimer = setInterval(() => {
      this.loadStats()
      this.loadAuditLogs(false)
    }, 10000)
  },
  beforeDestroy() {
    if (this.pollTimer) clearInterval(this.pollTimer)
  },
  methods: {
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
      } finally {
        this.statsLoading = false
      }
    },
    operationLabel(type) {
      const map = { create: '创建商品', update: '更新商品', delete: '删除商品' }
      return map[type] || type
    },
    formatFieldNames(kv) {
      if (!kv || kv === '—') return '—'
      const isKVFormat = kv.includes(' -> ')
      if (isKVFormat) {
        return kv.split(';').map(s => {
          const colonIdx = s.indexOf(':')
          return colonIdx > -1 ? s.slice(0, colonIdx).trim() : s.trim()
        }).filter(Boolean).join(', ')
      }
      return kv
    },
    async loadAuditLogs(showLoading = true) {
      if (showLoading) this.loading = true
      try {
        const params = { limit: 500 }
        if (this.traceID) params.trace_id = this.traceID.trim()
        const res = await api.get('/regulator/audit-logs', { params })
        const list = res.data.data || []
        this.auditLogs = list.map(item => ({
          ...item,
          operation_label: this.operationLabel(item.operation_type),
          trace_id_short: item.trace_id && item.trace_id.length > 12 ? item.trace_id.slice(0, 12) + '...' : item.trace_id
        }))
      } catch (e) {
        if (showLoading) this.$message.error(e.response?.data?.error || '加载审计日志失败')
        this.auditLogs = []
      } finally {
        if (showLoading) this.loading = false
      }
    }
  }
}
</script>

<style scoped>
.regulator-container {
  padding: 0;
}

.stats-row .stat-item {
  text-align: center;
}
.stats-row .stat-value {
  font-size: var(--font-size-card-number);
  font-weight: bold;
  color: var(--color-primary);
}
.stats-row .stat-label {
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

.changed-fields-cell {
  word-break: break-all;
  white-space: pre-wrap;
}
.expand-detail {
  padding: 12px 20px;
  background: #fafafa;
  font-size: 13px;
  line-height: 1.8;
}
.expand-detail p {
  margin: 4px 0;
  word-break: break-all;
}
</style>
