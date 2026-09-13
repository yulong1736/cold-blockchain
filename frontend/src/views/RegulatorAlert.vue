<template>
  <div class="page-card-wrap">
    <el-card>
      <div slot="header" class="card-header">
        <span>全局温控告警</span>
        <el-tag type="danger" size="small" style="margin-left: 8px">{{ refreshIntervalText }}</el-tag>
      </div>

      <el-tabs v-model="activeTab">
        <el-tab-pane label="按追溯码汇总" name="summary">
          <el-table :data="traceSummaryList" v-loading="loading" border :row-class-name="alertRowClassName">
            <el-table-column prop="trace_id_short" label="追溯码" width="220">
              <template slot-scope="scope">
                <TraceCodeCell :value="scope.row.trace_id" :short="true" :max-len="16" />
              </template>
            </el-table-column>
            <el-table-column prop="count" label="告警次数" width="120" align="center"></el-table-column>
            <el-table-column prop="latest_time" label="最新告警时间" width="240">
              <template slot-scope="scope">{{ scope.row.latest_time | formatDateTimeTZ }}</template>
            </el-table-column>
            <el-table-column label="操作" width="120" fixed="right">
              <template slot-scope="scope">
                <el-button type="text" size="small" @click="openTraceDetail(scope.row.trace_id)">查看明细</el-button>
              </template>
            </el-table-column>
          </el-table>
          <div v-if="traceSummaryList.length === 0 && !loading" class="empty-tip">暂无告警</div>
        </el-tab-pane>
        <el-tab-pane label="全部告警" name="all">
          <el-table :data="alerts" v-loading="loading" border :row-class-name="alertRowClassName">
            <el-table-column prop="record_time" label="告警时间" width="240">
              <template slot-scope="scope">{{ scope.row.record_time | formatDateTimeTZ }}</template>
            </el-table-column>
            <el-table-column prop="reason" label="告警原因" width="120">
              <template slot-scope="scope">
                <el-tag type="danger" effect="plain" size="small">
                  {{ scope.row.reason || '温控异常' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="location" label="告警地点"></el-table-column>
            <el-table-column label="追溯码" width="240">
              <template slot-scope="scope">
                <TraceCodeCell :value="scope.row.trace_id" :short="true" :max-len="16" />
              </template>
            </el-table-column>
            <el-table-column prop="warehouse_name" label="记录人" width="140"></el-table-column>
            <el-table-column prop="temperature" label="温度(℃)" width="100"></el-table-column>
            <el-table-column prop="humidity" label="湿度(%)" width="100"></el-table-column>
          </el-table>
          <div v-if="alerts.length === 0 && !loading" class="empty-tip">暂无告警</div>
        </el-tab-pane>
      </el-tabs>

      <TraceDetailDrawer
        :visible.sync="detailDrawerVisible"
        :trace-id="detailTraceId"
        title="该追溯码下告警记录"
        item-label="告警记录"
        :item-count="detailAlerts.length"
        summary="面向监管视角汇总该追溯码的全部告警详情，便于核对记录人、发生位置与环境异常趋势。"
      >
        <el-table :data="detailAlerts" border size="small" :row-class-name="alertRowClassName">
          <el-table-column prop="record_time" label="告警时间" width="220">
            <template slot-scope="scope">{{ scope.row.record_time | formatDateTimeTZ }}</template>
          </el-table-column>
          <el-table-column prop="reason" label="告警原因" width="100">
            <template slot-scope="scope">
              <el-tag type="danger" effect="plain" size="mini">
                {{ scope.row.reason || '温控异常' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="location" label="告警地点"></el-table-column>
          <el-table-column prop="warehouse_name" label="记录人" width="120"></el-table-column>
          <el-table-column prop="temperature" label="温度(℃)" width="90"></el-table-column>
          <el-table-column prop="humidity" label="湿度(%)" width="90"></el-table-column>
        </el-table>
      </TraceDetailDrawer>
    </el-card>
  </div>
</template>

<script>
import TraceDetailDrawer from '../components/TraceDetailDrawer.vue'
import TraceCodeCell from '../components/TraceCodeCell.vue'
import traceAlertMixin from '../mixins/traceAlertMixin'

const POLL_INTERVAL_MS = 5000

export default {
  name: 'RegulatorAlert',
  mixins: [traceAlertMixin],
  components: { TraceDetailDrawer, TraceCodeCell },
  methods: {
    alertRowClassName() {
      return 'row-abnormal'
    },
    getAlertConfig() {
      return {
        endpoint: '/regulator/alerts',
        limit: 200,
        pollIntervalMs: POLL_INTERVAL_MS,
        silentPollError: true
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
.empty-tip {
  text-align: center;
  color: #909399;
  padding: 20px;
}
</style>
