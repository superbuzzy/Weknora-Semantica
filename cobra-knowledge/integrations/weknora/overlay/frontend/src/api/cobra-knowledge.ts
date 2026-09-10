import axios from 'axios'

export type GraphViewMode = 'entity' | 'ontology'

export interface CobraGraphNode {
  id: string
  label: string
  kind: string
  group?: string
  subtitle?: string
  metadata?: Record<string, unknown>
}

export interface CobraGraphEdge {
  id: string
  source: string
  target: string
  label?: string
  kind?: string
  metadata?: Record<string, unknown>
}

export interface CobraGraphMeta {
  view: GraphViewMode | string
  knowledge_base_id: string
  total_nodes: number
  returned_nodes: number
  returned_edges: number
  truncated: boolean
  ontology_id?: string
  ontology_version?: string
}

export interface CobraGraphView {
  nodes: CobraGraphNode[]
  edges: CobraGraphEdge[]
  meta: CobraGraphMeta
}

const BASE_URL = (import.meta.env.VITE_COBRA_KNOWLEDGE_API_BASE || '/cobra-knowledge').replace(/\/$/, '')

const client = axios.create({
  baseURL: BASE_URL,
  timeout: 15000,
  headers: { 'Content-Type': 'application/json' },
})

client.interceptors.request.use((config) => {
  const token = localStorage.getItem('weknora_token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  const tenantID = localStorage.getItem('weknora_selected_tenant_id')
  if (tenantID) config.headers['X-Tenant-ID'] = tenantID
  config.headers['Accept-Language'] = localStorage.getItem('locale') || 'zh-CN'
  return config
})

export async function getCobraKnowledgeGraph(kbID: string, view: GraphViewMode, limit = 160): Promise<CobraGraphView> {
  const response = await client.get<CobraGraphView>(`/api/v1/knowledge-bases/${encodeURIComponent(kbID)}/graph`, {
    params: { view, limit },
  })
  return response.data
}
