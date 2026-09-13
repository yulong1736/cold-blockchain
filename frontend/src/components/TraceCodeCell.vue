<template>
  <span class="trace-copy-wrap">
    <el-tooltip v-if="showTooltip && value" :content="value" placement="top">
      <span class="trace-text">{{ displayValue }}</span>
    </el-tooltip>
    <span v-else class="trace-text">{{ displayValue }}</span>
    <el-button
      v-if="value"
      type="text"
      size="mini"
      icon="el-icon-document-copy"
      class="copy-btn"
      title="复制追溯码"
      @click="copyTraceId"
    ></el-button>
  </span>
</template>

<script>
export default {
  name: 'TraceCodeCell',
  props: {
    value: { type: String, default: '' },
    short: { type: Boolean, default: false },
    maxLen: { type: Number, default: 20 },
    showTooltip: { type: Boolean, default: true }
  },
  computed: {
    displayValue() {
      if (!this.value) return '-'
      if (!this.short || this.value.length <= this.maxLen) return this.value
      return this.value.slice(0, this.maxLen) + '...'
    }
  },
  methods: {
    async copyTraceId() {
      if (!this.value) return
      try {
        if (navigator.clipboard && window.isSecureContext) {
          await navigator.clipboard.writeText(this.value)
        } else {
          const input = document.createElement('textarea')
          input.value = this.value
          input.style.position = 'fixed'
          input.style.opacity = '0'
          document.body.appendChild(input)
          input.focus()
          input.select()
          document.execCommand('copy')
          document.body.removeChild(input)
        }
        this.$message.success('追溯码已复制')
      } catch (e) {
        this.$message.error('复制失败，请手动复制')
      }
    }
  }
}
</script>

<style scoped>
.trace-copy-wrap {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  max-width: 100%;
}
.trace-text {
  word-break: break-all;
}
.copy-btn {
  padding: 0;
}
</style>
