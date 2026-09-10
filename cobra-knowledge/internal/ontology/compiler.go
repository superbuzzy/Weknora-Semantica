package ontology

import (
	"fmt"
	"sort"
	"strings"

	"cobraknowledge.local/cobra-knowledge/internal/adapters/weknora"
	"cobraknowledge.local/cobra-knowledge/internal/model"
)

type WeKnoraExtractConfig struct {
	Enabled            bool                    `json:"enabled"`
	Text               string                  `json:"text,omitempty"`
	Tags               []string                `json:"tags,omitempty"`
	Nodes              []weknora.GraphNode     `json:"nodes,omitempty"`
	Relations          []weknora.GraphRelation `json:"relations,omitempty"`
	CustomInstructions string                  `json:"custom_instructions,omitempty"`
}

func CompileWeKnora(o model.Ontology) WeKnoraExtractConfig {
	classLabel := map[string]string{}
	var classLabels []string
	for _, c := range o.Classes { classLabel[c.ID] = c.Label; classLabels = append(classLabels, c.Label) }
	sort.Strings(classLabels)
	var nodes []weknora.GraphNode
	for _, c := range o.Classes {
		attrs := []string{"__type__=" + c.Label}
		for _, p := range o.Properties { if contains(p.DomainIDs, c.ID) { attrs = append(attrs, p.Label+"=<按原文抽取>") } }
		sort.Strings(attrs[1:])
		nodes = append(nodes, weknora.GraphNode{Name: "<" + c.Label + "实体>", Attributes: attrs})
	}
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].Name < nodes[j].Name })
	var rels []weknora.GraphRelation; var relLines []string
	for _, r := range o.Relations {
		if len(r.DomainIDs) == 0 || len(r.RangeIDs) == 0 { continue }
		dl := classLabel[r.DomainIDs[0]]; rl := classLabel[r.RangeIDs[0]]
		if dl == "" || rl == "" { continue }
		rels = append(rels, weknora.GraphRelation{Node1: "<"+dl+"实体>", Node2: "<"+rl+"实体>", Type: r.Label})
		relLines = append(relLines, fmt.Sprintf("- %s：%s -> %s", r.Label, dl, rl))
	}
	sort.Slice(rels, func(i, j int) bool { if rels[i].Type == rels[j].Type { return rels[i].Node1 < rels[j].Node1 }; return rels[i].Type < rels[j].Type })
	sort.Strings(relLines)
	var propLines []string
	for _, p := range o.Properties {
		var domains []string
		for _, id := range p.DomainIDs { if label := classLabel[id]; label != "" { domains = append(domains, label) } }
		if len(domains) > 0 { propLines = append(propLines, fmt.Sprintf("- %s.%s", strings.Join(domains, "/"), p.Label)) }
	}
	sort.Strings(propLines)
	instructions := strings.Join([]string{"按已审核/候选本体约束抽取实体关系。","允许实体类型："+strings.Join(classLabels,"、"),"每个实体必须在 attributes 中输出 __type__=<实体类型>，不得创造未定义类型。","允许关系：\n"+strings.Join(relLines,"\n"),"已知属性：\n"+strings.Join(propLines,"\n"),"若原文出现无法映射的新类型、新关系或新属性，不要强行归类；保留在候选发现流程中处理。","不得依据常识补充原文不存在的事实，实体、属性和关系均应能回溯到 Chunk 证据。"}, "\n\n")
	return WeKnoraExtractConfig{Enabled:true, Text:"CobraKnowledge ontology constrained graph extraction", Tags:classLabels, Nodes:nodes, Relations:rels, CustomInstructions:instructions}
}
