package graph

import (
	"testing"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

func TestResolverMergesSameTypeNormalizedName(t *testing.T) {
	typ := model.StableID("cls", "d", "线路")
	in := []model.Entity{{ID: "a", Name: "10kV 金牛线", BusinessType: "线路", TypeID: typ, Aliases: []string{"金牛线"}, Properties: map[string]interface{}{"电压等级": "10kV"}}, {ID: "b", Name: "金牛线", BusinessType: "线路", TypeID: typ, Aliases: []string{"10kV 金牛线"}, Properties: map[string]interface{}{"电压等级": "10kV"}}}
	r := NewEntityResolver().Resolve(in)
	if len(r.Entities) != 1 {
		t.Fatalf("entities=%d %+v", len(r.Entities), r)
	}
	if r.IDMap["b"] != "a" {
		t.Fatalf("idmap=%v", r.IDMap)
	}
}

func TestResolverDoesNotMergeDifferentTypes(t *testing.T) {
	in := []model.Entity{{ID: "a", Name: "金牛", BusinessType: "线路", TypeID: "line"}, {ID: "b", Name: "金牛", BusinessType: "变电站", TypeID: "sub"}}
	r := NewEntityResolver().Resolve(in)
	if len(r.Entities) != 2 {
		t.Fatalf("unexpected merge: %+v", r)
	}
}

func TestResolverPreservesPropertyConflict(t *testing.T) {
	typ := "line"
	in := []model.Entity{{ID: "a", Name: "金牛线", BusinessType: "线路", TypeID: typ, Properties: map[string]interface{}{"转供状态": "不可转供"}}, {ID: "b", Name: "金牛线", BusinessType: "线路", TypeID: typ, Properties: map[string]interface{}{"转供状态": "可转供"}}}
	r := NewEntityResolver().Resolve(in)
	if len(r.Conflicts) != 1 {
		t.Fatalf("expected conflict: %+v", r)
	}
	if r.Entities[0].Properties["转供状态"] != "不可转供" {
		t.Fatal("existing value should not be overwritten")
	}
}

func TestAssertionBuilderSeparatesFactsFromEntity(t *testing.T) {
	g := model.GraphSnapshot{Domain: "d", Entities: []model.Entity{{ID: "e1", Name: "金牛线", Properties: map[string]interface{}{"转供状态": "不可转供"}}}}
	a := NewAssertionBuilder("entity_graph").FromGraph(g)
	if len(a) != 1 || a[0].SubjectID != "e1" || a[0].Value != "不可转供" || a[0].SlotID == "" {
		t.Fatalf("assertions=%+v", a)
	}
}
