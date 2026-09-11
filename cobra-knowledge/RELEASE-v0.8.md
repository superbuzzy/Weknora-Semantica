# LeeClaw v0.8 Release

## 主题

**Skill Runtime + Knowledge Runtime**

v0.8 把 v0.7 的多人 Workspace / Knowledge 管理骨架真正接入 OpenClaw Agent 主链。

## 主要变化

1. OpenViking Skill 从管理资产升级为当前 Agent Turn 的动态 Runtime：语义发现、阈值筛选、L2 `SKILL.md` 加载与上下文注入。
2. Skill 选择固定为“个人 Skill 优先、共享 Skill 次之，同作用域按 score 排序”。
3. Skill `allowed-tools` 与 OpenClaw `PluginHookToolAuthority` 取交集，并通过 `toolsAllow` 只收窄当前 Turn 工具面；显式空列表是 deny-all。
4. 新增 `leeclaw_context_retrieve` / `leeclaw_context_get_evidence` 两个 Agent 原生 Knowledge Tool。
5. 新增 `internal/runtimecontext`，在线复用 Ontology → Semantic Catalog → Planner → Retriever → Arbiter → Context Pack。
6. Workspace 可声明 `knowledgeBaseIds`；Retrieve 和 Evidence 都在服务端执行 KB Scope，Evidence 读取 chunk 后再次核验 `knowledge_base_id`。
7. v0.7 profile-wide Chat Workspace 语义被 Session-bound Workspace 直接替换：已有 Session 不随页面切换漂移，新/reset Session 使用新的默认 Workspace。
8. OpenViking Session ID 使用 Workspace + profile + OpenClaw session identity 派生，reset 不复用旧 Memory Session。
9. `graphApiBaseUrl` 被 `coreApiBaseUrl` 直接替换；同一个 Core API 承载 Graph + Runtime Context。
10. Runtime API 使用独立内部 Bearer Token，浏览器 CORS 不开放可信 Runtime Header。
11. RAG fallback 保留结构化/实时来源 Gap，不把文档相关性误报成实时事实完整。
12. 三个 upstream submodule 不修改。

## 验证边界

- 每个 Turn 只自动激活一个 Skill；
- 不自动执行 Skill 辅助脚本/文件；
- 动态通用 Entity / Business Retriever 尚未完成；
- 多 KB Ontology merge 尚未启用；
- Workspace Registry / Session state / Audit / Ontology FSRegistry 仍为单 Gateway/单实例基线；
- Promotion、HA、全链路 Trace/Eval 和完整生产 CI 尚未完成；
- 只有在依赖完整的 CI 环境真实跑通后才可标记完整 OpenClaw pnpm bundle / E2E 通过。
