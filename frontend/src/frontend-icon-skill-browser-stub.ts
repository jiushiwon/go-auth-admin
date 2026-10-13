/**
 * frontend-icon-skill/src/index.ts 的浏览器 stub（v2：绝对路径 import）
 *
 * 背景：frontend-icon-skill 的 src/index.ts 顶层 import 'node:fs/promises'（CLI 工具
 *   generateIconToFile 用），浏览器端不能用 fs。
 *   BaseIcon.vue 内部 import { generateIcon, listTemplates } from '../src/index.ts'，被加载时
 *   沿相对路径 import ../src/index.ts → 浏览器报错 fs/promises。
 *
 * 修法：vite resolve.alias 把 'frontend-icon-skill/src/index.ts' 替换为本 stub。stub 用
 *   @fs/ 绝对路径绕开 frontend-icon-skill 的 package.json exports 限制（exports 字段没列
 *   './src/*'），直接从 .ts 文件拿 generateIcon 和 listTemplates。
 */
import { generateIcon } from '/@fs/D:/projects/java-vue-admin/frontend/node_modules/frontend-icon-skill/src/generator.ts'
import { listTemplates } from '/@fs/D:/projects/java-vue-admin/frontend/node_modules/frontend-icon-skill/src/registry.ts'
import type { IconOptions } from '/@fs/D:/projects/java-vue-admin/frontend/node_modules/frontend-icon-skill/src/generator.ts'

export type { IconOptions }
export { generateIcon, listTemplates }
