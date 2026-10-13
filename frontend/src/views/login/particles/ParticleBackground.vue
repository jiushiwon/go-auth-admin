<!--
  ParticleBackground.vue — particles 风格登录页的容器背景层

  封装 demo-components/shared/particle-engine.js（原 ES5 全局类）为 Vue 组件：
    <ParticleBackground /> 直接铺满父容器，提供粒子连线动画背景。

  props 透传引擎配置（颜色 / 数量 / 速度等），不重复声明。

  生命周期：
    - onMounted：等下一帧拿到 canvas DOM 后 new ParticleBackground
    - onBeforeUnmount：调 engine.destroy() 清理 animationFrame + resize listener
-->
<template>
  <canvas ref="canvasRef" class="particles-canvas" />
</template>

<script setup lang="ts">
import { onMounted, onBeforeUnmount, ref } from 'vue'
import './particle-engine'  /* 副作用导入：执行 particle-engine.js，挂 ParticleBackground 到 window */

/* ponytail: particle-engine.js 是 SKILL.md 提供的 ES5 全局类（window.ParticleBackground），
     不导出 ES module。Vue 组件通过全局声明拿到类型 + 实例。*/
declare global {
  interface Window {
    ParticleBackground: new (canvas: HTMLCanvasElement, options?: any) => {
      destroy(): void
    }
  }
}

interface Props {
  /** 粒子数（默认 80） */
  particleCount?: number
  /** 连线最大距离（默认 120） */
  connectionDistance?: number
  /** 粒子基础尺寸（默认 2） */
  particleSize?: number
  /** 移动速度（默认 1） */
  speed?: number
  /** 粒子 / 连线颜色（默认 #409eff） */
  color?: string
  /** 粒子不透明度（默认 0.6） */
  opacity?: number
  /** 鼠标交互半径（默认 150） */
  mouseRadius?: number
}

const props = withDefaults(defineProps<Props>(), {
  particleCount: 80,
  connectionDistance: 120,
  particleSize: 2,
  speed: 1,
  /* ponytail: 用品牌色 #409eff 默认值是 SKILL.md 演示，java-vue-admin 是铜橙主题，
     消费方应传主题色 token。在父组件 index.vue 里覆盖。*/
  color: '#6366f1',
  opacity: 0.6,
  mouseRadius: 150,
})

const canvasRef = ref<HTMLCanvasElement | null>(null)
let engine: InstanceType<Window['ParticleBackground']> | null = null

onMounted(() => {
  if (!canvasRef.value) return
  /* ponytail: 等下一帧让 canvas 进入 DOM 并拿到正确 width/height，
     否则第一次 new ParticleBackground 时 canvas 尺寸还是 0 */
  requestAnimationFrame(() => {
    if (!canvasRef.value) return
    engine = new window.ParticleBackground(canvasRef.value, {
      particleCount: props.particleCount,
      connectionDistance: props.connectionDistance,
      particleSize: props.particleSize,
      speed: props.speed,
      color: props.color,
      opacity: props.opacity,
      mouseRadius: props.mouseRadius,
    })
  })
})

onBeforeUnmount(() => {
  engine?.destroy()
  engine = null
})
</script>

<style scoped>
.particles-canvas {
  position: absolute;
  inset: 0;
  width: 100%;
  height: 100%;
  z-index: 0;
  pointer-events: auto;
}
</style>
