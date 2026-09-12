# LeeClaw v0.10.0 Release

## 主题：Experience Promotion

v0.10 将 Memory / Experience 到正式组织能力之间补成受控链路：候选 → 来源核验 → 去重/冲突 → 审核 → Gate → 发布 → 回滚。

### 已实现

- 新增 Candidate / Inspection / Review / Publication / Rollback 合同与状态机；
- 新增 Store 抽象和单实例 FSStore，带乐观版本控制；
- Knowledge Promotion：Chunk 证据复核、KB Scope、去重/冲突、WeKnora Manual Knowledge 发布与删除回滚；
- Skill Promotion：Session/Experience 真追溯、官方 Skill validate、Behavior Diff、Eval Gate、revision race check、共享 Skill 新建/更新/回滚；
- 独立内部 Promotion API + `LEECLAW_PROMOTION_TOKEN`；
- 独立 OpenClaw Promotion Plugin；admin/owner 才能 review/publish/rollback；
- Agent Tool Surface 保持只读 Knowledge Runtime，不增加正式发布/回滚工具；
- 正式 Knowledge/Skill 继续分别以 WeKnora/OpenViking 为唯一 Source of Truth；
- 活跃配置、兼容矩阵、验证脚本和插件 metadata 切换到 v0.10；
- 三个 upstream 源码继续零修改。

### 当前边界

FSStore 仅为单节点基线；完整共享持久化与 HA 在 v0.11。v0.10 只把 Eval Result 作为 Skill 发布门禁，完整 Trace/Eval/CI 在 v0.12。
