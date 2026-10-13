<!--
  接口文档（system/api）

  数据来源：后端 GET /api/api-registry，由反射从代码实时生成，
  因此增删接口、改结构体字段都会自动同步，不会像手写文档那样漂移。

  ★ 权限码：`system:api:view`（与后端 RequirePerm、AuthPermsAll、DB wg_sys_menu 三处一致）

  ★ 本页刻意不依赖 vue-* skill 组件，与 api-docs 页面策略一致，确保稳定渲染。
-->
<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { getApiRegistry, type ApiMeta, type ApiFieldMeta } from '@/api/apiRegistry';

const loading = ref(false);
const error = ref('');
const list = ref<ApiMeta[]>([]);
const keyword = ref('');
const methodFilter = ref('ALL');
const expanded = ref<Record<string, boolean>>({});

const METHODS = ['ALL', 'GET', 'POST', 'PUT', 'DELETE'];

async function load() {
  loading.value = true;
  error.value = '';
  try {
    const res = await getApiRegistry();
    list.value = res.list ?? [];
  } catch (e: any) {
    error.value = e?.message ?? '加载失败';
  } finally {
    loading.value = false;
  }
}

onMounted(load);

const filtered = computed(() => {
  const kw = keyword.value.trim().toLowerCase();
  return list.value.filter((x) => {
    if (methodFilter.value !== 'ALL' && x.method !== methodFilter.value) return false;
    if (!kw) return true;
    return (x.path + x.summary + x.moduleCn + (x.perm ?? '')).toLowerCase().includes(kw);
  });
});

const groups = computed(() => {
  const m = new Map<string, ApiMeta[]>();
  for (const x of filtered.value) {
    const arr = m.get(x.moduleCn) ?? [];
    arr.push(x);
    m.set(x.moduleCn, arr);
  }
  return [...m.entries()].map(([name, items]) => ({ name, items }));
});

function keyOf(x: ApiMeta) {
  return x.method + ' ' + x.path;
}
function toggle(x: ApiMeta) {
  const k = keyOf(x);
  expanded.value[k] = !expanded.value[k];
}
function isOpen(x: ApiMeta) {
  return !!expanded.value[keyOf(x)];
}

/** 把嵌套字段摊平，带缩进层级，便于表格展示 */
function flatten(fields?: ApiFieldMeta[], depth = 0): { f: ApiFieldMeta; depth: number }[] {
  if (!fields?.length) return [];
  const out: { f: ApiFieldMeta; depth: number }[] = [];
  for (const f of fields) {
    out.push({ f, depth });
    if (f.children?.length && depth < 1) {
      out.push(...flatten(f.children, depth + 1));
    }
  }
  return out;
}

const authLabel: Record<string, string> = {
  jwt: '需登录',
  app: '应用签名',
  public: '公开'
};

function openOpenAPI() {
  window.open('/api/api-registry/openapi', '_blank');
}
</script>

<template>
  <div class="page-api">
    <div class="api-head">
      <div>
        <h2>接口文档</h2>
        <p class="sub">
          由后端反射实时生成，共 {{ list.length }} 个接口、{{ groups.length }} 个模块。
          改代码即改文档，不会漂移。
        </p>
      </div>
      <button class="btn" @click="openOpenAPI">查看 OpenAPI 3.1</button>
    </div>

    <div class="toolbar">
      <input v-model="keyword" class="ipt" placeholder="搜索路径 / 说明 / 权限码" />
      <div class="methods">
        <button
          v-for="m in METHODS"
          :key="m"
          class="chip"
          :class="{ active: methodFilter === m }"
          @click="methodFilter = m"
        >
          {{ m }}
        </button>
      </div>
      <span class="count">命中 {{ filtered.length }} 条</span>
    </div>

    <div v-if="loading" class="state">加载中…</div>
    <div v-else-if="error" class="state err">{{ error }}</div>

    <div v-for="g in groups" :key="g.name" class="group">
      <h3 class="group-title">{{ g.name }} <span class="group-n">{{ g.items.length }}</span></h3>
      <table class="tbl">
        <thead>
          <tr>
            <th style="width: 74px">方法</th>
            <th style="width: 280px">路径</th>
            <th>说明</th>
            <th style="width: 190px">权限码</th>
            <th style="width: 72px">鉴权</th>
            <th style="width: 150px">入参 / 出参</th>
          </tr>
        </thead>
        <tbody>
          <template v-for="x in g.items" :key="keyOf(x)">
            <tr class="row" @click="toggle(x)">
              <td><span class="m" :class="'m-' + x.method.toLowerCase()">{{ x.method }}</span></td>
              <td class="path">{{ x.path }}</td>
              <td class="sum">{{ x.summary }}</td>
              <td class="perm">{{ x.perm || '—' }}</td>
              <td><span class="auth" :class="'a-' + x.auth">{{ authLabel[x.auth] || x.auth }}</span></td>
              <td class="types">
                <span class="t-req">{{ x.reqType }}</span>
                <span class="arrow">→</span>
                <span class="t-res">{{ x.respType }}</span>
              </td>
            </tr>
            <tr v-if="isOpen(x)" class="detail">
              <td colspan="6">
                <div class="detail-box">
                  <div class="meta-line">
                    <span><b>模块</b> {{ x.module }}（{{ x.moduleCn }}）</span>
                    <span v-if="x.pathVars?.length"><b>路径参数</b> {{ x.pathVars.join(', ') }}</span>
                    <span v-if="x.query?.length"><b>Query</b> {{ x.query.join(', ') }}</span>
                    <span><b>完整路径</b> {{ x.path }}</span>
                  </div>

                  <div class="two-col">
                    <div>
                      <h4>入参 {{ x.reqType !== 'void' ? `（${x.reqType}）` : '' }}</h4>
                      <table v-if="flatten(x.request).length" class="tbl inner">
                        <thead>
                          <tr><th>字段</th><th style="width: 90px">类型</th><th style="width: 60px">必填</th><th>说明</th></tr>
                        </thead>
                        <tbody>
                          <tr v-for="(it, i) in flatten(x.request)" :key="'req' + i">
                            <td :style="{ paddingLeft: 8 + it.depth * 16 + 'px' }">
                              {{ it.depth ? '└ ' : '' }}{{ it.f.name }}
                            </td>
                            <td>{{ it.f.type }}</td>
                            <td>{{ it.f.required ? '是' : '' }}</td>
                            <td>{{ it.f.desc || '' }}</td>
                          </tr>
                        </tbody>
                      </table>
                      <p v-else class="empty">无请求体</p>
                    </div>

                    <div>
                      <h4>出参 {{ x.respType !== 'void' ? `（${x.respType}）` : '' }}</h4>
                      <table v-if="flatten(x.respItem).length || flatten(x.response).length" class="tbl inner">
                        <thead>
                          <tr><th>字段</th><th style="width: 90px">类型</th><th>说明</th></tr>
                        </thead>
                        <tbody>
                          <tr v-for="(it, i) in flatten(x.respItem).length ? flatten(x.respItem) : flatten(x.response)" :key="'res' + i">
                            <td :style="{ paddingLeft: 8 + it.depth * 16 + 'px' }">
                              {{ it.depth ? '└ ' : '' }}{{ it.f.name }}
                            </td>
                            <td>{{ it.f.type }}</td>
                            <td>{{ it.f.desc || '' }}</td>
                          </tr>
                        </tbody>
                      </table>
                      <p v-else class="empty">无返回体</p>
                    </div>
                  </div>

                  <p v-if="x.respItem?.length" class="note">
                    出参为分页结构 PageResponse&lt;{{ x.respType.split('<')[1]?.replace('>', '') || 'item' }}&gt;，
                    上表列出的是 list 元素字段，外层另有 total / page / pageSize。
                  </p>
                </div>
              </td>
            </tr>
          </template>
        </tbody>
      </table>
    </div>

    <div v-if="!loading && !filtered.length" class="state">没有匹配的接口</div>
  </div>
</template>

<style scoped>
.page-api { padding: 4px 0 24px; }
.api-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 14px; }
.api-head h2 { margin: 0; font-size: 18px; font-weight: 600; color: #1f2937; }
.sub { margin: 6px 0 0; font-size: 13px; color: #6b7280; }
.btn { padding: 6px 14px; border: 1px solid #d1d5db; border-radius: 6px; background: #fff; color: #374151; cursor: pointer; font-size: 13px; }
.btn:hover { border-color: #2563eb; color: #2563eb; }

.toolbar { display: flex; align-items: center; gap: 12px; margin-bottom: 16px; flex-wrap: wrap; }
.ipt { width: 280px; padding: 6px 10px; border: 1px solid #d1d5db; border-radius: 6px; font-size: 13px; outline: none; }
.ipt:focus { border-color: #2563eb; }
.methods { display: flex; gap: 6px; }
.chip { padding: 4px 10px; border: 1px solid #d1d5db; border-radius: 999px; background: #fff; font-size: 12px; color: #4b5563; cursor: pointer; }
.chip.active { background: #2563eb; border-color: #2563eb; color: #fff; }
.count { font-size: 12px; color: #6b7280; }

.group { margin-bottom: 22px; }
.group-title { margin: 0 0 8px; font-size: 14px; font-weight: 600; color: #111827; }
.group-n { display: inline-block; margin-left: 4px; padding: 0 6px; border-radius: 8px; background: #f3f4f6; color: #6b7280; font-size: 12px; font-weight: 400; }

.tbl { width: 100%; border-collapse: collapse; font-size: 13px; background: #fff; border: 1px solid #e5e7eb; border-radius: 6px; overflow: hidden; }
.tbl th { background: #f9fafb; text-align: left; padding: 8px 10px; font-weight: 600; color: #4b5563; border-bottom: 1px solid #e5e7eb; white-space: nowrap; }
.tbl td { padding: 8px 10px; border-bottom: 1px solid #f3f4f6; color: #374151; vertical-align: top; }
.row { cursor: pointer; }
.row:hover { background: #f9fafb; }
.path { font-family: ui-monospace, Menlo, Consolas, monospace; font-size: 12px; color: #111827; }
.sum { color: #4b5563; }
.perm { font-family: ui-monospace, Menlo, Consolas, monospace; font-size: 12px; color: #6b7280; }
.types { font-size: 12px; white-space: nowrap; }
.t-req { color: #7c3aed; }
.t-res { color: #059669; }
.arrow { margin: 0 5px; color: #9ca3af; }

.m { display: inline-block; padding: 1px 7px; border-radius: 4px; font-size: 11px; font-weight: 600; color: #fff; }
.m-get { background: #2563eb; }
.m-post { background: #16a34a; }
.m-put { background: #d97706; }
.m-delete { background: #dc2626; }

.auth { display: inline-block; padding: 1px 6px; border-radius: 4px; font-size: 11px; background: #f3f4f6; color: #4b5563; }
.a-jwt { background: #dbeafe; color: #1d4ed8; }
.a-app { background: #fef3c7; color: #92400e; }
.a-public { background: #dcfce7; color: #15803d; }

.detail td { background: #fbfbfc; padding: 0; }
.detail-box { padding: 12px 14px; }
.meta-line { display: flex; gap: 18px; flex-wrap: wrap; font-size: 12px; color: #4b5563; margin-bottom: 10px; }
.meta-line b { color: #9ca3af; font-weight: 500; margin-right: 4px; }
.two-col { display: grid; grid-template-columns: 1fr 1fr; gap: 18px; }
@media (max-width: 900px) { .two-col { grid-template-columns: 1fr; } }
.two-col h4 { margin: 0 0 6px; font-size: 12px; font-weight: 600; color: #374151; }
.tbl.inner { font-size: 12px; }
.tbl.inner th { padding: 5px 8px; }
.tbl.inner td { padding: 5px 8px; }
.empty { margin: 0; font-size: 12px; color: #9ca3af; }
.note { margin: 10px 0 0; font-size: 12px; color: #92400e; }

.state { padding: 24px; text-align: center; color: #6b7280; font-size: 13px; }
.state.err { color: #dc2626; }
</style>
