# CobraKnowledge v0.2

CobraKnowledge v0.2 is a Go-native enterprise knowledge retrieval and ontology control plane. It treats WeKnora as a replaceable knowledge source and Semantica as a design reference only. Neither upstream project is modified or required at runtime.

## Core idea

```text
Wiki Graph        Entity Graph        Ontology Graph        Business Data
    \                 |                  /                     /
     \                |                 /                     /
              Retrieval Control Plane
       Semantic Resolver -> Planner -> Retrievers
                         -> Arbiter -> Context Assembler
                                  |
                              Context Pack
                                  |
                          Agent Runtime / Skill
```

The three graphs are knowledge assets. The core product is the retrieval strategy: what to search, where to search, how much evidence is enough, which source wins conflicts, and when retrieval should stop.

## Upstream isolation

WeKnora and Semantica are upstream references only. CobraKnowledge does not modify or import either project at runtime. Their evolution is absorbed through Adapter contracts.

See `ARCHITECTURE.md`, `docs/MODULES.md`, and `RELEASE-v0.2.md` for the current design and implementation status.
