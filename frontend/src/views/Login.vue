<template>
  <div class="login-page">
    <div class="login-shell">
      <section class="login-visual">
        <div class="brand-row">
          <i class="el-icon-connection brand-icon"></i>
          <span class="brand-text">coldchain hub</span>
        </div>
        <div class="visual-scene" aria-hidden="true">
          <div class="scene-blob blob-1"></div>
          <div class="scene-blob blob-2"></div>
          <div class="scene-desk"></div>
          <div class="scene-screen"></div>
          <div class="scene-person person-left"></div>
          <div class="scene-person person-right"></div>
          <div class="scene-plant"></div>
        </div>
        <p class="visual-note">跨境冷链物流溯源与监管平台</p>
      </section>

      <section class="login-panel">
        <div class="panel-header">
          <h1>Welcome Back :)</h1>
          <p>请输入账号和密码，安全登录系统</p>
        </div>

        <el-alert
          v-if="loginBanner"
          :title="loginBanner"
          type="error"
          show-icon
          :closable="true"
          class="login-banner"
          @close="loginBanner = ''"
        />

        <el-form
          ref="loginForm"
          :model="loginForm"
          :rules="rules"
          label-position="top"
          class="login-form"
          @submit.native.prevent="handleLogin"
        >
          <el-form-item label="Email Address" prop="username">
            <el-input
              v-model="loginForm.username"
              placeholder="请输入用户名"
              prefix-icon="el-icon-message"
              clearable
              @input="loginBanner = ''"
              @keyup.enter.native="handleLogin"
            />
          </el-form-item>

          <el-form-item label="Password" prop="password">
            <el-input
              v-model="loginForm.password"
              type="password"
              placeholder="请输入密码"
              prefix-icon="el-icon-lock"
              show-password
              @input="loginBanner = ''"
              @keyup.enter.native="handleLogin"
            />
          </el-form-item>

          <div class="panel-meta">
            <span class="remember-text">
              <i class="el-icon-success"></i>
              Remember Me
            </span>
            <button
              type="button"
              class="forgot-trigger"
              :aria-expanded="String(forgotVisible)"
              aria-controls="forgot-password-dialog"
              @click="openForgotDialog"
            >
              Forgot Password?
            </button>
          </div>

          <el-form-item class="form-actions">
            <el-button
              type="primary"
              class="btn-login"
              :loading="loading"
              @click="handleLogin"
            >
              Login Now
            </el-button>
            <el-button class="btn-register" @click="$router.push('/register')">
              Create Account
            </el-button>
          </el-form-item>
        </el-form>
      </section>
    </div>

    <el-dialog
      title="找回密码"
      :visible.sync="forgotVisible"
      width="460px"
      class="forgot-dialog"
      :close-on-click-modal="false"
      @close="handleForgotDialogClose"
    >
      <div id="forgot-password-dialog" class="forgot-card">
        <div class="forgot-card__hero">
          <div class="forgot-card__badge">
            <i class="el-icon-message"></i>
          </div>
          <div>
            <h2>通过邮箱安全找回密码</h2>
            <p>{{ forgotSubtitle }}</p>
          </div>
        </div>

        <div class="forgot-steps" aria-hidden="true">
          <span :class="['forgot-step', 'is-active']">1. 验证邮箱</span>
          <span :class="['forgot-step', { 'is-active': codeSent }]">2. 重置密码</span>
        </div>

        <el-form
          ref="forgotFormRef"
          :model="forgotForm"
          :rules="forgotRules"
          label-position="top"
          class="forgot-form"
        >
          <el-form-item label="邮箱" prop="email">
            <el-input
              v-model.trim="forgotForm.email"
              placeholder="请输入注册邮箱"
              prefix-icon="el-icon-message"
              clearable
            >
              <template slot="append">
                <el-button
                  class="send-code-button"
                  :loading="sendCodeLoading"
                  :disabled="sendCodeLoading || sendCodeCountdown > 0"
                  @click="handleSendCode"
                >
                  {{ sendCodeButtonText }}
                </el-button>
              </template>
            </el-input>
          </el-form-item>

          <div class="forgot-note">
            <i class="el-icon-info"></i>
            <span>{{ forgotNote }}</span>
          </div>

          <transition name="fade-slide">
            <div v-if="codeSent" class="forgot-reset-fields">
              <el-alert
                type="info"
                :closable="false"
                title="验证码已发送，请在 10 分钟内完成密码重置"
                class="forgot-alert"
              />

              <el-form-item label="验证码" prop="code">
                <el-input
                  v-model.trim="forgotForm.code"
                  maxlength="6"
                  placeholder="请输入 6 位验证码"
                  clearable
                />
              </el-form-item>

              <el-form-item label="新密码" prop="new_password">
                <el-input
                  v-model="forgotForm.new_password"
                  type="password"
                  placeholder="请输入新密码（至少 6 位）"
                  show-password
                />
              </el-form-item>

              <el-form-item label="确认新密码" prop="confirm_password">
                <el-input
                  v-model="forgotForm.confirm_password"
                  type="password"
                  placeholder="请再次输入新密码"
                  show-password
                />
              </el-form-item>
            </div>
          </transition>
        </el-form>

        <div class="forgot-actions">
          <el-button class="btn-forgot-secondary" @click="forgotVisible = false">取消</el-button>
          <el-button
            type="primary"
            class="btn-forgot-primary"
            :disabled="!codeSent"
            :loading="forgotLoading"
            @click="handleResetPassword"
          >
            重置密码
          </el-button>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { mapActions } from 'vuex'
import { resetForgotPassword, sendForgotPasswordCode } from '../api'

export default {
  name: 'Login',
  data() {
    const validateForgotEmail = (rule, value, callback) => {
      const email = (value || '').trim()
      const emailPattern = /^[^\s@]+@[^\s@]+\.[^\s@]+$/
      if (!email) {
        callback(new Error('请输入邮箱'))
        return
      }
      if (!emailPattern.test(email)) {
        callback(new Error('请输入正确的邮箱格式'))
        return
      }
      callback()
    }
    const validateForgotConfirm = (rule, value, callback) => {
      if (value !== this.forgotForm.new_password) {
        callback(new Error('两次输入的新密码不一致'))
        return
      }
      callback()
    }
    return {
      loginForm: {
        username: '',
        password: ''
      },
      loginBanner: '',
      forgotVisible: false,
      forgotLoading: false,
      sendCodeLoading: false,
      sendCodeCountdown: 0,
      sendCodeTimer: null,
      codeSent: false,
      forgotForm: {
        email: '',
        code: '',
        new_password: '',
        confirm_password: ''
      },
      loading: false,
      rules: {
        username: [
          { required: true, message: '请输入用户名', trigger: 'blur' }
        ],
        password: [
          { required: true, message: '请输入密码', trigger: 'blur' }
        ]
      },
      forgotRules: {
        email: [
          { validator: validateForgotEmail, trigger: 'blur' }
        ],
        code: [
          { required: true, message: '请输入验证码', trigger: 'blur' }
        ],
        new_password: [
          { required: true, message: '请输入新密码', trigger: 'blur' },
          { min: 6, message: '新密码至少 6 位', trigger: 'blur' }
        ],
        confirm_password: [
          { required: true, message: '请再次输入新密码', trigger: 'blur' },
          { validator: validateForgotConfirm, trigger: 'blur' }
        ]
      }
    }
  },
  computed: {
    sendCodeButtonText() {
      return this.sendCodeCountdown > 0 ? `${this.sendCodeCountdown}s 后重发` : '发送验证码'
    },
    forgotSubtitle() {
      return this.codeSent
        ? '验证码已发送到邮箱，请继续完成密码重置。'
        : '请先在登录框填写待找回账号的用户名，再在此输入与该账号一致的注册邮箱以接收验证码。'
    },
    forgotNote() {
      return this.codeSent
        ? '请检查邮箱收件箱或垃圾邮件，并在验证码有效期内完成操作。'
        : '为了保护账号安全，验证码仅用于本次重置流程，发送后 10 分钟内有效。'
    }
  },
  beforeDestroy() {
    this.clearSendCodeTimer()
  },
  methods: {
    ...mapActions(['login']),
    async handleLogin() {
      this.$refs.loginForm.validate(async (valid) => {
        if (valid) {
          this.loginBanner = ''
          this.loading = true
          try {
            await this.login(this.loginForm)
            this.loginBanner = ''
            this.$message.success('登录成功')
            this.$router.push('/dashboard')
          } catch (error) {
            const msg = error.response?.data?.error || '登录失败'
            this.loginBanner = msg
          } finally {
            this.loading = false
          }
        }
      })
    },
    openForgotDialog() {
      this.forgotVisible = true
    },
    handleForgotDialogClose() {
      this.forgotLoading = false
      this.sendCodeLoading = false
      this.clearSendCodeTimer()
      this.sendCodeCountdown = 0
      this.codeSent = false
      this.forgotForm = {
        email: '',
        code: '',
        new_password: '',
        confirm_password: ''
      }
      if (this.$refs.forgotFormRef) {
        this.$refs.forgotFormRef.resetFields()
      }
    },
    clearSendCodeTimer() {
      if (this.sendCodeTimer) {
        clearInterval(this.sendCodeTimer)
        this.sendCodeTimer = null
      }
    },
    startSendCodeCountdown(seconds = 60) {
      this.clearSendCodeTimer()
      this.sendCodeCountdown = seconds
      this.sendCodeTimer = setInterval(() => {
        if (this.sendCodeCountdown <= 1) {
          this.clearSendCodeTimer()
          this.sendCodeCountdown = 0
          return
        }
        this.sendCodeCountdown -= 1
      }, 1000)
    },
    handleSendCode() {
      const username = (this.loginForm.username || '').trim()
      if (!username) {
        this.$message.warning('请先在登录框填写待找回账号的用户名')
        return
      }
      this.$refs.forgotFormRef.validateField('email', async (message) => {
        if (message) return
        this.sendCodeLoading = true
        try {
          const response = await sendForgotPasswordCode({
            username,
            email: this.forgotForm.email
          })
          this.codeSent = true
          this.startSendCodeCountdown()
          this.$message.success(response.data?.message || '验证码已发送')
        } catch (error) {
          this.$message.error(error.response?.data?.error || '验证码发送失败')
        } finally {
          this.sendCodeLoading = false
        }
      })
    },
    handleResetPassword() {
      this.$refs.forgotFormRef.validate(async (valid) => {
        if (!valid) return
        this.forgotLoading = true
        try {
          await resetForgotPassword({
            email: this.forgotForm.email,
            code: this.forgotForm.code,
            new_password: this.forgotForm.new_password
          })
          this.$message.success('密码重置成功，请使用新密码登录')
          this.loginForm.password = ''
          this.forgotVisible = false
          this.clearSendCodeTimer()
          this.sendCodeCountdown = 0
          this.codeSent = false
          this.forgotForm = {
            email: '',
            code: '',
            new_password: '',
            confirm_password: ''
          }
          this.$nextTick(() => {
            if (this.$refs.forgotFormRef) {
              this.$refs.forgotFormRef.clearValidate()
            }
          })
        } catch (error) {
          this.$message.error(error.response?.data?.error || '密码重置失败')
        } finally {
          this.forgotLoading = false
        }
      })
    }
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #eceff4;
  padding: 28px;
  font-family: 'Nunito Sans', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
}

.login-shell {
  width: min(1200px, 96vw);
  min-height: 640px;
  background: #eceff4;
  display: grid;
  grid-template-columns: 1.2fr 0.9fr;
  gap: 42px;
  align-items: center;
}

.login-visual {
  padding: 12px 4px;
}

.brand-row {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  color: #1f7fe5;
  margin-bottom: 14px;
}

.brand-icon {
  font-size: 20px;
}

.brand-text {
  font-weight: 700;
  letter-spacing: 0.02em;
}

.visual-scene {
  position: relative;
  height: 420px;
  border-radius: 18px;
  background: linear-gradient(180deg, #edf5ff 0%, #f5f9ff 100%);
  overflow: hidden;
}

.scene-blob {
  position: absolute;
  border-radius: 50%;
  background: rgba(79, 172, 254, 0.15);
}

.blob-1 {
  width: 260px;
  height: 260px;
  left: 40px;
  top: 50px;
}

.blob-2 {
  width: 220px;
  height: 220px;
  right: 36px;
  top: 96px;
}

.scene-desk {
  position: relative;
  width: 76%;
  height: 150px;
  background: linear-gradient(160deg, #7fc7f7 0%, #51acef 100%);
  border-radius: 12px 12px 18px 18px;
  top: 220px;
  left: 12%;
  box-shadow: 0 14px 30px rgba(80, 149, 207, 0.24);
}

.scene-screen {
  position: absolute;
  width: 120px;
  height: 20px;
  background: #1d7ed2;
  border-radius: 8px;
  left: 45%;
  top: 235px;
}

.scene-person {
  position: absolute;
  width: 44px;
  height: 120px;
  border-radius: 30px;
}

.person-left {
  left: 145px;
  top: 185px;
  background: linear-gradient(180deg, #ffd06c 0%, #f7b63f 100%);
}

.person-right {
  left: 300px;
  top: 177px;
  background: linear-gradient(180deg, #8364d8 0%, #6e4fc2 100%);
}

.scene-plant {
  position: absolute;
  width: 66px;
  height: 76px;
  right: 82px;
  top: 262px;
  border-radius: 12px;
  background: linear-gradient(180deg, #42a5f5 0%, #2b8ee8 100%);
}

.visual-note {
  margin-top: 16px;
  color: #5b677a;
  font-size: 14px;
  font-weight: 500;
}

.login-panel {
  background: rgba(255, 255, 255, 0.82);
  border: 1px solid #d9e4ef;
  border-radius: 16px;
  padding: 34px 32px 28px;
  box-shadow: 0 18px 34px rgba(20, 45, 84, 0.08);
}

.panel-header h1 {
  font-size: 40px;
  line-height: 1.2;
  letter-spacing: -0.02em;
  color: #1f2a44;
  margin: 0;
}

.panel-header p {
  color: #5f6d84;
  margin: 10px 0 20px;
  font-size: 15px;
}

.login-banner {
  margin: 0 0 18px;
}

.login-form :deep(.el-form-item__label) {
  color: #4c5c74;
  font-weight: 600;
  font-size: 13px;
  padding-bottom: 4px;
}

.login-form :deep(.el-input__inner) {
  height: 44px;
  border-radius: 10px;
  background: #f8fbff;
  border-color: #d7e2ef;
}

.login-form :deep(.el-input__inner:focus) {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15);
}

.panel-meta {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: -2px 0 18px;
  font-size: 13px;
}

.remember-text {
  color: #43a047;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.forgot-trigger {
  appearance: none;
  border: none;
  background: transparent;
  color: #8695ab;
  cursor: pointer;
  transition: color 0.2s ease;
  font: inherit;
  padding: 0;
}

.forgot-trigger:hover,
.forgot-trigger:focus {
  color: #2a78df;
  outline: none;
}

.form-actions :deep(.el-form-item__content) {
  display: flex;
  gap: 12px;
}

.btn-login,
.btn-register {
  flex: 1;
  height: 42px;
  border-radius: 999px;
  font-weight: 600;
}

.btn-login {
  background: linear-gradient(135deg, #2a8bf2 0%, #1d73e7 100%) !important;
  border: none !important;
}

.btn-register {
  border-color: #d7e2ef !important;
  color: #51627b !important;
  background: #ffffff !important;
}

.btn-register:hover {
  border-color: #aac8ed !important;
  color: #2a78df !important;
}

.forgot-alert {
  margin-bottom: 18px;
}

.forgot-dialog :deep(.el-dialog) {
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid #d9e4ef;
  box-shadow: 0 24px 50px rgba(20, 45, 84, 0.12);
  overflow: hidden;
}

.forgot-dialog :deep(.el-dialog__header) {
  padding: 18px 24px 0;
}

.forgot-dialog :deep(.el-dialog__title) {
  color: #1f2a44;
  font-size: 18px;
  font-weight: 700;
}

.forgot-dialog :deep(.el-dialog__body) {
  padding: 12px 24px 24px;
}

.forgot-dialog :deep(.el-dialog__headerbtn) {
  top: 18px;
  right: 18px;
}

.forgot-card {
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.96) 0%, rgba(248, 251, 255, 0.95) 100%);
}

.forgot-card__hero {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  margin-bottom: 18px;
}

.forgot-card__badge {
  width: 42px;
  height: 42px;
  border-radius: 14px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, rgba(42, 139, 242, 0.16) 0%, rgba(29, 115, 231, 0.2) 100%);
  color: #1d73e7;
  font-size: 18px;
  flex-shrink: 0;
}

.forgot-card__hero h2 {
  margin: 0 0 6px;
  color: #1f2a44;
  font-size: 22px;
  line-height: 1.3;
}

.forgot-card__hero p {
  margin: 0;
  color: #5f6d84;
  font-size: 14px;
  line-height: 1.6;
}

.forgot-steps {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 10px;
  margin-bottom: 18px;
}

.forgot-step {
  padding: 10px 12px;
  border-radius: 12px;
  background: rgba(238, 244, 252, 0.9);
  color: #7c8ba3;
  font-size: 13px;
  font-weight: 600;
  text-align: center;
  transition: all 0.2s ease;
}

.forgot-step.is-active {
  background: rgba(42, 139, 242, 0.12);
  color: #1d73e7;
  box-shadow: inset 0 0 0 1px rgba(42, 139, 242, 0.14);
}

.forgot-form :deep(.el-form-item__label) {
  color: #4c5c74;
  font-weight: 600;
  font-size: 13px;
  padding-bottom: 4px;
}

.forgot-form :deep(.el-input__inner) {
  height: 44px;
  border-radius: 10px;
  background: #f8fbff;
  border-color: #d7e2ef;
}

.forgot-form :deep(.el-input__inner:focus) {
  border-color: #3b82f6;
  box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.15);
}

.forgot-note {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin: -2px 0 18px;
  padding: 12px 14px;
  border-radius: 12px;
  background: rgba(237, 245, 255, 0.8);
  color: #5f6d84;
  font-size: 13px;
  line-height: 1.6;
}

.forgot-note i {
  color: #2a78df;
  margin-top: 2px;
}

.forgot-reset-fields {
  padding-top: 2px;
}

.send-code-button {
  min-width: 112px;
  color: #1d73e7;
  font-weight: 600;
}

.forgot-actions {
  display: flex;
  gap: 12px;
  margin-top: 22px;
}

.btn-forgot-primary,
.btn-forgot-secondary {
  flex: 1;
  height: 42px;
  border-radius: 999px;
  font-weight: 600;
}

.btn-forgot-primary {
  background: linear-gradient(135deg, #2a8bf2 0%, #1d73e7 100%) !important;
  border: none !important;
}

.btn-forgot-secondary {
  border-color: #d7e2ef !important;
  color: #51627b !important;
  background: #ffffff !important;
}

.btn-forgot-secondary:hover {
  border-color: #aac8ed !important;
  color: #2a78df !important;
}

.fade-slide-enter-active,
.fade-slide-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.fade-slide-enter,
.fade-slide-leave-to {
  opacity: 0;
  transform: translateY(6px);
}

@media (max-width: 1024px) {
  .login-shell {
    grid-template-columns: 1fr;
    gap: 20px;
  }
  .login-visual {
    order: 2;
  }
  .login-panel {
    order: 1;
  }
  .visual-scene {
    height: 260px;
  }
  .scene-desk {
    top: 132px;
  }
  .scene-screen {
    top: 145px;
  }
  .person-left {
    top: 96px;
  }
  .person-right {
    top: 92px;
  }
  .scene-plant {
    top: 170px;
  }
}

@media (max-width: 560px) {
  .login-page {
    padding: 14px;
  }
  .login-panel {
    padding: 24px 18px 18px;
  }
  .panel-header h1 {
    font-size: 30px;
  }
  .form-actions :deep(.el-form-item__content) {
    flex-direction: column;
  }
  .forgot-dialog :deep(.el-dialog) {
    width: calc(100vw - 24px) !important;
    margin-top: 24px !important;
  }
  .forgot-actions {
    flex-direction: column;
  }
}
</style>
