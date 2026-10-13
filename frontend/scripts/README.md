# 主题色生成脚本（vue-theme-skill 接入）

**默认主题：copper（铜色 #c08457）** —— 用户指定 java-vue-admin dogfooding 主题色。

## 用法

```bash
cd D:/projects/java-vue-admin/frontend

# 生成 copper 主题（默认）
node scripts/generate-tokens.cjs

# 生成其他主题
node scripts/generate-tokens.cjs blue
node scripts/generate-tokens.cjs green
node scripts/generate-tokens.cjs purple
node scripts/generate-tokens.cjs red
```

输出：`src/styles/tokens.css`（覆盖现有文件）。

## 可用主题

| 主题 | hue | 主色 hex |
|---|---|---|
| blue | 211 | `#3b82f6` |
| green | 152 | `#1bc16d` |
| purple | 262 | `#9146f0` |
| red | 0 | `#ea3e3e` |
| orange | 33 | `#ff9d28` |
| cyan | 180 | `#0fb2bf` |
| pink | 330 | `#f048a8` |
| slate | 220 | `#677387` |
| brown | 26 | `#c08160` |
| **copper** | **26** | **#c08457** ← 默认 |

## 原理

`generate-tokens.cjs`（从 `vue-theme-skill/templates/scripts/` 拷来）按 HSL 算法生成 9 级色阶（50-950），输出标准化的 CSS 变量：

- `--color-primary-{50~950}`：色阶
- `--color-primary` / `-light` / `-dark`：简写
- `--color-success` / `--color-warning` / `--color-danger`：语义色
- `--color-bg` / `-surface` / `-text` / `-border`：中性色
- `--space-{1~16}` / `--font-{2xs~4xl}` / `--radius-{sm~full}`：尺寸/字号/圆角
- `--height-{button,input,card,table-row}-{sm,md,lg}`：组件高度
- `--icon-{xs~xl}` / `--shadow-{sm~xl}`：图标/阴影

所有命名对齐 `uniapp-theme-skill`，跨技能可复用。

## 修改主题色的流程

1. 跑 `node scripts/generate-tokens.cjs <theme>`
2. 重启 vite（`HMR` 应自动生效，但深色覆盖需要确认）
3. 浏览器 DevTools 查 `getComputedStyle(:root).--color-primary` 看新色值
4. 不要手改 `tokens.css`（generator 会覆盖）
