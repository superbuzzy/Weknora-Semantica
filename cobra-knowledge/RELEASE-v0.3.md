# CobraKnowledge v0.3.0

v0.3 把“本体图”从后台语义资产推进到 WeKnora 可视化界面，同时保持 WeKnora 和 CobraKnowledge 的存储、发布、升级边界。

## 本版重点

1. 新增统一 `GraphView` DTO，实体图和本体图共用节点/边/元数据契约。
2. 新增 `cobra-graph-api`：
   - `view=entity` 只读 WeKnora 现有 Neo4j；
   - `view=ontology` 读取当前 KB 绑定的 CobraKnowledge Ontology。
3. 本体图可视化包含 Class、DataProperty、继承关系、Class-Property 关系和 ObjectRelation。
4. WeKnora Graph Settings 新增 `GraphExplorer`，通过“实体图 / 本体图”按钮即时切换，同一画布渲染。
5. 增加 KB -> Ontology 外部绑定文件，同一本体可复用到多个知识库。
6. Graph API 权限默认委托 WeKnora RBAC，避免本体接口绕过原系统权限。
7. WeKnora 集成采用派生 overlay：官方上游目录不修改，拉取更新后重新应用小补丁。

## 兼容策略

- WeKnora：零业务内核修改；前端只有一个小补丁和两个新增 overlay 文件。
- Semantica：仍为方法参考，运行时零依赖。
- Neo4j：实体图只读，不改变 WeKnora Label/Property/Relationship Schema。
- Ontology：仍使用 CobraKnowledge 自研模型；可视化只消费统一 DTO。

## API

```text
GET /healthz
GET /api/v1/knowledge-bases/{kb_id}/graph?view=entity&limit=160
GET /api/v1/knowledge-bases/{kb_id}/graph?view=ontology
```

## 验证

- `go test ./...`：通过；
- `go vet ./...`：通过；
- `make build`：三个 Go 二进制均构建通过；
- `cobra-graph-api`：`/healthz` 与 `view=ontology` 冒烟测试通过；
- WeKnora overlay：在干净的 0.8.0 源码快照上 dry-run 与 application 通过；
- Overlay 新增 TypeScript 脚本：TypeScript 5.8 `transpileModule` 语法检查通过；
- 当前执行环境安装 WeKnora 前端 npm 依赖时发生网络/执行超时，因此未把完整 `npm build` 标记为已验证；应在正常前端依赖环境补跑官方构建命令。

完整部署说明见 `docs/GRAPH_VISUALIZATION.md`。
