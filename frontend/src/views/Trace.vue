<template>
  <div class="trace-container">
    <el-card>
      <div slot="header">
        <span>商品追溯查询</span>
      </div>

      <el-form :inline="true" @submit.native.prevent>
        <el-form-item v-if="isConsumer" label="我的商品">
          <el-select
            v-model="traceID"
            placeholder="请选择你购买的商品追溯码"
            filterable
            clearable
            style="width: 420px"
            :loading="myProductsLoading"
          >
            <el-option
              v-for="p in myProducts"
              :key="p.trace_id"
              :label="formatMyProductOption(p)"
              :value="p.trace_id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="追溯码">
          <el-input
            v-model="traceID"
            :placeholder="isConsumer ? '请选择或输入你购买商品的追溯码' : '请输入追溯码'"
            style="width: 400px"
          ></el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="queryTrace" :loading="loading">查询</el-button>
        </el-form-item>
      </el-form>

      <el-divider></el-divider>

      <div v-if="productInfo" class="trace-result">
        <el-card class="trace-section-card">
          <h3 class="section-title">商品基本信息</h3>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="商品编号">{{ productInfo.product_id }}</el-descriptions-item>
            <el-descriptions-item label="批次号">{{ productInfo.batch_number }}</el-descriptions-item>
            <el-descriptions-item label="追溯码">
              <TraceCodeCell :value="productInfo.trace_id" :short="false" />
            </el-descriptions-item>
            <el-descriptions-item label="原产地">{{ productInfo.origin }}</el-descriptions-item>
            <el-descriptions-item label="生产时间">{{ productInfo.production_time | formatDateTimeTZ }}</el-descriptions-item>
            <el-descriptions-item label="目的地">{{ productInfo.destination }}</el-descriptions-item>
            <el-descriptions-item label="生产商">{{ productInfo.producer_name }}</el-descriptions-item>
            <el-descriptions-item label="运输公司">{{ productInfo.transport_company }}</el-descriptions-item>
            <el-descriptions-item label="区块链哈希" :span="2">
              <span class="blockchain-hash">{{ productInfo.blockchain_hash || '—' }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="区块号">
              <el-tag v-if="productInfo.block_number != null" size="small" type="info">{{ productInfo.block_number }}</el-tag>
              <span v-else>—</span>
            </el-descriptions-item>
            <el-descriptions-item label="交易哈希">
              <span v-if="productInfo.blockchain_tx_hash" class="blockchain-tx-row">
                <span class="tx-hash blockchain-hash" :title="productInfo.blockchain_tx_hash">{{ productInfo.blockchain_tx_hash }}</span>
                <el-button type="text" size="mini" icon="el-icon-document-copy" @click="copyTxHash(productInfo.blockchain_tx_hash)" title="复制">复制</el-button>
              </span>
              <span v-else>—</span>
            </el-descriptions-item>
          </el-descriptions>
        </el-card>

        <el-card class="trace-section-card" style="margin-top: 20px" v-if="temperatureRecords.length > 0">
          <h3 class="section-title">温控记录</h3>
          <el-table :data="temperatureRecords" border :row-class-name="tableRowClassName">
            <el-table-column prop="record_time" label="记录时间" width="220">
              <template slot-scope="scope">{{ scope.row.record_time | formatDateTimeTZ }}</template>
            </el-table-column>
            <el-table-column prop="temperature" label="温度(℃)" width="100"></el-table-column>
            <el-table-column prop="humidity" label="湿度(%)" width="100"></el-table-column>
            <el-table-column prop="location" label="位置"></el-table-column>
            <el-table-column prop="warehouse_name" label="记录人" width="120"></el-table-column>
            <el-table-column prop="is_abnormal" label="状态" width="80">
              <template slot-scope="scope">
                <span v-if="scope.row.is_abnormal"><i class="el-icon-warning"></i> </span>
                <el-tag :type="scope.row.is_abnormal ? 'danger' : 'success'" size="small">
                  {{ scope.row.is_abnormal ? '异常' : '正常' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="存证哈希" min-width="200">
              <template slot-scope="scope">
                <span class="blockchain-hash" :title="scope.row.blockchain_hash">{{ scope.row.blockchain_hash || '—' }}</span>
              </template>
            </el-table-column>
            <el-table-column label="链上交易" min-width="200">
              <template slot-scope="scope">
                <span v-if="scope.row.blockchain_tx_hash" class="tx-hash blockchain-hash" :title="scope.row.blockchain_tx_hash">{{ scope.row.blockchain_tx_hash }}</span>
                <span v-else>—</span>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <el-card class="trace-section-card" style="margin-top: 20px" v-if="transportNodes.length > 0">
          <h3 class="section-title">物流轨迹</h3>
          <el-steps :active="transportNodes.length" align-center finish-status="success" style="margin-bottom: 20px">
            <el-step
              v-for="(node, index) in transportNodes"
              :key="index"
              :title="node.node_name"
              :description="node.location"
            ></el-step>
          </el-steps>
          <el-table :data="transportNodes" border>
            <el-table-column prop="node_name" label="节点名称" width="150"></el-table-column>
            <el-table-column prop="location" label="位置"></el-table-column>
            <el-table-column prop="arrival_time" label="到达时间" width="220">
              <template slot-scope="scope">{{ scope.row.arrival_time | formatDateTimeTZ }}</template>
            </el-table-column>
            <el-table-column prop="logistics_name" label="物流员" width="120"></el-table-column>
            <el-table-column prop="status" label="状态" width="100">
              <template slot-scope="scope">
                <el-tag :type="statusType(scope.row.status)" :class="scope.row.status === 'delivered' ? 'status-delivered' : ''" size="small">
                  {{ statusLabel(scope.row.status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="存证哈希" min-width="200">
              <template slot-scope="scope">
                <span class="blockchain-hash" :title="scope.row.blockchain_hash">{{ scope.row.blockchain_hash || '—' }}</span>
              </template>
            </el-table-column>
            <el-table-column label="链上交易" min-width="200">
              <template slot-scope="scope">
                <span v-if="scope.row.blockchain_tx_hash" class="tx-hash blockchain-hash" :title="scope.row.blockchain_tx_hash">{{ scope.row.blockchain_tx_hash }}</span>
                <span v-else>—</span>
              </template>
            </el-table-column>
          </el-table>
        </el-card>

        <el-card class="trace-section-card" style="margin-top: 20px" v-if="histories.length > 0">
          <h3 class="section-title">操作历史</h3>
          <div class="history-text-list">
            <div v-for="(history, index) in histories" :key="index" class="history-text-item">
              <pre class="history-block">{{ formatHistoryBlock(history) }}</pre>
            </div>
          </div>
        </el-card>
      </div>
    </el-card>
  </div>
</template>

<script>
import api from '../api'
import TraceCodeCell from '../components/TraceCodeCell.vue'

export default {
  name: 'Trace',
  components: { TraceCodeCell },
  data() {
    return {
      traceID: '',
      loading: false,
      myProductsLoading: false,
      myProducts: [],
      productInfo: null,
      temperatureRecords: [],
      transportNodes: [],
      histories: []
    }
  },
  computed: {
    isConsumer() {
      return this.$store.getters.userRole === 'consumer'
    }
  },
  mounted() {
    if (this.isConsumer) {
      this.loadMyProducts()
    }
  },
  methods: {
    formatMyProductOption(p) {
      const trace = p.trace_id || ''
      const short = trace.length > 16 ? trace.slice(0, 16) + '...' : trace
      return `${short}（${p.product_id || ''}）`
    },
    async loadMyProducts() {
      this.myProductsLoading = true
      try {
        const res = await api.get('/consumer/products')
        this.myProducts = res.data.data || []
      } catch (e) {
        this.$message.error(e.response?.data?.error || '加载我的商品失败')
      } finally {
        this.myProductsLoading = false
      }
    },
    tableRowClassName({ row }) {
      return row.is_abnormal ? 'row-abnormal' : ''
    },
    copyTxHash(txHash) {
      if (!txHash) return
      const input = document.createElement('input')
      input.value = txHash
      document.body.appendChild(input)
      input.select()
      try {
        document.execCommand('copy')
        this.$message.success('已复制到剪贴板')
      } catch (e) {
        this.$message.error('复制失败')
      }
      document.body.removeChild(input)
    },
    formatDate(val) {
      return this.$options.filters.formatDateTimeTZ(val) || '-'
    },
    async queryTrace() {
      if (!this.traceID) {
        this.$message.warning('请输入追溯码')
        return
      }

      this.loading = true
      try {
        const tracePath = this.isConsumer
          ? `/consumer/products/trace/${this.traceID}`
          : `/products/trace/${this.traceID}`
        const historyPath = this.isConsumer
          ? `/consumer/products/history/${this.traceID}`
          : `/products/history/${this.traceID}`

        const response = await api.get(tracePath)
        this.productInfo = response.data.data.product
        this.temperatureRecords = response.data.data.temperature_records || []
        this.transportNodes = response.data.data.transport_nodes || []

        const historyResponse = await api.get(historyPath)
        this.histories = historyResponse.data.data || []

        this.$message.success('查询成功')
      } catch (error) {
        const msg = error.response?.data?.error || ''
        if (error.response?.status === 404) {
          this.$message.warning('未找到该追溯码，请核对后重试')
        } else if (msg && (msg.includes('integrity') || msg.includes('mismatch') || msg.includes('data integrity'))) {
          this.$message.error('链上数据校验未通过，可能存在篡改风险，请谨慎核对')
        } else {
          this.$message.error(msg || '查询失败，请检查追溯码是否正确')
        }
        this.productInfo = null
        this.temperatureRecords = []
        this.transportNodes = []
        this.histories = []
      } finally {
        this.loading = false
      }
    },
    statusType(status) {
      const map = { in_transit: 'primary', arrived: 'success', delivered: '' }
      return map[status] || ''
    },
    statusLabel(status) {
      const map = { in_transit: '运输中', arrived: '已到达', delivered: '已交付' }
      return map[status] || status
    },
    formatHistoryBlock(h) {
      const traceStr = h.trace_id || '—'
      let opStr = '—'
      let detailLine = ''
      if (h.operation_type === 'create') {
        opStr = '创建'
      } else if (h.operation_type === 'delete') {
        opStr = '删除'
      } else if (h.operation_type === 'update') {
        const kv = (h.changed_fields && h.changed_fields !== '—') ? h.changed_fields.trim() : ''
        if (kv) {
          const isKVFormat = kv.includes(' -> ')
          if (isKVFormat) {
            const segments = kv.split(';').map(s => s.trim()).filter(Boolean)
            const fieldNames = segments.map(s => {
              const colonIdx = s.indexOf(':')
              return colonIdx > -1 ? s.slice(0, colonIdx).trim() : s
            }).join(', ')
            opStr = fieldNames + ' 字段更新'
            detailLine = '\n具体变化：' + segments.join('；')
          } else {
            opStr = kv + ' 字段更新'
          }
        } else {
          opStr = '部分字段更新'
        }
      }
      const timeStr = this.formatDate(h.created_at) || '—'
      const hashStr = (h.old_hash && h.new_hash) ? `${h.old_hash} -> ${h.new_hash}` : (h.blockchain_hash || h.new_hash || '—')
      return `商品信息变更：\n追溯码：${traceStr}\n操作：${opStr}${detailLine}\n时间戳：${timeStr}\nblockchain_hash变化：${hashStr}`
    }
  }
}
</script>

<style scoped>
.trace-container {
  padding: 20px;
}

.trace-result {
  margin-top: 20px;
}

.history-text-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.history-text-item {
  margin-bottom: 16px;
}
.history-text-item:last-child {
  margin-bottom: 0;
}
.history-block {
  font-size: 13px;
  line-height: 1.8;
  color: #303133;
  margin: 0;
  padding: 12px 14px;
  background: #fafafa;
  border-radius: 4px;
  border: 1px solid #ebeef5;
  white-space: pre-wrap;
  word-break: break-all;
}
.section-title {
  font-size: var(--font-size-section);
  margin-bottom: 12px;
  color: var(--text-primary);
  font-weight: 600;
}
.blockchain-tx-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}
.tx-hash {
  color: var(--color-primary);
  word-break: break-all;
}
.status-delivered {
  background-color: #9c27b0;
  border-color: #9c27b0;
  color: #fff;
}

.app-theme-dark .trace-section-card {
  border: 1px solid #334155;
}

.app-theme-dark .section-title {
  color: #e5e7eb;
}

.app-theme-dark .history-block {
  color: #d1d5db;
  background: #111827;
  border-color: #334155;
}
</style>
