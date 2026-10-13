<!--
  BaseIcon —— 全项目图标唯一入口（v2：图标模板内联到组件，零外部依赖）

  图标方案（唯一口径，详见 frontend/ICONS.md）：
    1) DB `wg_sys_menu.icon` / 路由 `meta.icon` 统一存 **kebab-case 图标名**（lucide 命名风格，
       如 `layout-dashboard` / `users` / `settings`），不存 Unicode 图形字符；
    2) 本组件把图标名映射到内联的 SVG 模板并渲染；
    3) 未登记的图标名走兜底模板，绝不渲染成空白，也不会把原始字符串直接显示出来。

  v2 改造原因：原版 import { generateIcon, listTemplates } from 'frontend-icon-skill' 会触发
    frontend-icon-skill/src/index.ts 的 'node:fs/promises' 顶层 import（CLI 工具 generateIconToFile 用），
    浏览器报错 "Module 'node:fs/promises' has been externalized for browser compatibility"。
    现在把模板 hardcode 到组件内部，零外部依赖，零 fs import。
-->
<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  /** 图标名（kebab-case，如 'layout-dashboard' / 'users' / 'settings'） */
  name: string
  /** 图标尺寸（px） */
  size?: number
  /** 描边色（CSS color） */
  color?: string
  /** 描边粗细 */
  strokeWidth?: number
}>(), {
  size: 24,
  color: 'currentColor',
  strokeWidth: 2,
})

/**
 * SVG 模板（lucide 风格）。所有图标都用 stroke="currentColor"，运行时通过 props.color 替换。
 * 占位符：{{size}} / {{color}} / {{strokeWidth}}
 */
const TEMPLATES: Record<string, string> = {
  dashboard: '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><rect x="3" y="3" width="7" height="9"/><rect x="14" y="3" width="7" height="5"/><rect x="14" y="12" width="7" height="9"/><rect x="3" y="16" width="7" height="5"/></svg>',
  user: '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>',
  role: '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="16" r="2"/><path d="M12 14V8"/><circle cx="12" cy="6" r="2"/><path d="M5 10a7 7 0 0 1 14 0"/><path d="M3 18l1.5-4"/><path d="M21 18l-1.5-4"/></svg>',
  menu: '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><line x1="4" y1="6" x2="20" y2="6"/><line x1="4" y1="12" x2="20" y2="12"/><line x1="4" y1="18" x2="20" y2="18"/></svg>',
  settings: '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><path d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z"/><circle cx="12" cy="12" r="3"/></svg>',
  building: '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="2" width="16" height="20" rx="2"/><path d="M9 22v-4h6v4"/><path d="M8 6h.01"/><path d="M16 6h.01"/><path d="M12 6h.01"/><path d="M12 10h.01"/><path d="M12 14h.01"/><path d="M16 10h.01"/><path d="M16 14h.01"/><path d="M8 10h.01"/><path d="M8 14h.01"/></svg>',
  box: '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><polyline points="3.27 6.96 12 12.01 20.73 6.96"/><line x1="12" y1="22.08" x2="12" y2="12"/></svg>',
  chart: '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><line x1="4" y1="20" x2="4" y2="10"/><line x1="10" y1="20" x2="10" y2="4"/><line x1="16" y1="20" x2="16" y2="14"/><line x1="20" y1="20" x2="20" y2="8"/><line x1="2" y1="20" x2="22" y2="20"/></svg>',
  tree: '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>',
  logout: '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" y1="12" x2="9" y2="12"/></svg>',
  add: '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><line x1="12" y1="5" x2="12" y2="19"/><line x1="5" y1="12" x2="19" y2="12"/></svg>',
  edit: '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>',
  delete: '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"/><path d="M19 6l-2 14a2 2 0 0 1-2 2H9a2 2 0 0 1-2-2L5 6"/><path d="M10 11v6"/><path d="M14 11v6"/></svg>',
  search: '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"/><line x1="21" y1="21" x2="16.65" y2="16.65"/></svg>',
  success: '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>',
  warning: '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>',
  'chevron-down': '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg>',
  'chevron-left': '<svg xmlns="http://www.w3.org/2000/svg" width="{{size}}" height="{{size}}" viewBox="0 0 24 24" fill="none" stroke="{{color}}" stroke-width="{{strokeWidth}}" stroke-linecap="round" stroke-linejoin="round"><polyline points="15 18 9 12 15 6"/></svg>',
}

/** 图标名 → 模板名映射（DB 存的 lucide 风格名 → 模板名） */
const ICON_TEMPLATE_MAP: Record<string, string> = {
  'layout-dashboard': 'dashboard',
  users: 'user',
  'shield-check': 'role',
  'building-2': 'building',
  briefcase: 'box',
  blocks: 'box',
  'link-2': 'tree',
  package: 'box',
  'bar-chart-3': 'chart',
  plug: 'box',
  plus: 'add',
  pencil: 'edit',
  'trash-2': 'delete',
  'alert-triangle': 'warning',
  'circle-check': 'success',
}

/** 未登记图标名的兜底模板（保证有图形，不会空白） */
const FALLBACK_TEMPLATE = 'menu'

const warned = new Set<string>()

function escapeXml(s: string): string {
  return s.replace(/&/g, '&').replace(/</g, '<').replace(/>/g, '>')
}

const svg = computed(() => {
  const mapped = ICON_TEMPLATE_MAP[props.name]
  const templateName = mapped ?? (TEMPLATES[props.name] ? props.name : FALLBACK_TEMPLATE)
  const template = TEMPLATES[templateName]
  if (!template) return ''
  if (!mapped && !TEMPLATES[props.name] && !warned.has(props.name)) {
    warned.add(props.name)
    console.warn(`[BaseIcon] 未登记的图标名 "${props.name}"，已兜底渲染 "${FALLBACK_TEMPLATE}"。`)
  }
  return template
    .replace(/\{\{size\}\}/g, String(props.size))
    .replace(/\{\{color\}\}/g, escapeXml(props.color))
    .replace(/\{\{strokeWidth\}\}/g, String(props.strokeWidth))
})
</script>

<template>
  <span class="base-icon" v-html="svg" />
</template>

<style scoped>
.base-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  line-height: 0;
  vertical-align: middle;
}
.base-icon :deep(svg) {
  display: block;
}
</style>
