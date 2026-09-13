<template>
  <div class="products-container">
    <el-card>
      <div slot="header">
        <span>商品信息管理</span>
        <el-button style="float: right; margin-left: 8px" @click="refreshProducts">刷新</el-button>
        <el-button type="primary" style="float: right" @click="openCreateDialog">新增商品</el-button>
      </div>

      <el-form :inline="true" class="search-form">
        <el-form-item label="追溯码">
          <el-input v-model="searchTraceId" placeholder="追溯码" clearable style="width: 200px"></el-input>
        </el-form-item>
        <el-form-item label="批次号">
          <el-input v-model="searchBatch" placeholder="批次号" clearable style="width: 160px"></el-input>
        </el-form-item>
        <el-form-item label="原产地">
          <el-input v-model="searchOrigin" placeholder="原产地" clearable style="width: 140px"></el-input>
        </el-form-item>
      </el-form>

      <el-table :data="paginatedProducts" v-loading="loading">
        <el-table-column prop="product_id" label="商品编号" width="140"></el-table-column>
        <el-table-column prop="batch_number" label="批次号" width="120"></el-table-column>
        <el-table-column label="追溯码" width="260">
          <template slot-scope="scope">
            <TraceCodeCell :value="scope.row.trace_id" :short="true" :max-len="16" />
          </template>
        </el-table-column>
        <el-table-column prop="origin" label="原产地" min-width="100"></el-table-column>
        <el-table-column prop="production_time" label="生产时间" width="180">
          <template slot-scope="scope">{{ scope.row.production_time | formatDateTimeTZ }}</template>
        </el-table-column>
        <el-table-column prop="consignee" label="收货人" width="120"></el-table-column>
        <el-table-column prop="destination" label="目的地" min-width="90"></el-table-column>
        <el-table-column label="区块链状态" width="100" align="center">
          <template slot-scope="scope">
            <el-tag :type="scope.row.blockchain_tx_hash ? 'success' : 'info'" size="small">
              {{ scope.row.blockchain_tx_hash ? '已上链' : '待上链' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template slot-scope="scope">
            <el-button size="mini" @click="editProduct(scope.row)">编辑</el-button>
            <el-button size="mini" @click="openHistoryDrawer(scope.row)">历史</el-button>
            <el-button size="mini" type="danger" @click="deleteProduct(scope.row.id)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="pagination"
        :current-page="currentPage"
        :page-sizes="[10, 20, 50, 100]"
        :page-size="pageSize"
        :total="filteredProducts.length"
        layout="total, sizes, prev, pager, next"
        @size-change="pageSize = $event"
        @current-change="currentPage = $event"
      ></el-pagination>

      <el-dialog :title="dialogTitle" :visible.sync="showDialog" width="620px">
        <el-form :model="productForm" label-width="120px">
          <el-form-item label="商品编号">
            <el-input v-model="productForm.product_id"></el-input>
          </el-form-item>
          <el-form-item label="批次号">
            <el-input v-model="productForm.batch_number"></el-input>
          </el-form-item>
          <el-form-item label="原产地">
            <el-input v-model="productForm.origin"></el-input>
          </el-form-item>
          <el-form-item label="生产时间">
            <el-date-picker v-model="productForm.production_time" type="datetime" style="width: 100%"></el-date-picker>
          </el-form-item>
          <el-form-item label="最低温度(℃)">
            <el-input-number v-model="productForm.temperature_min"></el-input-number>
          </el-form-item>
          <el-form-item label="最高温度(℃)">
            <el-input-number v-model="productForm.temperature_max"></el-input-number>
          </el-form-item>
          <el-form-item label="最低湿度(%)">
            <el-input-number v-model="productForm.humidity_min" :min="0" :max="100"></el-input-number>
          </el-form-item>
          <el-form-item label="最高湿度(%)">
            <el-input-number v-model="productForm.humidity_max" :min="0" :max="100"></el-input-number>
          </el-form-item>
          <el-form-item label="运输公司">
            <el-input v-model="productForm.transport_company"></el-input>
          </el-form-item>
          <el-form-item label="收货人">
            <el-input v-model="productForm.consignee" placeholder="填写消费者账号用户名"></el-input>
          </el-form-item>
          <el-form-item label="目的地">
            <el-input v-model="productForm.destination"></el-input>
          </el-form-item>
          <el-form-item label="海关信息">
            <el-input type="textarea" v-model="productForm.customs_info"></el-input>
          </el-form-item>
        </el-form>
        <span slot="footer">
          <el-button @click="showDialog = false">取消</el-button>
          <el-button type="primary" :loading="saveLoading" @click="saveProduct">保存</el-button>
        </span>
      </el-dialog>

      <TraceDetailDrawer
        :visible.sync="historyDrawerVisible"
        :trace-id="historyTraceId"
        title="商品变更历史"
        item-label="历史记录"
        :item-count="historyList.length"
        summary="按时间查看该追溯码关联的商品创建、更新与删除审计记录，便于核对操作人与变更说明。"
      >
        <el-table :data="historyList" border size="small" v-loading="historyLoading">
          <el-table-column prop="operation_type" label="操作" width="100">
            <template slot-scope="scope">
              <el-tag :type="historyOpTagType(scope.row.operation_type)" size="mini">
                {{ historyOpLabel(scope.row.operation_type) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="operator_name" label="操作人" width="100"></el-table-column>
          <el-table-column prop="change_details" label="变更说明" min-width="200" show-overflow-tooltip></el-table-column>
          <el-table-column prop="created_at" label="时间" width="170">
            <template slot-scope="scope">{{ scope.row.created_at | formatDateTimeTZ }}</template>
          </el-table-column>
        </el-table>
      </TraceDetailDrawer>
    </el-card>
  </div>
</template>

<script>
import api from '../api'
import TraceCodeCell from '../components/TraceCodeCell.vue'
import TraceDetailDrawer from '../components/TraceDetailDrawer.vue'

export default {
  name: 'Products',
  components: { TraceCodeCell, TraceDetailDrawer },
  data() {
    return {
      products: [],
      loading: false,
      showDialog: false,
      dialogTitle: '新增商品',
      searchTraceId: '',
      searchBatch: '',
      searchOrigin: '',
      currentPage: 1,
      pageSize: 10,
      productForm: {
        product_id: '',
        batch_number: '',
        origin: '',
        production_time: '',
        temperature_min: -18,
        temperature_max: -2,
        humidity_min: 0,
        humidity_max: 100,
        transport_company: '',
        consignee: '',
        destination: '',
        customs_info: ''
      },
      originalForm: null,
      editingId: null,
      historyDrawerVisible: false,
      historyTraceId: '',
      historyList: [],
      historyLoading: false,
      saveLoading: false,
      chainWatchJobs: {},
      blockchainPollTimer: null,
      blockchainDisconnectedAlertShown: false,
      lastBlockchainConnected: null
    }
  },
  computed: {
    filteredProducts() {
      let list = this.products || []
      if (this.searchTraceId) {
        const s = this.searchTraceId.trim().toLowerCase()
        list = list.filter(p => (p.trace_id || '').toLowerCase().includes(s))
      }
      if (this.searchBatch) {
        const s = this.searchBatch.trim().toLowerCase()
        list = list.filter(p => (p.batch_number || '').toLowerCase().includes(s))
      }
      if (this.searchOrigin) {
        const s = this.searchOrigin.trim().toLowerCase()
        list = list.filter(p => (p.origin || '').toLowerCase().includes(s))
      }
      return list
    },
    paginatedProducts() {
      const start = (this.currentPage - 1) * this.pageSize
      return this.filteredProducts.slice(start, start + this.pageSize)
    }
  },
  mounted() {
    this.loadProducts()
    this.checkBlockchainStatus()
    this.blockchainPollTimer = setInterval(() => this.checkBlockchainStatus(), 8000)
  },
  beforeDestroy() {
    if (this.blockchainPollTimer) {
      clearInterval(this.blockchainPollTimer)
      this.blockchainPollTimer = null
    }
    Object.keys(this.chainWatchJobs).forEach(id => this.stopWatchOnChain(id))
  },
  methods: {
    async refreshProducts(silent = false) {
      await this.loadProducts()
      if (!silent) this.$message.success('列表已刷新')
    },
    async loadProducts(silent = false) {
      if (!silent) this.loading = true
      try {
        const response = await api.get('/products')
        this.products = response.data.data || []
      } catch (error) {
        if (!silent) {
          this.$message.error(error.response?.data?.error || '获取商品列表失败')
        }
      } finally {
        if (!silent) this.loading = false
      }
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
              '当前无法连接区块链节点，商品将保持「待上链」；节点恢复后系统会自动补录上链。请检查链配置与网络。',
              '区块链未连接',
              { type: 'warning', confirmButtonText: '知道了' }
            )
          }
        } else {
          this.blockchainDisconnectedAlertShown = false
          const hasPending = (this.products || []).some(p => !p.blockchain_tx_hash)
          if (wasDown || hasPending) {
            await this.loadProducts(true)
          }
        }
      } catch (_) {
        /* 未登录或网络异常时不反复弹窗 */
      }
    },
    validateTemperatureHumidity() {
      const tmin = Number(this.productForm.temperature_min)
      const tmax = Number(this.productForm.temperature_max)
      const hmin = Number(this.productForm.humidity_min)
      const hmax = Number(this.productForm.humidity_max)
      if (tmax <= tmin) {
        this.$message.error('商品温度参数错误：最高温度必须大于最低温度')
        return false
      }
      if (hmax <= hmin) {
        this.$message.error('商品湿度参数错误：最高湿度必须大于最低湿度')
        return false
      }
      return true
    },
    async saveProduct() {
      if (!this.validateTemperatureHumidity()) return
      this.saveLoading = true
      let watchProductID = null
      try {
        if (this.editingId) {
          const editableKeys = [
            'product_id', 'batch_number', 'origin', 'production_time',
            'temperature_min', 'temperature_max', 'humidity_min', 'humidity_max',
            'transport_company', 'consignee', 'destination', 'customs_info'
          ]
          const diff = {}
          const normalize = v => (v === null || v === undefined) ? '' : String(v)
          for (const k of editableKeys) {
            const oldVal = normalize(this.originalForm ? this.originalForm[k] : undefined)
            const newVal = normalize(this.productForm[k])
            if (oldVal !== newVal) {
              diff[k] = this.productForm[k]
            }
          }
          if (Object.keys(diff).length === 0) {
            this.$message.info('没有字段发生变更')
            this.showDialog = false
            return
          }
          await api.put(`/products/${this.editingId}`, diff)
          this.$message.success('更新成功')
          watchProductID = this.editingId
        } else {
          const res = await api.post('/products', this.productForm)
          this.$message.success('创建成功。链上存证异步进行中，请稍后刷新查看状态')
          watchProductID = res?.data?.data?.id || null
        }
        this.showDialog = false
        this.resetForm()
        await this.loadProducts()
        if (watchProductID) {
          this.startWatchOnChain(watchProductID)
        }
      } catch (error) {
        const msg = error.response?.data?.error || '操作失败'
        this.$message.error(msg)
      } finally {
        this.saveLoading = false
      }
    },
    editProduct(row) {
      this.editingId = row.id
      this.dialogTitle = '编辑商品'
      const editableKeys = [
        'product_id', 'batch_number', 'origin', 'production_time',
        'temperature_min', 'temperature_max', 'humidity_min', 'humidity_max',
        'transport_company', 'consignee', 'destination', 'customs_info'
      ]
      const snapshot = {}
      for (const k of editableKeys) {
        snapshot[k] = row[k] !== undefined ? row[k] : (k.includes('humidity') ? (k === 'humidity_min' ? 0 : 100) : '')
      }
      this.productForm = { ...snapshot }
      this.originalForm = { ...snapshot }
      this.showDialog = true
    },
    async deleteProduct(id) {
      try {
        await this.$confirm('确定要删除该商品吗？', '提示', { type: 'warning' })
        await api.delete(`/products/${id}`)
        this.$message.success('删除成功')
        this.loadProducts()
      } catch (error) {
        if (error !== 'cancel') {
          this.$message.error('删除失败')
        }
      }
    },
    resetForm() {
      this.productForm = {
        product_id: '',
        batch_number: '',
        origin: '',
        production_time: '',
        temperature_min: -18,
        temperature_max: -2,
        humidity_min: 0,
        humidity_max: 100,
        transport_company: '',
        consignee: '',
        destination: '',
        customs_info: ''
      }
      this.originalForm = null
      this.editingId = null
      this.dialogTitle = '新增商品'
    },
    openCreateDialog() {
      this.resetForm()
      this.showDialog = true
    },
    historyOpLabel(type) {
      if (type === 'create') return '创建'
      if (type === 'update') return '更新'
      if (type === 'delete') return '删除'
      return type || '—'
    },
    historyOpTagType(type) {
      if (type === 'create') return 'success'
      if (type === 'update') return 'warning'
      if (type === 'delete') return 'danger'
      return 'info'
    },
    async openHistoryDrawer(row) {
      this.historyTraceId = row.trace_id || ''
      this.historyDrawerVisible = true
      this.historyList = []
      if (!this.historyTraceId) return
      this.historyLoading = true
      try {
        const res = await api.get(`/products/history/${this.historyTraceId}`)
        this.historyList = res.data.data || []
      } catch (e) {
        this.$message.error(e.response?.data?.error || '加载历史失败')
      } finally {
        this.historyLoading = false
      }
    },
    stopWatchOnChain(productID) {
      const key = String(productID)
      if (this.chainWatchJobs[key]) {
        clearInterval(this.chainWatchJobs[key])
        delete this.chainWatchJobs[key]
      }
    },
    startWatchOnChain(productID) {
      if (!productID) return
      const key = String(productID)
      this.stopWatchOnChain(key)

      let inFlight = false
      let attempts = 0
      const maxAttempts = 20 // 最长约 60s
      let seenNoHash = false
      const tick = async () => {
        if (inFlight) return
        inFlight = true
        attempts += 1
        try {
          await this.refreshProducts(true)
          const list = this.products || []
          const item = list.find(p => String(p.id) === key)
          if (item && !item.blockchain_tx_hash) {
            seenNoHash = true
          }
          if (item && item.blockchain_tx_hash) {
            this.stopWatchOnChain(key)
            if (seenNoHash) {
              this.$message.success(`检测到商品「${item.product_id || item.trace_id}」交易哈希已生成，已自动刷新`)
            }
            return
          }
          if (attempts >= maxAttempts) {
            this.stopWatchOnChain(key)
            this.$message.info('上链仍在处理中，请稍后手动刷新查看')
          }
        } catch (_) {
          if (attempts >= maxAttempts) {
            this.stopWatchOnChain(key)
          }
        } finally {
          inFlight = false
        }
      }

      this.chainWatchJobs[key] = setInterval(tick, 3000)
      tick()
    }
  }
}
</script>

<style scoped>
.products-container {
  padding: 20px;
}
.search-form {
  margin-bottom: 16px;
}
.pagination {
  margin-top: 16px;
  text-align: right;
}
</style>
