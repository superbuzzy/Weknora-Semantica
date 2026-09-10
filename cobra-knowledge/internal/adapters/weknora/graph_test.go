package weknora

import "testing"

func TestNormalizerPreservesBusinessTypePropertiesAndEvidence(t *testing.T) {
	raw := []byte(`{"domain":"distribution","knowledge_base_id":"kb1","knowledge_id":"k1","graph":{"node":[{"name":"金牛线","chunks":["c1","c2"],"attributes":["__type__=线路","电压等级=10kV","转供状态=不可转供"]},{"name":"一号台区","chunks":["c2"],"attributes":["__type__=台区"]}],"relation":[{"node1":"金牛线","node2":"一号台区","type":"供电"}]}}`)
	g, err := NewNormalizer().NormalizeJSON(raw, "fallback")
	if err != nil {
		t.Fatal(err)
	}
	if g.Domain != "distribution" || len(g.Entities) != 2 || len(g.Relations) != 1 {
		t.Fatalf("unexpected graph: %+v", g)
	}
	var lineFound bool
	for _, e := range g.Entities {
		if e.Name == "金牛线" {
			lineFound = true
			if e.BusinessType != "线路" {
				t.Fatalf("type=%s", e.BusinessType)
			}
			if e.Properties["电压等级"] != "10kV" {
				t.Fatalf("properties=%v", e.Properties)
			}
			if len(e.Evidence) != 2 {
				t.Fatalf("evidence=%d", len(e.Evidence))
			}
		}
	}
	if !lineFound {
		t.Fatal("line entity missing")
	}
	if g.Relations[0].Quality != "derived" || len(g.Relations[0].Evidence) != 1 || g.Relations[0].Evidence[0].ChunkID != "c2" {
		t.Fatalf("relation evidence=%+v", g.Relations[0])
	}
}

func TestNormalizerMarksMissingType(t *testing.T) {
	raw := []byte(`{"node":[{"name":"未知对象","chunks":["c1"]}]}`)
	g, err := NewNormalizer().NormalizeJSON(raw, "d")
	if err != nil {
		t.Fatal(err)
	}
	if g.Entities[0].BusinessType != "未分类实体" || len(g.Warnings) == 0 {
		t.Fatalf("unexpected: %+v", g)
	}
}
