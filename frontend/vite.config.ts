import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import path from 'path'

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: [
      // 把第三方子技能（如 vue-button-skill/BaseButton.vue 等）里
      // 'frontend-icon-skill/components/BaseIcon.vue' 的 import 重定向到本地 BaseIcon.vue。
      // 原因：frontend-icon-skill 的 components/BaseIcon.vue 内部 '../src/index.ts'，
      // 而 src/index.ts 顶层 import 'node:fs/promises'（CLI 工具 generateIconToFile 用），
      // 浏览器端不能用 fs → 报 "Module 'node:fs/promises' has been externalized"。
      // 本地 src/components/icons/BaseIcon.vue 已把 18 个 SVG 模板 hardcode 进组件，零外部依赖。
      // alias.find 匹配完整模块 specifier（含 .pnpm 前缀），因为 pnpm 的软链会让
      // 'frontend-icon-skill/...' 解析成 .pnpm/.../frontend-icon-skill/.../frontend-icon-skill/...
      {
        find: /(\/node_modules\/.pnpm\/.*?\/)?frontend-icon-skill\/components\/BaseIcon\.vue$/,
        replacement: path.resolve(__dirname, 'src/components/icons/BaseIcon.vue'),
      },
      {
        find: '@',
        replacement: path.resolve(__dirname, 'src')
      },
    ],
    dedupe: ['vue']
  },
  server: {
    port: 5173,
    host: '0.0.0.0',
    proxy: {
      '/api': {
        target: 'http://localhost:8080',
        changeOrigin: true,
        ws: false,
        secure: false
      }
    }
  },
  optimizeDeps: {
    // 软链接包预构建（11 个 vue-*-skill + 4 个核心）
    include: [
      'vue',
      'vue-router',
      'pinia',
      'axios',
      'vue-card-skill',
      'vue-button-skill',
      'vue-tag-skill',
      'vue-dropdown-skill',
      'vue-input-skill',
      'vue-select-skill',
      'vue-checkbox-skill',
      'vue-switch-skill',
      'vue-table-skill',
      'vue-form-skill',
      'vue-tree-skill',
      'vue-dialog-skill',
      'vue-layout-skill',
      // 'vue-login-skill',         // 见 exclude 注释
      // 'frontend-icon-skill',     // 见 exclude 注释
    ],
    /*
     * 排除两个含特殊副作用的包：
     *
     * 1. vue-login-skill：它的 LoginForm.vue 含 unscoped <style>（.login-card--frosted、
     *    backdrop-filter、color-mix 等），被预构建时 vite optimizeDeps 只取 JS，<style> 被丢弃
     *    → 登录页 frosted 毛玻璃完全失效（用户报告"一片空白就登录一个文案能看见"）。
     *    排除后 vite 走完整 .vue SFC 编译流水线，<style> 正常注入。
     *
     * 2. frontend-icon-skill：package.json main = src/index.ts，**顶层 import 'node:fs/promises'**
     *    （CLI 工具 generateIconToFile 用）。子技能如 vue-button-skill/BaseButton.vue 通过
     *    'frontend-icon-skill/components/BaseIcon.vue' 路径引用 BaseIcon，被预打包时连带把 fs/promises
     *    带到浏览器 → 浏览器报错 "Module 'node:fs/promises' has been externalized"。
     *    排除后子技能 import 直接走 plugin-vue 处理 .vue，不触发 src/index.ts 的 fs import。
     */
    exclude: ['vue-login-skill', 'frontend-icon-skill']
  }
})
