# Upstream references

CobraKnowledge keeps WeKnora and Semantica as upstream references rather than vendoring or modifying their source code.

- WeKnora: https://github.com/Tencent/WeKnora
- Semantica: https://github.com/semantica-agi/semantica

The runtime under `cobra-knowledge/` does not import Semantica and does not require either upstream repository to be checked out. Use `cobra-knowledge/scripts/sync-upstreams.sh` in a workspace layout where sibling directories `upstream/weknora` and `upstream/semantica` are real git clones when you need local source inspection.
