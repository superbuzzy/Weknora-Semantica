# CobraKnowledge v0.4.0

## 主题

**Ontology Registry：把本体从文件级配置升级为可注册、发布、绑定、回滚、审计的一等资产，同时保持 WeKnora 上游低侵入。**

## 核心变化

1. 新增 `ontology.Registry` 存储无关接口与 `FSRegistry` 默认实现。
2. Ontology payload 注册后不可变，使用 SHA-256 校验历史版本完整性。
3. 新增 Manifest，独立维护 candidate/published 状态与 `active_version`。
4. KB Binding 支持：
   - `active`：自动跟随当前正式发布版本；
   - `pinned`：固定到指定已发布版本。
5. 发布新版本只移动 active 指针；回滚只重新激活历史 published 版本，不修改历史内容。
6. 注册、发布、激活、KB 绑定写入 append-only Audit Event。
7. v0.3 `FileOntologyBindings` 从运行链路移除，GraphView 改为通过 Registry Resolver 获取正式本体。
8. WeKnora 前端 GraphExplorer **不需要修改**，继续使用原 `view=entity|ontology` 契约。
9. Registry REST 管理接口使用独立 `X-Cobra-Admin-Token`，避免把 WeKnora 普通用户 Token 等同于本体治理权限。
10. Context MCP 可优先通过 `COBRA_ONTOLOGY_REGISTRY_ROOT + COBRA_ONTOLOGY_KB_ID` 获取正式版本；`COBRA_ONTOLOGY_FILE` 仅保留本地开发兼容入口。

## 对 WeKnora 上游的影响

v0.4 对 WeKnora **零新增侵入**：

- 不改 WeKnora Go 后端；
- 不改 GraphRAG 建图/检索；
- 不改 Neo4j Label/Relation Schema；
- 不新增 WeKnora 表；
- 不修改现有 v0.3 Overlay 文件和 GraphSettings 小补丁；
- 浏览器请求路径保持不变。

## 验证门禁

发布前执行：

```bash
go test ./...
go vet ./...
make build
```

并验证：

- Registry version immutability；
- publish / active / pinned；
- rollback；
- audit；
- Registry GraphView；
- Graph API 兼容原实体图/本体图接口；
- 普通 WeKnora Token 不能调用 Registry 管理 API。
