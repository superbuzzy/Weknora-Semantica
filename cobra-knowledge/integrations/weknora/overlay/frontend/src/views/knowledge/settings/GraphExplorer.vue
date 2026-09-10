<template>
  <div class="cobra-graph-explorer">
    <div class="graph-explorer-header">
      <div>
        <div class="graph-explorer-title">图谱预览</div>
        <div class="graph-explorer-desc">在同一画布中切换查看 Neo4j 实体图和 CobraKnowledge 本体图。</div>
      </div>
      <div class="graph-explorer-actions">
        <div class="view-switcher" role="group" aria-label="图谱类型">
          <t-button size="small" :theme="viewMode === 'entity' ? 'primary' : 'default'" @click="viewMode = 'entity'">
            实体图
          </t-button>
          <t-button size="small" :theme="viewMode === 'ontology' ? 'primary' : 'default'" @click="viewMode = 'ontology'">
            本体图
          </t-button>
        </div>
        <t-button size="small" variant="outline" :loading="loading" @click="loadGraph">刷新</t-button>
      </div>
    </div>

    <div v-if="!knowledgeBaseID" class="graph-state graph-state--empty">
      当前页面没有知识库 ID，进入具体知识库后可查看图谱。
    </div>
    <div v-else-if="errorMessage" class="graph-state graph-state--error">
      {{ errorMessage }}
    </div>
    <div v-else-if="loading && !graph" class="graph-state">正在加载图谱…</div>
    <div v-else-if="graph && graph.nodes.length === 0" class="graph-state graph-state--empty">
      {{ viewMode === 'entity' ? '当前知识库还没有实体图数据。' : '当前知识库还没有绑定本体。' }}
    </div>

    <template v-else-if="graph">
      <div class="graph-summary">
        <span>{{ viewMode === 'entity' ? '实体图' : '本体图' }}</span>
        <span>{{ graph.meta.returned_nodes }} 个节点</span>
        <span>{{ graph.meta.returned_edges }} 条关系</span>
        <span v-if="viewMode === 'ontology' && graph.meta.ontology_version">本体版本 {{ graph.meta.ontology_version }}</span>
        <span v-if="graph.meta.truncated" class="graph-summary-warning">图谱较大，当前仅展示前 {{ graph.meta.returned_nodes }} 个节点</span>
      </div>

      <div class="graph-workspace">
        <div class="graph-canvas-wrap">
          <svg class="graph-canvas" viewBox="0 0 1000 600" role="img" :aria-label="viewMode === 'entity' ? '实体关系图' : '本体关系图'">
            <defs>
              <marker id="cobra-arrow" viewBox="0 0 10 10" refX="9" refY="5" markerWidth="6" markerHeight="6" orient="auto-start-reverse">
                <path d="M 0 0 L 10 5 L 0 10 z" class="graph-arrow" />
              </marker>
            </defs>

            <g class="edge-layer">
              <g v-for="edge in visibleEdges" :key="edge.id">
                <line
                  v-if="positions[edge.source] && positions[edge.target]"
                  :x1="positions[edge.source].x"
                  :y1="positions[edge.source].y"
                  :x2="positions[edge.target].x"
                  :y2="positions[edge.target].y"
                  class="graph-edge"
                  marker-end="url(#cobra-arrow)"
                />
                <text
                  v-if="showEdgeLabels && edge.label && positions[edge.source] && positions[edge.target]"
                  :x="(positions[edge.source].x + positions[edge.target].x) / 2"
                  :y="(positions[edge.source].y + positions[edge.target].y) / 2 - 4"
                  class="edge-label"
                >{{ edge.label }}</text>
              </g>
            </g>

            <g class="node-layer">
              <template v-for="node in graph.nodes" :key="node.id">
                <g
                  v-if="positions[node.id]"
                  class="graph-node"
                  :class="[`graph-node--${node.kind}`, { 'graph-node--selected': selectedNode?.id === node.id }]"
                  :transform="`translate(${positions[node.id].x}, ${positions[node.id].y})`"
                  tabindex="0"
                  role="button"
                  :aria-label="node.label"
                  @click="selectedNode = node"
                  @keydown.enter="selectedNode = node"
                >
                  <circle :r="nodeRadius(node.kind)" />
                  <text :y="nodeRadius(node.kind) + 16" text-anchor="middle" class="node-label">{{ displayLabel(node.label) }}</text>
                  <title>{{ node.label }}{{ node.subtitle ? ` — ${node.subtitle}` : '' }}</title>
                </g>
              </template>
            </g>
          </svg>
        </div>

        <aside class="graph-detail" v-if="selectedNode">
          <div class="graph-detail-head">
            <div>
              <div class="graph-detail-kind">{{ kindLabel(selectedNode.kind) }}</div>
              <div class="graph-detail-title">{{ selectedNode.label }}</div>
            </div>
            <t-button size="small" variant="text" @click="selectedNode = null">关闭</t-button>
          </div>
          <div v-if="selectedNode.subtitle" class="graph-detail-subtitle">{{ selectedNode.subtitle }}</div>
          <dl class="graph-detail-list">
            <template v-for="item in metadataRows" :key="item.key">
              <dt>{{ item.key }}</dt>
              <dd>{{ item.value }}</dd>
            </template>
          </dl>
        </aside>
      </div>
    </template>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { getCobraKnowledgeGraph, type CobraGraphEdge, type CobraGraphNode, type CobraGraphView, type GraphViewMode } from '@/api/cobra-knowledge'

const route = useRoute()
const knowledgeBaseID = computed(() => String(route.params.kbId || ''))
const viewMode = ref<GraphViewMode>('entity')
const loading = ref(false)
const errorMessage = ref('')
const graph = ref<CobraGraphView | null>(null)
const selectedNode = ref<CobraGraphNode | null>(null)
const positions = reactive<Record<string, { x: number; y: number; vx: number; vy: number }>>({})
let animationFrame = 0

const visibleEdges = computed<CobraGraphEdge[]>(() => graph.value?.edges || [])
const showEdgeLabels = computed(() => (graph.value?.edges.length || 0) <= 80)
const metadataRows = computed(() => {
  const metadata = selectedNode.value?.metadata || {}
  return Object.entries(metadata).map(([key, value]) => ({
    key,
    value: formatMetadata(value),
  })).slice(0, 12)
})

function formatMetadata(value: unknown): string {
  if (Array.isArray(value)) return value.map(String).join('、') || '-'
  if (value && typeof value === 'object') return JSON.stringify(value)
  return value === null || value === undefined || value === '' ? '-' : String(value)
}

function displayLabel(label: string): string {
  return label.length > 16 ? `${label.slice(0, 15)}…` : label
}

function kindLabel(kind: string): string {
  if (kind === 'class') return '本体类型'
  if (kind === 'property') return '本体属性'
  return '实体'
}

function nodeRadius(kind: string): number {
  if (kind === 'class') return 14
  if (kind === 'property') return 10
  return 11
}

async function loadGraph() {
  if (!knowledgeBaseID.value) return
  loading.value = true
  errorMessage.value = ''
  selectedNode.value = null
  stopSimulation()
  try {
    graph.value = await getCobraKnowledgeGraph(knowledgeBaseID.value, viewMode.value, 160)
    await nextTick()
    initializeLayout()
  } catch (error: any) {
    graph.value = null
    const backendMessage = error?.response?.data?.error
    errorMessage.value = backendMessage || error?.message || '图谱加载失败，请检查 CobraKnowledge Graph API 配置。'
  } finally {
    loading.value = false
  }
}

function initializeLayout() {
  const data = graph.value
  for (const key of Object.keys(positions)) delete positions[key]
  if (!data?.nodes.length) return

  const cx = 500
  const cy = 300
  const radius = Math.min(240, 100 + data.nodes.length * 2.2)
  data.nodes.forEach((node, index) => {
    const angle = (Math.PI * 2 * index) / data.nodes.length
    const ring = radius * (0.72 + (index % 5) * 0.055)
    positions[node.id] = {
      x: cx + Math.cos(angle) * ring,
      y: cy + Math.sin(angle) * ring,
      vx: 0,
      vy: 0,
    }
  })
  startSimulation()
}

function startSimulation() {
  stopSimulation()
  let alpha = 1
  const tick = () => {
    const data = graph.value
    if (!data || alpha < 0.025) {
      animationFrame = 0
      return
    }
    const nodes = data.nodes

    // Bounded force simulation. The API caps the interactive view at 160 nodes.
    for (let i = 0; i < nodes.length; i++) {
      const a = positions[nodes[i].id]
      if (!a) continue
      for (let j = i + 1; j < nodes.length; j++) {
        const b = positions[nodes[j].id]
        if (!b) continue
        const dx = b.x - a.x
        const dy = b.y - a.y
        const dist2 = Math.max(dx * dx + dy * dy, 90)
        if (dist2 > 70000) continue
        const force = (3800 * alpha) / dist2
        const dist = Math.sqrt(dist2)
        const fx = (dx / dist) * force
        const fy = (dy / dist) * force
        a.vx -= fx
        a.vy -= fy
        b.vx += fx
        b.vy += fy
      }
    }

    for (const edge of data.edges) {
      const a = positions[edge.source]
      const b = positions[edge.target]
      if (!a || !b) continue
      const dx = b.x - a.x
      const dy = b.y - a.y
      const dist = Math.max(Math.sqrt(dx * dx + dy * dy), 1)
      const target = edge.kind === 'has_property' ? 92 : 135
      const force = (dist - target) * 0.0035 * alpha
      const fx = (dx / dist) * force
      const fy = (dy / dist) * force
      a.vx += fx
      a.vy += fy
      b.vx -= fx
      b.vy -= fy
    }

    for (const node of nodes) {
      const p = positions[node.id]
      if (!p) continue
      p.vx += (500 - p.x) * 0.0007 * alpha
      p.vy += (300 - p.y) * 0.0007 * alpha
      p.vx *= 0.87
      p.vy *= 0.87
      p.x = Math.max(28, Math.min(972, p.x + p.vx))
      p.y = Math.max(28, Math.min(572, p.y + p.vy))
    }

    alpha *= 0.965
    animationFrame = requestAnimationFrame(tick)
  }
  animationFrame = requestAnimationFrame(tick)
}

function stopSimulation() {
  if (animationFrame) cancelAnimationFrame(animationFrame)
  animationFrame = 0
}

watch(viewMode, () => loadGraph())
watch(knowledgeBaseID, () => loadGraph())
onMounted(loadGraph)
onBeforeUnmount(stopSimulation)
</script>

<style scoped lang="less">
.cobra-graph-explorer {
  width: 100%;
  border: 1px solid var(--td-component-stroke);
  border-radius: 10px;
  background: var(--td-bg-color-container);
  overflow: hidden;
}

.graph-explorer-header {
  min-height: 64px;
  padding: 14px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  border-bottom: 1px solid var(--td-component-stroke);
}
.graph-explorer-title { font-size: 15px; font-weight: 600; color: var(--td-text-color-primary); }
.graph-explorer-desc { margin-top: 4px; font-size: 12px; color: var(--td-text-color-secondary); }
.graph-explorer-actions, .view-switcher { display: flex; align-items: center; gap: 8px; }
.graph-summary { display: flex; gap: 16px; flex-wrap: wrap; padding: 10px 16px; font-size: 12px; color: var(--td-text-color-secondary); border-bottom: 1px solid var(--td-component-stroke); }
.graph-summary-warning { color: var(--td-warning-color); }
.graph-state { min-height: 220px; display: flex; align-items: center; justify-content: center; padding: 24px; color: var(--td-text-color-secondary); }
.graph-state--error { color: var(--td-error-color); }
.graph-workspace { display: flex; min-height: 520px; }
.graph-canvas-wrap { min-width: 0; flex: 1; background: var(--td-bg-color-secondarycontainer); }
.graph-canvas { width: 100%; height: 520px; display: block; }
.graph-detail { width: 280px; flex: 0 0 280px; padding: 16px; border-left: 1px solid var(--td-component-stroke); background: var(--td-bg-color-container); overflow: auto; }
.graph-detail-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 8px; }
.graph-detail-kind { font-size: 11px; color: var(--td-text-color-placeholder); }
.graph-detail-title { margin-top: 2px; font-size: 16px; font-weight: 600; word-break: break-all; }
.graph-detail-subtitle { margin-top: 12px; font-size: 12px; color: var(--td-text-color-secondary); line-height: 1.6; }
.graph-detail-list { margin: 16px 0 0; display: grid; grid-template-columns: 88px minmax(0, 1fr); gap: 8px 10px; font-size: 12px; }
.graph-detail-list dt { color: var(--td-text-color-placeholder); }
.graph-detail-list dd { margin: 0; color: var(--td-text-color-secondary); word-break: break-all; }
.graph-edge { stroke: var(--td-component-stroke); stroke-width: 1.3; opacity: .78; }
.graph-arrow { fill: var(--td-text-color-placeholder); }
.edge-label { fill: var(--td-text-color-placeholder); font-size: 10px; paint-order: stroke; stroke: var(--td-bg-color-secondarycontainer); stroke-width: 3px; }
.graph-node { cursor: pointer; outline: none; }
.graph-node circle { fill: var(--td-brand-color-light); stroke: var(--td-brand-color); stroke-width: 1.8; transition: stroke-width .15s ease, opacity .15s ease; }
.graph-node--class circle { fill: var(--td-success-color-light); stroke: var(--td-success-color); }
.graph-node--property circle { fill: var(--td-warning-color-light); stroke: var(--td-warning-color); }
.graph-node--selected circle, .graph-node:focus circle { stroke-width: 4; }
.node-label { fill: var(--td-text-color-primary); font-size: 11px; pointer-events: none; paint-order: stroke; stroke: var(--td-bg-color-secondarycontainer); stroke-width: 3px; }

@media (max-width: 900px) {
  .graph-explorer-header { align-items: flex-start; flex-direction: column; }
  .graph-workspace { flex-direction: column; }
  .graph-detail { width: auto; flex-basis: auto; border-left: 0; border-top: 1px solid var(--td-component-stroke); }
}
</style>
