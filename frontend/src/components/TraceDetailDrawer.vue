<template>
  <el-drawer
    :visible.sync="visibleSync"
    size="60%"
    direction="rtl"
    :before-close="handleClose"
    :with-header="false"
    custom-class="trace-detail-drawer-shell"
  >
    <div class="trace-detail-drawer" :class="{ 'trace-detail-drawer--empty': !traceId }">
      <div class="drawer-hero">
        <div class="drawer-hero__eyebrow">
          <i class="el-icon-document-checked"></i>
          <span>Trace Detail</span>
        </div>
        <div class="drawer-hero__top">
          <div class="drawer-hero__title">
            <h2>{{ title }}</h2>
            <p>{{ summaryText }}</p>
          </div>
          <div v-if="itemCount !== null" class="drawer-count-pill">
            {{ itemCount }} 条{{ itemLabel }}
          </div>
        </div>
        <div v-if="traceId" class="trace-meta-card">
          <span class="trace-meta-card__label">追溯码</span>
          <TraceCodeCell :value="traceId" :short="false" />
        </div>
      </div>

      <section v-if="traceId" class="drawer-content-card">
        <slot></slot>
      </section>

      <div v-else class="drawer-empty-state">
        <i class="el-icon-document-delete"></i>
        <p>暂无可展示的追溯明细</p>
      </div>
    </div>
  </el-drawer>
</template>

<script>
import TraceCodeCell from './TraceCodeCell.vue'

export default {
  name: 'TraceDetailDrawer',
  components: { TraceCodeCell },
  props: {
    visible: { type: Boolean, default: false },
    traceId: { type: String, default: '' },
    title: { type: String, default: '该追溯码下记录' },
    itemLabel: { type: String, default: '记录' },
    itemCount: { type: Number, default: null },
    summary: { type: String, default: '' }
  },
  computed: {
    visibleSync: {
      get() {
        return this.visible
      },
      set(v) {
        this.$emit('update:visible', v)
      }
    },
    summaryText() {
      if (this.summary) return this.summary
      return `集中查看该追溯码关联的${this.itemLabel}，便于快速核对状态、时间与链上同步情况。`
    }
  },
  methods: {
    handleClose() {
      this.$emit('update:visible', false)
    }
  }
}
</script>

<style scoped>
.trace-detail-drawer {
  min-height: 100%;
  background: linear-gradient(180deg, #eef5ff 0%, #f7faff 32%, #f4f7fb 100%);
  padding: 28px 24px 24px;
}

.trace-detail-drawer--empty {
  display: flex;
  flex-direction: column;
}

.drawer-hero {
  position: relative;
  padding: 24px;
  border-radius: 22px;
  background: linear-gradient(135deg, rgba(37, 99, 235, 0.96) 0%, rgba(59, 130, 246, 0.92) 60%, rgba(96, 165, 250, 0.88) 100%);
  color: #ffffff;
  box-shadow: 0 18px 40px rgba(37, 99, 235, 0.22);
  overflow: hidden;
}

.drawer-hero::after {
  content: '';
  position: absolute;
  inset: auto -72px -90px auto;
  width: 220px;
  height: 220px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.14);
}

.drawer-hero__eyebrow {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 7px 12px;
  border-radius: 999px;
  background: rgba(255, 255, 255, 0.16);
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.drawer-hero__top {
  position: relative;
  z-index: 1;
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  margin-top: 18px;
}

.drawer-hero__title h2 {
  margin: 0;
  font-size: 28px;
  line-height: 1.2;
  font-weight: 800;
}

.drawer-hero__title p {
  margin: 10px 0 0;
  max-width: 540px;
  font-size: 14px;
  line-height: 1.7;
  color: rgba(255, 255, 255, 0.88);
}

.drawer-count-pill {
  position: relative;
  z-index: 1;
  flex-shrink: 0;
  padding: 10px 14px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.18);
  border: 1px solid rgba(255, 255, 255, 0.24);
  font-size: 13px;
  font-weight: 700;
}

.trace-meta-card {
  position: relative;
  z-index: 1;
  margin-top: 18px;
  padding: 16px 18px;
  border-radius: 16px;
  background: rgba(255, 255, 255, 0.12);
  border: 1px solid rgba(255, 255, 255, 0.16);
}

.trace-meta-card__label {
  display: block;
  margin-bottom: 8px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.04em;
  color: rgba(255, 255, 255, 0.74);
}

:deep(.trace-meta-card .trace-copy-wrap) {
  display: inline-flex;
  align-items: center;
  gap: 10px;
}

:deep(.trace-meta-card .trace-text) {
  color: #ffffff;
  font-weight: 700;
  letter-spacing: 0.01em;
}

:deep(.trace-meta-card .copy-btn.el-button) {
  min-width: 34px;
  height: 34px;
  padding: 0 10px;
  border-radius: 10px;
  border: 1px solid rgba(255, 255, 255, 0.18);
  background: linear-gradient(135deg, #f59e0b 0%, #f97316 100%);
  color: #ffffff;
  box-shadow: 0 10px 18px rgba(249, 115, 22, 0.28);
  transition: transform 0.2s ease, box-shadow 0.2s ease, filter 0.2s ease;
}

:deep(.trace-meta-card .copy-btn.el-button:hover),
:deep(.trace-meta-card .copy-btn.el-button:focus) {
  color: #ffffff;
  transform: translateY(-1px);
  filter: brightness(1.04);
  box-shadow: 0 12px 22px rgba(249, 115, 22, 0.34);
}

.drawer-content-card {
  margin-top: 18px;
  padding: 18px;
  border-radius: 22px;
  background: rgba(255, 255, 255, 0.94);
  border: 1px solid #dbe7f3;
  box-shadow: 0 18px 36px rgba(15, 23, 42, 0.06);
}

.drawer-empty-state {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-direction: column;
  gap: 10px;
  color: #6b7b93;
  min-height: 280px;
}

.drawer-empty-state i {
  font-size: 28px;
  color: #8cabd4;
}

:deep(.trace-detail-drawer-shell) {
  background: #f4f7fb;
}

:deep(.trace-detail-drawer-shell .el-drawer__body) {
  padding: 0;
  overflow-y: auto;
}

:deep(.drawer-content-card .el-table) {
  border-radius: 16px;
  overflow: hidden;
}

:deep(.drawer-content-card .el-table th) {
  background: #f5f8fc;
  color: #4c5f79;
  font-weight: 700;
}

:deep(.drawer-content-card .el-table td) {
  color: #24364d;
}

:deep(.drawer-content-card .el-tag) {
  border-radius: 999px;
  padding: 0 10px;
}

@media (max-width: 1024px) {
  .trace-detail-drawer {
    padding: 18px 16px 16px;
  }

  .drawer-hero {
    padding: 18px;
    border-radius: 18px;
  }

  .drawer-hero__top {
    flex-direction: column;
  }

  .drawer-hero__title h2 {
    font-size: 24px;
  }
}
</style>
