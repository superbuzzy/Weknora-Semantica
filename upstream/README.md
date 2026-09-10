# Upstream 源码目录

LeeClaw 将三个核心上游按独立源码树管理，避免把上游代码与自研 Plugin / Adapter / Core 混在一起。

```text
upstream/
├── openclaw/     # Agent Runtime / Control UI / Plugin Runtime
├── openviking/   # Memory / Session / Experience / Skill
└── weknora/      # Knowledge Base / Document / Wiki / RAG / Graph / RBAC
```

三个目录使用 Git submodule 固定到经过 v0.5 兼容验证的发布版本：

| 目录 | 上游仓库 | 固定版本 | 固定 Commit |
|---|---|---|---|
| `upstream/openclaw` | `openclaw/openclaw` | `v2026.9.3` | `1391f7cd2d40ab5bbcf2f5f831d3a64f520e72d7` |
| `upstream/openviking` | `volcengine/OpenViking` | `v0.4.19` | `f3afef11637f2d7c11e4b1f36ed2f90630737cdc` |
| `upstream/weknora` | `Tencent/WeKnora` | `v0.8.0` | `1edcd54b43606d9079bb36650efe3f68707a79ea` |

## 获取完整源码

```bash
git clone --recurse-submodules https://github.com/superbuzzy/Weknora-Semantica.git
```

已有仓库：

```bash
git submodule update --init --recursive
```

## 上游升级原则

不要直接在 `upstream/*` 中长期开发 LeeClaw 功能。自研代码继续放在 `cobra-knowledge/integrations/`、Core、Adapter 和 Compatibility 目录。

升级某个上游时：

```text
拉取新 tag/commit
  ↓
更新对应 submodule pointer
  ↓
跑 compatibility / contract / E2E
  ↓
只在 Adapter / Plugin 边界吸收差异
  ↓
通过后更新 main
```

这样主仓既保留三套完整源码入口，又不会把约 5 万个上游文件复制进 LeeClaw 自身 Git 历史，后续与官方同步也只需要移动三个明确的版本指针。

Semantica 仍只作为设计研究参考，不作为 LeeClaw v0.5 运行时上游组件。
