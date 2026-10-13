import { get } from '@/utils/request';

/**
 * 接口元数据（后端由反射实时生成，源为代码本身）。
 *
 * 数据来源：GET /api/api-registry
 * 另有 OpenAPI 3.1：GET /api/api-registry/openapi
 *
 * 与手写的 api-docs 页面不同，这里的数据不会漂移：
 * 后端结构体增删字段、改 json tag 后会立刻反映到这里。
 */

export interface ApiFieldMeta {
  name: string;
  type: string;
  required: boolean;
  desc?: string;
  children?: ApiFieldMeta[];
}

export interface ApiMeta {
  method: string;
  path: string;
  module: string;
  moduleCn: string;
  perm?: string;
  /** 鉴权方式：jwt / app / public */
  auth: string;
  summary: string;
  pathVars: string[];
  query: string[];
  reqType: string;
  respType: string;
  request: ApiFieldMeta[];
  response: ApiFieldMeta[];
  respItem?: ApiFieldMeta[];
}

export interface ApiRegistryResponse {
  total: number;
  list: ApiMeta[];
}

/** 全量接口清单（需权限 system:api:view） */
// 注意：与 api/ 下其他文件保持一致，URL 自带 /api 前缀
// （本项目 VITE_API_BASE_URL 为空，不叠加前缀，否则会变成 /api/api/xxx）
export function getApiRegistry(): Promise<ApiRegistryResponse> {
  return get('/api/api-registry');
}
