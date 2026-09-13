<template>
  <div class="account-page page-content-area">
    <div class="account-shell">
      <aside class="account-sidebar" aria-label="账号概览">
        <div class="account-hero">
          <div class="hero-bg" aria-hidden="true"></div>
          <div class="hero-inner">
            <h1 class="hero-title">账号与安全</h1>
            <p class="hero-subtitle">
              管理登录身份、联系方式与密码，保障冷链溯源操作安全。
            </p>

            <div class="profile-card glass">
              <div class="profile-avatar" aria-hidden="true">
                <i class="el-icon-user"></i>
              </div>
              <div class="profile-body">
                <div class="profile-name">{{ displayUsername }}</div>
                <div class="profile-row">
                  <el-tag size="small" type="info" effect="plain">{{ roleLabel }}</el-tag>
                </div>
                <div class="profile-email-line" v-if="userEmail">{{ userEmail }}</div>
                <div class="profile-email-line muted" v-else>未绑定邮箱</div>
              </div>
            </div>

            <ul class="security-hints" role="list">
              <li class="hint-item">
                <i class="el-icon-lock"></i>
                <span>定期更换密码，勿与他人共用账号。</span>
              </li>
              <li class="hint-item">
                <i class="el-icon-message"></i>
                <span>邮箱用于通知与安全校验，请保持有效。</span>
              </li>
              <li class="hint-item" v-if="userRole === 'consumer'">
                <i class="el-icon-warning-outline"></i>
                <span>消费者修改用户名会同步更新关联商品的收货人展示名。</span>
              </li>
            </ul>
          </div>
        </div>
      </aside>

      <main class="account-main">
        <el-card shadow="hover" class="glass account-form-card">
          <div slot="header" class="form-card-header">
            <span class="section-title">账户设置</span>
            <span class="form-card-sub">修改将立即作用于当前会话（部分需重新登录后全局生效）。</span>
          </div>

          <div class="form-sections">
            <!-- 更改名字 -->
            <section class="form-section" aria-labelledby="sec-username">
              <h2 id="sec-username" class="form-section-title">更改用户名</h2>
              <el-form :model="nameForm" label-width="80px" label-position="top" class="account-form">
                <el-form-item label="新用户名">
                  <el-input
                    v-model="nameForm.username"
                    placeholder="请输入新用户名"
                    class="form-input-wide"
                    clearable
                  />
                </el-form-item>
                <el-button
                  type="primary"
                  size="small"
                  class="section-action"
                  @click="submitName"
                  :loading="nameLoading"
                >
                  保存
                </el-button>
              </el-form>
            </section>

            <el-divider class="section-divider" />

            <!-- 更改邮箱 -->
            <section class="form-section" aria-labelledby="sec-email">
              <h2 id="sec-email" class="form-section-title">更改邮箱</h2>
              <p class="form-section-desc">
                以下为系统当前记录的邮箱。更换时请先确认新邮箱可正常收信，保存后立即生效。
              </p>
              <el-form
                :model="emailForm"
                :rules="emailRules"
                ref="emailForm"
                label-width="80px"
                label-position="top"
                class="account-form account-form-email"
              >
                <el-form-item label="现有邮箱">
                  <el-input
                    :value="userEmail || ''"
                    placeholder="尚未绑定邮箱"
                    readonly
                    class="form-input-wide email-readonly-input"
                    aria-readonly="true"
                  >
                    <i
                      slot="suffix"
                      class="el-input__icon el-icon-lock email-lock-suffix"
                      aria-hidden="true"
                      title="系统记录，不可在此修改"
                    />
                  </el-input>
                </el-form-item>
                <el-form-item label="新邮箱" prop="email">
                  <el-input
                    v-model.trim="emailForm.email"
                    placeholder="请输入新邮箱"
                    class="form-input-wide"
                    clearable
                    autocomplete="email"
                  />
                </el-form-item>
                <el-button
                  type="primary"
                  size="small"
                  class="section-action"
                  @click="submitEmail"
                  :loading="emailLoading"
                >
                  保存
                </el-button>
              </el-form>
            </section>

            <el-divider class="section-divider" />

            <!-- 更改密码 -->
            <section class="form-section" aria-labelledby="sec-password">
              <h2 id="sec-password" class="form-section-title">更改密码</h2>
              <el-form
                :model="pwdForm"
                :rules="pwdRules"
                ref="pwdForm"
                label-width="100px"
                label-position="top"
                class="account-form"
              >
                <el-form-item label="原密码" prop="old_password">
                  <el-input
                    v-model="pwdForm.old_password"
                    type="password"
                    placeholder="请输入原密码"
                    show-password
                    class="form-input-wide"
                    autocomplete="current-password"
                  />
                </el-form-item>
                <el-form-item label="新密码" prop="new_password">
                  <el-input
                    v-model="pwdForm.new_password"
                    type="password"
                    placeholder="请输入新密码（至少6位）"
                    show-password
                    class="form-input-wide"
                    autocomplete="new-password"
                  />
                </el-form-item>
                <el-form-item label="确认新密码" prop="confirm_password">
                  <el-input
                    v-model="pwdForm.confirm_password"
                    type="password"
                    placeholder="请再次输入新密码"
                    show-password
                    class="form-input-wide"
                    autocomplete="new-password"
                  />
                </el-form-item>
                <el-button
                  type="primary"
                  size="small"
                  class="section-action"
                  @click="submitPassword"
                  :loading="pwdLoading"
                >
                  修改密码
                </el-button>
              </el-form>
            </section>

            <el-divider class="section-divider" />

            <!-- 退出 -->
            <section class="form-section form-section-logout" aria-labelledby="sec-logout">
              <h2 id="sec-logout" class="form-section-title">退出当前账号</h2>
              <p class="logout-hint">退出后需重新登录才能访问业务数据。</p>
              <el-button type="danger" plain class="section-action logout-btn" @click="handleLogout">
                退出登录
              </el-button>
            </section>
          </div>
        </el-card>
      </main>
    </div>
  </div>
</template>

<script>
import api from '../api'
import { mapActions, mapGetters } from 'vuex'

const ROLE_LABELS = {
  producer: '生产商',
  warehouse: '仓储商',
  logistics: '物流商',
  consumer: '消费者',
  regulator: '监管'
}

export default {
  name: 'Account',
  data() {
    const validateConfirm = (rule, value, callback) => {
      if (value !== this.pwdForm.new_password) {
        callback(new Error('两次输入的新密码不一致'))
      } else {
        callback()
      }
    }
    return {
      nameForm: { username: '' },
      nameLoading: false,
      emailForm: { email: '' },
      emailLoading: false,
      pwdForm: {
        old_password: '',
        new_password: '',
        confirm_password: ''
      },
      pwdLoading: false,
      emailRules: {
        email: [
          { required: true, message: '请输入新邮箱', trigger: 'blur' },
          { type: 'email', message: '邮箱格式不正确', trigger: ['blur', 'change'] }
        ]
      },
      pwdRules: {
        old_password: [{ required: true, message: '请输入原密码', trigger: 'blur' }],
        new_password: [
          { required: true, message: '请输入新密码', trigger: 'blur' },
          { min: 6, message: '新密码至少6位', trigger: 'blur' }
        ],
        confirm_password: [
          { required: true, message: '请确认新密码', trigger: 'blur' },
          { validator: validateConfirm, trigger: 'blur' }
        ]
      }
    }
  },
  computed: {
    ...mapGetters(['username', 'userRole']),
    userEmail() {
      return this.$store.state.user?.email || ''
    },
    displayUsername() {
      return (this.username && String(this.username).trim()) || '—'
    },
    roleLabel() {
      return ROLE_LABELS[this.userRole] || this.userRole || '—'
    }
  },
  mounted() {
    this.nameForm.username = this.username || ''
    this.emailForm.email = ''
  },
  methods: {
    ...mapActions(['logout']),
    async submitName() {
      if (this.userRole === 'consumer' && this.userEmail) {
        this.$message.warning('消费者改名会触发收货人字段同步，请先确认邮箱可用')
      }
      const name = (this.nameForm.username || '').trim()
      if (!name) {
        this.$message.warning('请输入新用户名')
        return
      }
      this.nameLoading = true
      try {
        const response = await api.put('/auth/profile', { username: name })
        const profile = response.data?.data || {}
        this.$store.commit('SET_USER', {
          ...this.$store.state.user,
          username: profile.username || name,
          email: profile.email || this.$store.state.user?.email || ''
        })
        this.$message.success('用户名已更新')
      } catch (e) {
        this.$message.error(e.response?.data?.error || '更新失败')
      } finally {
        this.nameLoading = false
      }
    },
    submitEmail() {
      this.$refs.emailForm.validate(async (valid) => {
        if (!valid) return
        this.emailLoading = true
        try {
          const response = await api.put('/auth/profile', { email: this.emailForm.email })
          const profile = response.data?.data || {}
          this.emailForm.email = ''
          this.$store.commit('SET_USER', {
            ...this.$store.state.user,
            username: profile.username || this.$store.state.user?.username || '',
            email: profile.email || this.emailForm.email
          })
          this.$message.success('邮箱已更新')
        } catch (e) {
          this.$message.error(e.response?.data?.error || '更新失败')
        } finally {
          this.emailLoading = false
        }
      })
    },
    submitPassword() {
      this.$refs.pwdForm.validate(async (valid) => {
        if (!valid) return
        this.pwdLoading = true
        try {
          await api.put('/auth/password', {
            old_password: this.pwdForm.old_password,
            new_password: this.pwdForm.new_password
          })
          this.$message.success('密码已更新')
          this.pwdForm = { old_password: '', new_password: '', confirm_password: '' }
          this.$refs.pwdForm.resetFields()
        } catch (e) {
          this.$message.error(e.response?.data?.error || '修改失败')
        } finally {
          this.pwdLoading = false
        }
      })
    },
    handleLogout() {
      this.logout()
      this.$router.push('/login')
    }
  }
}
</script>

<style scoped>
.account-page {
  padding-top: 0;
}

.account-shell {
  display: flex;
  align-items: flex-start;
  gap: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.account-sidebar {
  flex: 0 0 min(360px, 38%);
  min-width: 0;
}

.account-main {
  flex: 1;
  min-width: 0;
}

/* 与 Dashboard 看板 hero 一致的渐变与圆角 */
.account-hero {
  position: relative;
  overflow: hidden;
  border-radius: 12px;
  padding: 22px 24px;
  background: linear-gradient(135deg, #e0f2fe 0%, #eff6ff 40%, #dbeafe 100%);
  min-height: 100%;
}

.hero-bg {
  position: absolute;
  inset: 0;
  opacity: 0.45;
  pointer-events: none;
  background: radial-gradient(circle at 80% 20%, rgba(59, 130, 246, 0.15) 0%, transparent 50%);
}

.hero-inner {
  position: relative;
}

.hero-title {
  font-size: var(--font-size-page-title);
  font-weight: 700;
  color: #0f172a;
  margin-bottom: 8px;
  line-height: 1.2;
}

.hero-subtitle {
  font-size: 0.95rem;
  color: #475569;
  line-height: 1.55;
  margin-bottom: 20px;
}

.profile-card {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 14px 16px;
  border-radius: 10px;
  border: 1px solid rgba(148, 163, 184, 0.35);
  margin-bottom: 20px;
}

.glass {
  background: var(--bg-card) !important;
  backdrop-filter: blur(10px);
}

.profile-avatar {
  width: 48px;
  height: 48px;
  border-radius: 999px;
  background: linear-gradient(135deg, #3b82f6 0%, #60a5fa 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
  box-shadow: 0 10px 18px rgba(59, 130, 246, 0.35);
}

.profile-avatar i {
  font-size: 22px;
}

.profile-body {
  flex: 1;
  min-width: 0;
}

.profile-name {
  font-weight: 600;
  font-size: 1.05rem;
  color: #0f172a;
  word-break: break-all;
}

.profile-row {
  margin-top: 6px;
}

.profile-email-line {
  margin-top: 8px;
  font-size: 0.875rem;
  color: #334155;
  word-break: break-all;
}

.profile-email-line.muted {
  color: #64748b;
}

.security-hints {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.hint-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  font-size: 0.85rem;
  color: #334155;
  line-height: 1.45;
}

.hint-item i {
  color: var(--color-primary);
  margin-top: 2px;
  flex-shrink: 0;
}

/* 右侧表单卡片 */
.account-form-card {
  border-radius: 12px;
  border: 1px solid rgba(148, 163, 184, 0.25);
}

.form-card-header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.section-title {
  font-size: var(--font-size-section);
  font-weight: 600;
  color: var(--text-primary);
}

.form-card-sub {
  font-size: 0.8rem;
  color: var(--text-secondary);
  font-weight: 400;
}

.form-sections {
  padding-top: 4px;
}

.form-section-title {
  font-size: 0.95rem;
  font-weight: 600;
  color: #0f172a;
  margin-bottom: 14px;
}

.form-section-desc {
  font-size: 0.85rem;
  color: var(--text-secondary);
  line-height: 1.55;
  margin: -6px 0 16px;
  max-width: 520px;
}

/* 现有邮箱：外观与新邮箱一致，只读；锁图标提示不可改 */
.email-readonly-input :deep(.el-input__inner) {
  cursor: default;
}

.email-lock-suffix {
  cursor: help;
  color: var(--text-secondary);
}

.form-section-logout .form-section-title {
  margin-bottom: 8px;
}

.logout-hint {
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin-bottom: 12px;
}

.account-form :deep(.el-form-item__label) {
  color: var(--text-regular);
  font-weight: 500;
}

.form-input-wide {
  max-width: 420px;
  width: 100%;
}

.section-action {
  margin-top: 4px;
}

.logout-btn {
  cursor: pointer;
}

.section-divider {
  margin: 22px 0;
}

@media (prefers-reduced-motion: reduce) {
  .profile-card,
  .account-form-card {
    transition: none;
  }
}

@media (max-width: 1024px) {
  .account-shell {
    flex-direction: column;
  }

  .account-sidebar {
    flex: none;
    width: 100%;
  }
}

@media (max-width: 480px) {
  .hero-title {
    font-size: 1.35rem;
  }

  .form-input-wide {
    max-width: 100%;
  }
}

/* 深色主题：与 Dashboard 对齐 */
:deep(.app-theme-dark) .account-hero {
  background: linear-gradient(135deg, #0f172a 0%, #1e293b 50%, #0b1220 100%);
}

:deep(.app-theme-dark) .hero-title {
  color: #e5e7eb;
}

:deep(.app-theme-dark) .hero-subtitle {
  color: #cbd5e1;
}

:deep(.app-theme-dark) .profile-card {
  border-color: #334155;
  background: #1f2937 !important;
}

:deep(.app-theme-dark) .profile-name {
  color: #e5e7eb;
}

:deep(.app-theme-dark) .profile-email-line {
  color: #cbd5e1;
}

:deep(.app-theme-dark) .profile-email-line.muted {
  color: #94a3b8;
}

:deep(.app-theme-dark) .hint-item {
  color: #cbd5e1;
}

:deep(.app-theme-dark) .form-section-title {
  color: #e5e7eb;
}

:deep(.app-theme-dark) .form-section-desc {
  color: #94a3b8;
}

:deep(.app-theme-dark) .email-lock-suffix {
  color: #9ca3af;
}
</style>
