# 图标系统

## 概述

本项目所有图标统一通过 [frontend-icon-skill](../../../vibecoding-portal/skills/vibeCoding/frontend/frontend-icon-skill/SKILL.md) 生成。

frontend-icon-skill 提供 15+ 内置 SVG 模板（menu/action/status/arrow 四类），参数化生成（颜色/尺寸/粗细），无需引入任何第三方图标库。

## 在项目中的使用

### 方式一：组件式（推荐）

直接用 `<BaseIcon name="dashboard" />`：

```vue
<script setup>
import { BaseIcon } from '@/components/icons/BaseIcon.vue'
</script>

<template>
  <BaseIcon name="dashboard" :size="20" />
  <BaseIcon name="user" color="#6366f1" />
  <BaseIcon name="success" />
</template>
```

**name 必须在 frontend-icon-skill 模板清单内**（详见 [template-list.md](../../../vibecoding-portal/skills/vibeCoding/frontend/frontend-icon-skill/references/template-list.md)）。

### 方式二：直接生成 SVG 字符串

```typescript
import { generateIcon } from 'frontend-icon-skill'
const svg = generateIcon('menu', { size: 16, color: '#4f46e5' })
```

## 生成静态文件到 static/icons/

适用于菜单 icon 字段、后端返回 icon name 后端读取 SVG 文件的场景：

```typescript
import { generateIconSet } from 'frontend-icon-skill'

await generateIconSet(
  ['dashboard', 'user', 'role', 'menu', 'tree', 'building', 'box', 'chart'],
  { size: 24, color: 'currentColor', strokeWidth: 2 },
  './static/icons'
)
```

输出到 `static/icons/dashboard.svg` 等文件，可被后端 `icon` 字段引用。

## 添加新图标

### 方式 A：扩展 frontend-icon-skill 模板库（推荐）

在 `vibecoding-portal/skills/vibeCoding/frontend/frontend-icon-skill/icon-templates/<分类>/<name>.svg.template` 加新模板（参考已有模板格式 + 占位符 `{{size}}`/`{{color}}`/`{{strokeWidth}}`），同时在 `src/registry.ts` 加映射。

### 方式 B：调用 icon-image-catch-skill 抓取

如果图标不在 lucide / iconify 等开源库，由 frontend-icon-skill 自动 fallback 到 icon-image-catch-skill 抓取 PNG/SVG 源后转 SVG 模板。

## 不做

- **不引入第三方图标库**（lucide-vue-next / @iconify/vue）：与 vue-admin-skill 零第三方依赖原则冲突
- **不写内联 SVG**：所有 SVG path 都通过 frontend-icon-skill 管理
- **不维护多套图标系统**：vue-admin-skill 仅依赖 frontend-icon-skill 一个入口