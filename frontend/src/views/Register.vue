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
        <p class="visual-note">加入跨境冷链物流溯源与监管平台</p>
      </section>

      <section class="login-panel register-panel">
        <div class="panel-header">
          <h1>Create Account :)</h1>
          <p>填写信息完成注册，已有账号可返回登录</p>
        </div>

        <el-form
          ref="registerForm"
          :model="registerForm"
          :rules="rules"
          label-position="top"
          class="login-form register-form"
          @submit.native.prevent="handleRegister"
        >
          <el-form-item label="Username" prop="username">
            <el-input
              v-model="registerForm.username"
              placeholder="请输入用户名"
              prefix-icon="el-icon-user"
              clearable
            />
          </el-form-item>
          <el-form-item label="Password" prop="password">
            <el-input
              v-model="registerForm.password"
              type="password"
              placeholder="至少 6 位字符"
              prefix-icon="el-icon-lock"
              show-password
            />
          </el-form-item>
          <el-form-item label="Confirm Password" prop="confirmPassword">
            <el-input
              v-model="registerForm.confirmPassword"
              type="password"
              placeholder="请再次输入密码"
              prefix-icon="el-icon-lock"
              show-password
            />
          </el-form-item>
          <el-form-item label="Email" prop="email">
            <el-input
              v-model="registerForm.email"
              placeholder="请输入邮箱地址"
              prefix-icon="el-icon-message"
              clearable
            />
          </el-form-item>
          <el-form-item label="Role" prop="role">
            <el-select
              v-model="registerForm.role"
              placeholder="请选择您的角色"
              style="width: 100%"
            >
              <el-option label="生产商" value="producer" />
              <el-option label="仓储商" value="warehouse" />
              <el-option label="物流商" value="logistics" />
              <el-option label="消费者" value="consumer" />
              <el-option label="监管" value="regulator" />
            </el-select>
          </el-form-item>
          <el-form-item label="Company" prop="companyName">
            <el-input
              v-model="registerForm.companyName"
              placeholder="选填，物流/生产商建议填写"
              prefix-icon="el-icon-office-building"
              clearable
            />
          </el-form-item>

          <div class="panel-meta register-meta">
            <span class="remember-text">
              <i class="el-icon-info"></i>
              角色与公司将用于数据隔离
            </span>
          </div>

          <el-form-item class="form-actions">
            <el-button
              type="primary"
              class="btn-login"
              :loading="loading"
              @click="handleRegister"
            >
              Sign Up
            </el-button>
            <el-button class="btn-register" @click="$router.push('/login')">
              Back to Login
            </el-button>
          </el-form-item>

          <p class="register-footer">
            已有账号？
            <router-link to="/login" class="register-link">立即登录</router-link>
          </p>
        </el-form>
      </section>
    </div>
  </div>
</template>

<script>
import { mapActions } from 'vuex'

export default {
  name: 'Register',
  data() {
    const validateConfirmPassword = (rule, value, callback) => {
      if (value !== this.registerForm.password) {
        callback(new Error('两次输入密码不一致'))
      } else {
        callback()
      }
    }
    return {
      registerForm: {
        username: '',
        password: '',
        confirmPassword: '',
        email: '',
        role: '',
        companyName: ''
      },
      loading: false,
      rules: {
        username: [
          { required: true, message: '请输入用户名', trigger: 'blur' }
        ],
        password: [
          { required: true, message: '请输入密码', trigger: 'blur' },
          { min: 6, message: '密码长度至少 6 位', trigger: 'blur' }
        ],
        confirmPassword: [
          { required: true, message: '请确认密码', trigger: 'blur' },
          { validator: validateConfirmPassword, trigger: 'blur' }
        ],
        email: [
          { required: true, message: '请输入邮箱', trigger: 'blur' },
          { type: 'email', message: '请输入正确的邮箱地址', trigger: 'blur' }
        ],
        role: [
          { required: true, message: '请选择角色', trigger: 'change' }
        ]
      }
    }
  },
  methods: {
    ...mapActions(['register']),
    async handleRegister() {
      this.$refs.registerForm.validate(async (valid) => {
        if (valid) {
          this.loading = true
          try {
            await this.register({
              username: this.registerForm.username,
              password: this.registerForm.password,
              email: this.registerForm.email,
              role: this.registerForm.role,
              company_name: this.registerForm.companyName
            })
            this.$message.success('注册成功，请登录')
            this.$router.push('/login')
          } catch (error) {
            this.$message.error(error.response?.data?.error || '注册失败')
          } finally {
            this.loading = false
          }
        }
      })
    }
  }
}
</script>

<style scoped>
/* 与 Login.vue 保持同一套布局与视觉 */
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
  align-items: start;
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

.register-panel {
  max-height: min(90vh, 880px);
  overflow-y: auto;
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
  margin: 10px 0 16px;
  font-size: 15px;
}

.login-form :deep(.el-form-item) {
  margin-bottom: 12px;
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

.login-form :deep(.el-select .el-input__inner) {
  height: 44px;
}

.panel-meta.register-meta {
  margin: 4px 0 14px;
}

.remember-text {
  color: #5f6d84;
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
}

.remember-text .el-icon-info {
  color: #2a8bf2;
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

.register-footer {
  text-align: center;
  font-size: 14px;
  color: #5f6d84;
  margin: 8px 0 0;
}

.register-link {
  color: #2a8bf2;
  font-weight: 600;
  text-decoration: none;
  margin-left: 4px;
}

.register-link:hover {
  text-decoration: underline;
}

@media (max-width: 1024px) {
  .login-shell {
    grid-template-columns: 1fr;
    gap: 20px;
    align-items: stretch;
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
}
</style>
