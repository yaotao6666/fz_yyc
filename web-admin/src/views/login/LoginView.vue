<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { APP_TITLE } from '@/config/env'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const loading = ref(false)

const form = reactive({
  username: '',
  password: ''
})

async function handleSubmit() {
  if (!form.username.trim()) {
    ElMessage.warning('请输入商家账号')
    return
  }
  if (!form.password.trim()) {
    ElMessage.warning('请输入密码')
    return
  }

  loading.value = true
  try {
    await authStore.login(form.username.trim(), form.password.trim())
    ElMessage.success('登录成功')
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/dashboard'
    await router.replace(redirect)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <div class="title-wrap">
        <div class="eyebrow">Merchant Admin</div>
        <h1>{{ APP_TITLE }}</h1>
        <p>商家后台管理系统，请使用商家账号登录以管理商品、订单与服务人员。</p>
      </div>

      <el-form label-position="top" @submit.prevent>
        <el-form-item label="商家账号">
          <el-input v-model="form.username" placeholder="请输入商家账号" size="large" @keyup.enter="handleSubmit" />
        </el-form-item>
        <el-form-item label="登录密码">
          <el-input v-model="form.password" show-password type="password" placeholder="请输入密码" size="large" @keyup.enter="handleSubmit" />
        </el-form-item>
        <el-button type="primary" size="large" :loading="loading" class="submit-btn" @click="handleSubmit">
          登录后台
        </el-button>
      </el-form>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background:
    radial-gradient(circle at top left, rgba(59, 130, 246, 0.16), transparent 30%),
    radial-gradient(circle at bottom right, rgba(14, 165, 233, 0.16), transparent 28%),
    linear-gradient(180deg, #eef3fb 0%, #f8fafc 100%);
}

.login-card {
  width: 100%;
  max-width: 460px;
  padding: 36px;
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 32px 60px rgba(15, 23, 42, 0.12);
}

.eyebrow {
  display: inline-flex;
  padding: 6px 12px;
  border-radius: 999px;
  background: #e8f0fe;
  color: #1d4ed8;
  font-size: 12px;
  font-weight: 600;
}

.title-wrap h1 {
  margin: 16px 0 12px;
  font-size: 30px;
  color: #111827;
}

.title-wrap p {
  margin: 0 0 28px;
  color: #6b7280;
  line-height: 1.7;
}

.submit-btn {
  width: 100%;
  margin-top: 8px;
}
</style>
