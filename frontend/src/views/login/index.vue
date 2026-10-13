<!--
  views/login/index.vue - 登录页入口（v3：用 vue-login-skill 的 LoginForm frosted + 自带页面背景）

  frosted 的 ponytail 约束：依赖 backdrop-filter 看到背后内容，必须给 .login-page 加非纯色背景
  （铜色径向渐变 + 噪点纹理），否则白底上半透明卡片不可见。

  业务逻辑（userStore.login + 路由跳转 + 错误提示）在本组件内绑定，
  LoginForm 只负责视觉与表单校验（vue-form-skill 在 LoginForm 内部完成 form-contract §十三 R2）。
-->
<template>
  <div class="login-page">
    <div class="login-page__bg" aria-hidden="true" />

    <LoginForm
      variant="frosted"
      title="Vue Admin"
      subtitle="基于 wg-skills 技能矩阵 · copper 主题"
      :loading="loading"
      username-placeholder="请输入用户名"
      password-placeholder="请输入密码"
      submit-text="登 录"
      :show-remember="true"
      :show-forgot="false"
      :show-register="false"
      remember-label="记住我"
      @submit="handleSubmit"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { LoginForm } from 'vue-login-skill'
import { useUserStore } from '@/store/user'
import { extractErrorMessage } from '@/utils/error'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const loading = ref(false)

async function handleSubmit(values: { username: string; password: string; remember?: boolean }) {
  loading.value = true
  try {
    // userStore.login 内部已串行调用 fetchUserInfo + loadMenus
    await userStore.login({ username: values.username, password: values.password })
    const redirect = (route.query.redirect as string) || '/dashboard'
    router.push(redirect)
  } catch (e) {
    // LoginForm 自身不带错误提示（设计成专注视觉），用 alert 兜底
    // 生产环境应接 toast-skill
    alert(extractErrorMessage(e) || '登录失败，请检查用户名密码')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
/*
  ponytail: frosted 毛玻璃依赖 backdrop-filter 看到背后内容，必须配套页面背景：
    - 铜色径向渐变（用 tokens.css 的 --color-primary-{50..950}）
    - 装饰光斑（绝对定位，不阻挡交互）
  否则白底上半透明卡片不可见（用户报告的"一片空白"就是这个原因）。
*/
.login-page {
  position: relative;
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: var(--space-6);
  background:
    radial-gradient(at 20% 20%, var(--color-primary-300) 0%, transparent 50%),
    radial-gradient(at 80% 30%, var(--color-primary-400) 0%, transparent 50%),
    radial-gradient(at 50% 90%, var(--color-primary-700) 0%, transparent 60%),
    linear-gradient(135deg, var(--color-primary-100), var(--color-primary-200));
  overflow: hidden;
}

/* 背景光斑装饰层（aria-hidden 不进入无障碍树） */
.login-page__bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
  background-image:
    radial-gradient(circle at 30% 40%, rgba(255, 255, 255, 0.4) 0%, transparent 8%),
    radial-gradient(circle at 70% 60%, rgba(255, 255, 255, 0.3) 0%, transparent 10%),
    radial-gradient(circle at 50% 80%, rgba(255, 255, 255, 0.25) 0%, transparent 12%);
  z-index: 0;
}

/* LoginForm 浮在背景之上 */
.login-page > :deep(.login-card) {
  position: relative;
  z-index: 1;
  max-width: 420px;
  width: 100%;
}
</style>
