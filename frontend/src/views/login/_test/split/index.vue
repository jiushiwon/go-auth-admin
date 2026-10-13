<!--
  _test/split/index.vue - 验证 split 风格（半屏 50/50 分屏）

  复用 vue-login-skill 的 SplitLayout + LoginForm variant="split"
  对照 demo: D:/projects/vibecoding-portal/skills/vibeCoding/frontend/vue/vue-base-skill/vue-complex-skill/vue-login-skill/demo-components/login-page/html/04-split.html
-->
<template>
  <div v-if="message" class="login-toast" :data-type="message.type">
    {{ message.text }}
  </div>
  <SplitLayout>
    <template #brand>
      <div class="brand-stack">
        <div class="brand-logo">M</div>
        <h2 class="brand-title">管理系统</h2>
        <p class="brand-subtitle">一站式企业级后台管理解决方案</p>
        <ul class="brand-features">
          <li>
            <svg viewBox="0 0 24 24"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
            强大的 RBAC 权限管理
          </li>
          <li>
            <svg viewBox="0 0 24 24"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
            可视化数据看板
          </li>
          <li>
            <svg viewBox="0 0 24 24"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
            完整的审计日志
          </li>
          <li>
            <svg viewBox="0 0 24 24"><path d="M9 16.17L4.83 12l-1.42 1.41L9 19 21 7l-1.41-1.41z"/></svg>
            企业级安全保障
          </li>
        </ul>
      </div>
    </template>

    <template #form>
      <LoginForm
        variant="split"
        title="欢迎回来"
        subtitle="请登录您的账户"
        username-placeholder="请输入邮箱"
        password-placeholder="请输入密码"
        :loading="loading"
        @submit="handleSubmit"
      >
        <template #divider>
          <div class="login-card__divider">或</div>
        </template>

        <template #social>
          <div class="social-login">
            <div class="social-login__btn" aria-label="微信登录">
              <svg viewBox="0 0 24 24"><path d="M9 11.5a1 1 0 100-2 1 1 0 000 2zm6 0a1 1 0 100-2 1 1 0 000 2zM7.5 5C4.46 5 2 7.24 2 10c0 1.6.85 3.03 2.18 3.96L3.5 16l2.36-1.18c.5.1 1.04.18 1.64.18.27 0 .53-.02.79-.05-.16-.5-.29-1.02-.29-1.55 0-2.76 2.69-5 6-5 .27 0 .53.02.79.05C14.21 6.55 11.16 5 7.5 5z"/></svg>
            </div>
            <div class="social-login__btn" aria-label="GitHub 登录">
              <svg viewBox="0 0 24 24"><path d="M12 2C6.48 2 2 6.48 2 12c0 4.42 2.87 8.17 6.84 9.5.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34-.46-1.16-1.11-1.47-1.11-1.47-.91-.62.07-.6.07-.6 1 .07 1.53 1.03 1.53 1.03.87 1.52 2.34 1.07 2.91.83.09-.65.35-1.09.63-1.34-2.22-.25-4.55-1.11-4.55-4.94 0-1.1.39-1.99 1.03-2.69-.1-.25-.45-1.27.1-2.65 0 0 .84-.27 2.75 1.02.79-.22 1.65-.33 2.5-.33s1.71.11 2.5.33c1.91-1.29 2.75-1.02 2.75-1.02.55 1.38.2 2.4.1 2.65.64.7 1.03 1.59 1.03 2.69 0 3.84-2.34 4.68-4.57 4.93.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0012 2z"/></svg>
            </div>
            <div class="social-login__btn" aria-label="Google 登录">
              <svg viewBox="0 0 24 24"><path d="M21.35 11.1H12v3.2h5.35c-.5 2.42-2.42 4.5-5.35 4.5-3.32 0-6-2.68-6-6s2.68-6 6-6c1.5 0 2.85.55 3.9 1.45l2.4-2.4C16.85 4.05 14.55 3 12 3 7.05 3 3 7.05 3 12s4.05 9 9 9c4.65 0 8.7-3.5 8.7-9 0-.6-.15-1.25-.35-1.9z"/></svg>
            </div>
          </div>
        </template>

        <template #extraLinks>
          <div class="login-card__extra-links">
            <a href="#">注册账户</a>
            <span>|</span>
            <a href="#">手机号登录</a>
          </div>
        </template>
      </LoginForm>
    </template>
  </SplitLayout>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { LoginForm, SplitLayout } from 'vue-login-skill'
import { useUserStore } from '@/store/user'
import { extractErrorMessage } from '@/utils/error'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()
const loading = ref(false)
const message = ref<{ type: 'success' | 'error'; text: string } | null>(null)

async function handleSubmit(values: { username: string; password: string }) {
  loading.value = true
  try {
    // ponytail: store.login() 内部串好 setToken + fetchUserInfo + loadMenus，
    // 消费方只 await + 处理跳转/错误，不要再 import authApi 重复设置 token
    await userStore.login(values)
    message.value = { type: 'success', text: `登录成功，欢迎回来 ${userStore.userInfo?.nickname ?? values.username}` }
    const redirect = (route.query.redirect as string) || '/'
    setTimeout(() => router.replace(redirect), 600)
  } catch (err: unknown) {
    message.value = { type: 'error', text: extractErrorMessage(err) }
  } finally {
    loading.value = false
    setTimeout(() => { message.value = null }, 2500)
  }
}
</script>

<style scoped>
/* 极简 toast：浮动右下角 */
.login-toast {
  position: fixed;
  right: var(--space-6);
  bottom: var(--space-6);
  padding: var(--space-4) var(--space-6);
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
  font-size: var(--font-sm);
  z-index: 9999;
}

/* SplitLayout 内部 .split-layout__brand 是 flex 居中，
   槽位内容要包一个 column 容器让 logo / title / features 垂直堆叠 */
/* SplitLayout 内部 .split-layout__brand 是 flex 居中容器；
   槽位用 .brand-stack 作为唯一根，让 logo / title / features 垂直堆叠 + 居中 */
.brand-stack {
  display: flex;
  flex-direction: column;
  align-items: center;
  max-width: 400px;
  text-align: center;
}

.brand-logo {
  width: 64px;
  height: 64px;
  margin: 0 auto var(--space-6);
  border-radius: var(--radius-lg);
  background: rgba(255, 255, 255, 0.15);
  border: 1px solid rgba(255, 255, 255, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: var(--font-3xl);
  font-weight: var(--weight-semibold);
  backdrop-filter: blur(10px);
}
.brand-features li {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  font-size: var(--font-base);
}
.brand-features svg {
  width: 20px;
  height: 20px;
  fill: var(--color-success);
  flex-shrink: 0;
}

.social-login {
  display: flex;
  justify-content: center;
  gap: var(--space-4);
  margin: var(--space-4) 0 var(--space-6);
}
.social-login__btn {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: var(--color-bg);
  border: 1px solid var(--color-border);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: all 0.2s;
}
.social-login__btn:hover {
  border-color: var(--color-primary);
  transform: translateY(-1px);
}
.social-login__btn svg {
  width: 22px;
  height: 22px;
  fill: var(--color-text-secondary);
}
.social-login__btn:hover svg {
  fill: var(--color-primary);
}
</style>