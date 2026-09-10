# v0.4 三图可视化：实体图 / 本体图切换

## 1. 目标

v0.3 将本体图正式作为可视化知识资产接入 WeKnora 页面。当前官方 WeKnora 前端没有独立的 Neo4j 实体图可视化 REST 页面，因此 CobraKnowledge 在知识库“图谱”设置区域补充一个统一 `GraphExplorer`：

- 默认显示 **实体图**：只读 WeKnora 当前 Neo4j GraphRAG 数据；
- 点击 **本体图**：显示该知识库绑定的 CobraKnowledge Ontology；
- 两种视图使用同一个 `GraphView` 数据契约和同一前端画布；
- Wiki 图仍由 WeKnora 原 Wiki Browser 负责，不与本次实体/本体切换混合。

因此三张图的 UI 边界为：

```text
Wiki Browser             Graph Settings / GraphExplorer
    │                               │
    ▼                        ┌──────┴──────┐
 Wiki 图                    实体图        本体图
WeKnora Wiki API          WeKnora Neo4j  Cobra Ontology
```

## 2. 为什么不直接改 WeKnora Neo4j Repository

本体图属于 CobraKnowledge 治理资产，实体图属于 WeKnora GraphRAG 运行资产。v0.3 不把两类数据强行写成同一 Neo4j Schema，也不修改 WeKnora `RetrieveGraphRepository`。

CobraKnowledge 新增只读 `Neo4jHTTPSource`，按照 WeKnora 当前 `ENTITY<kb_id>` Label 规则读取实体图，再转换成统一 `GraphView`。本体图由 `BuildOntologyView` 从 Ontology JSON/Registry 投影为同一 DTO。

如果以后 WeKnora 提供正式实体图可视化 API，只需要替换 `EntitySource` Adapter，前端无需修改。

## 3. GraphView 契约

```json
{
  "nodes": [
    {
      "id": "class:line",
      "label": "线路",
      "kind": "class",
      "metadata": {}
    }
  ],
  "edges": [
    {
      "id": "relation:supplies:line:area",
      "source": "class:line",
      "target": "class:area",
      "label": "供电",
      "kind": "object_relation"
    }
  ],
  "meta": {
    "view": "ontology",
    "knowledge_base_id": "kb-id",
    "total_nodes": 10,
    "returned_nodes": 10,
    "returned_edges": 8,
    "truncated": false,
    "ontology_version": "1.0.0"
  }
}
```

本体图投影规则：

- `OntologyClass` -> `kind=class` 节点；
- `DataProperty` -> `kind=property` 节点；
- `ParentIDs` -> `subclass_of` 边；
- `DomainIDs` -> Class 到 Property 的 `has_property` 边；
- `ObjectRelation` -> Domain Class 到 Range Class 的带业务关系名称的边。

## 4. Graph API

启动：

```bash
export COBRA_NEO4J_URL=http://neo4j:7474
export COBRA_NEO4J_USER=neo4j
export COBRA_NEO4J_PASSWORD='***'
export COBRA_ONTOLOGY_REGISTRY_ROOT=/app/data/ontology-registry
export COBRA_WEKNORA_BASE_URL=http://weknora:8080
export COBRA_GRAPH_AUTH_MODE=weknora

go run ./cmd/graph-api -listen :8090
```

接口：

```text
GET /api/v1/knowledge-bases/{kb_id}/graph?view=entity&limit=160
GET /api/v1/knowledge-bases/{kb_id}/graph?view=ontology
GET /healthz
```

实体图展示默认限制为 160 个节点，最大 500，避免浏览器对大图做全量力导向布局。`meta.truncated=true` 时前端给出截断提示。

## 5. 本体与知识库绑定

v0.4 不再由 Graph API 读取 `ontology-bindings.json`。绑定进入 Ontology Registry，支持 active / pinned 两种模式：

```text
KB -> ontology_id -> active published version
KB -> ontology_id@version   (pinned)
```

本体 payload 注册后不可变；发布和回滚只修改 Registry Manifest 与 active 指针。详见 `docs/ONTOLOGY_REGISTRY.md`。

旧 `configs/ontology-bindings.example.json` 仅作为 v0.3 迁移参考，不再是生产运行配置。

## 6. 权限

生产默认 `COBRA_GRAPH_AUTH_MODE=weknora`。Graph API 收到页面请求后，将以下头转发给 WeKnora：

- `Authorization`
- `X-Tenant-ID`
- `Accept-Language`

随后调用：

```text
GET /api/v1/knowledge-bases/{kb_id}
```

只有 WeKnora 返回成功时才读取实体图/本体图。这使图谱查看沿用 WeKnora 原 RBAC 与跨空间权限，不在 CobraKnowledge 复制权限规则。

开发环境可显式设置 `COBRA_GRAPH_AUTH_MODE=off`，启动时会输出警告。

## 7. WeKnora Overlay

不在 upstream checkout 上直接改代码。v0.3 起提供，v0.4 原样沿用：

```text
integrations/weknora/
├── overlay/
│   └── frontend/src/
│       ├── api/cobra-knowledge.ts
│       └── views/knowledge/settings/GraphExplorer.vue
├── patches/
│   └── 0001-add-ontology-graph-switcher.patch
└── apply-overlay.sh
```

执行：

```bash
./integrations/weknora/apply-overlay.sh \
  ../upstream/weknora \
  ../build/weknora-v0.4
```

脚本复制出派生构建树后再应用补丁。`upstream/weknora` 不产生修改，因此以后仍可正常 `git pull`。如果官方 `GraphSettings.vue` 发生破坏性变化，补丁 dry-run 会直接失败，要求人工更新这一处适配，而不是静默覆盖官方新逻辑。

## 8. 反向代理

推荐同源：

```nginx
location /cobra-knowledge/ {
    proxy_pass http://cobra-graph-api:8090/;
    proxy_set_header Authorization $http_authorization;
    proxy_set_header X-Tenant-ID $http_x_tenant_id;
    proxy_set_header Accept-Language $http_accept_language;
}
```

前端默认 `VITE_COBRA_KNOWLEDGE_API_BASE=/cobra-knowledge`。同源部署时不需要 CORS。

## 9. 当前仍保持的边界

- 不把 Wiki 图并入 GraphExplorer；Wiki 图继续使用官方成熟 Wiki Browser。
- 不把正式本体强制写入 WeKnora Neo4j Schema；GraphView 是显示契约，不是存储耦合。
- 不做全量百万节点图可视化；大图后续应采用搜索/局部展开，而不是浏览器一次性渲染。
- 不修改 WeKnora GraphRAG 的建图与检索逻辑。
