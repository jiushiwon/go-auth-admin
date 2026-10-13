<!--
  FlipLoginContainer.vue — flip 风格的容器组件

  遵循 SKILL.md 形态三：3D 翻转
    .login-page--flip           提供 perspective: 1000px
    .flip-container             transform-style: preserve-3d; transition: transform 0.8s
    .flip-container.flipped     transform: rotateY(180deg)
    .login-card--flip           position: absolute; backface-visibility: hidden（默认正面）
    .login-card--flip-back      transform: rotateY(180deg)（预旋 180°，翻过来时正对用户）

  包含两个 LoginForm 实例：
    - 正面（variant="flip"）：登录表单
    - 反面（variant="flip-back"）：注册表单（用 #extraLinks 槽替换默认注册引导，
      槽里直接给"已有账号？立即登录"链接，让用户切回正面）

  isFlipped 状态由内部管理，登录/注册成功后由父组件 @success 回调通知。
-->
<template>
  <div class="flip-container" :class="{ flipped: isFlipped }">
    <!-- 正面：登录（沿用 frosted 卡片样式 + flip 定位） -->
    <LoginForm
      variant="frosted"
      :title="'登录'"
      :subtitle="'欢迎回来'"
      :username-placeholder="'请输入用户名 / 邮箱'"
      :password-placeholder="'请输入密码'"
      :submit-text="'登 录'"
      :show-register="false"
      :loading="loginLoading"
      @submit="onLoginSubmit"
    >
      <template #extraLinks>
        <div class="login-card__register">
          还没有账号？
          <base-button variant="link" size="sm" @click="flipToRegister">立即注册</base-button>
        </div>
      </template>
    </LoginForm>

    <!-- 反面：注册（同 frosted 卡片样式，自身 rotateY(180deg) 让翻过来正对用户） -->
    <LoginForm
      variant="frosted"
      :title="'注册'"
      :subtitle="'创建你的账号'"
      :username-placeholder="'请输入用户名'"
      :password-placeholder="'请输入密码'"
      :submit-text="'注 册'"
      :show-remember="false"
      :show-forgot="false"
      :show-register="false"
      :loading="registerLoading"
      @submit="onRegisterSubmit"
    >
      <template #extraLinks>
        <div class="login-card__register">
          已有账号？
          <base-button variant="link" size="sm" @click="flipToLogin">立即登录</base-button>
        </div>
      </template>
    </LoginForm>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { LoginForm } from 'vue-login-skill'
import { BaseButton } from 'vue-button-skill'

const emit = defineEmits<{
  login: [values: { username: string; password: string }]
  register: [values: { username: string; password: string }]
}>()

const isFlipped = ref(false)
const loginLoading = ref(false)
const registerLoading = ref(false)

function flipToRegister() {
  isFlipped.value = true
}

function flipToLogin() {
  isFlipped.value = false
}

async function onLoginSubmit(values: { username: string; password: string }) {
  loginLoading.value = true
  try {
    emit('login', values)
    /* ponytail: 父组件 @login 回调决定何时翻回或跳转。loading 状态由父组件控制更干净，
       但此处保留简单实现 —— 父组件成功后会路由跳转，组件自然卸载。*/
  } finally {
    /* 父组件如果失败会保持原状态；成功则跳转走 unload 路径
       （loading 在组件 unmount 时自动清理）*/
    setTimeout(() => { loginLoading.value = false }, 500)
  }
}

async function onRegisterSubmit(values: { username: string; password: string }) {
  registerLoading.value = true
  try {
    emit('register', values)
  } finally {
    setTimeout(() => { registerLoading.value = false }, 500)
  }
}
</script>

<style scoped>
/* SKILL.md 形态三：3D 翻转 — 容器 */
.flip-container {
  position: relative;
  width: 420px;
  height: 480px;
  /* ponytail: SKILL.md 默认是 transition: transform 0.8s（慢）；
     用 0.6s 更跟手但不晃眼。ease-in-out 起步收尾平滑。*/
  transform-style: preserve-3d;
  transition: transform 0.6s ease-in-out;
}

.flip-container.flipped {
  transform: rotateY(180deg);
}

/* ponytail: SKILL.md 形态三 CSS 只给 .login-card--flip 设了 absolute + backface-hidden，
     没给卡片配色。这里让卡片沿用 frosted 风格（半透明 + blur + 紫蓝边框），
     并把第二张卡片预旋 180° 让它翻过来时正对用户。*/
.flip-container :deep(.login-card) {
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.12) !important;
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  border: 1px solid rgba(99, 102, 241, 0.35) !important;
  box-shadow: 0 8px 32px rgba(99, 102, 241, 0.15) !important;
  backface-visibility: hidden;
  -webkit-backface-visibility: hidden;
}

/* 第二张卡片（注册）：预旋 180°，翻过来时正面对着用户 */
.flip-container :deep(.login-card:nth-child(2)) {
  transform: rotateY(180deg);
}
</style>
