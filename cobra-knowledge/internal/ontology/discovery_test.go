package ontology

import (
	"context"
	"strings"
	"testing"
	"cobraknowledge.local/cobra-knowledge/internal/model"
)
func sampleGraph() model.GraphSnapshot { lineID:=model.StableID("ent","d","line");areaID:=model.StableID("ent","d","area");subID:=model.StableID("ent","d","sub");return model.GraphSnapshot{Domain:"d",Entities:[]model.Entity{{ID:lineID,Name:"金牛线",BusinessType:"线路",TypeID:model.StableID("cls","d","线路"),Properties:map[string]interface{}{"电压等级":"10kV","转供状态":"不可转供"}},{ID:areaID,Name:"一号台区",BusinessType:"台区",TypeID:model.StableID("cls","d","台区")},{ID:subID,Name:"金牛变",BusinessType:"变电站",TypeID:model.StableID("cls","d","变电站")}},Relations:[]model.Relation{{SourceID:lineID,TargetID:areaID,BusinessType:"供电"},{SourceID:lineID,TargetID:subID,BusinessType:"所属"}}}}
func TestDiscoverCandidateOntology(t *testing.T){o,report,err:=NewDiscoveryService(nil).Discover(context.Background(),sampleGraph());if err!=nil{t.Fatal(err)};if len(o.Classes)!=3||len(o.Properties)!=2||len(o.Relations)!=2{t.Fatalf("ontology counts: %+v",o)};if len(report.Classes)!=3{t.Fatalf("report=%+v",report)};for _,r:=range o.Relations{if len(r.DomainIDs)!=1||len(r.RangeIDs)!=1{t.Fatalf("relation domain/range missing: %+v",r)}}}
func TestCompileWeKnoraUsesChineseLabels(t *testing.T){o,_,_:=NewDiscoveryService(nil).Discover(context.Background(),sampleGraph());cfg:=CompileWeKnora(o);if !strings.Contains(cfg.CustomInstructions,"线路")||!strings.Contains(cfg.CustomInstructions,"供电：线路 -> 台区"){t.Fatalf("instructions=%s",cfg.CustomInstructions)};if len(cfg.Nodes)!=3||len(cfg.Relations)!=2{t.Fatalf("cfg=%+v",cfg)}}
func TestValidateGraphRejectsWrongDirection(t *testing.T){o,_,_:=NewDiscoveryService(nil).Discover(context.Background(),sampleGraph());g:=sampleGraph();g.Relations[0].SourceID=g.Entities[1].ID;g.Relations[0].TargetID=g.Entities[0].ID;v:=ValidateGraph(g,o);var found bool;for _,x:=range v{if x.Code=="RELATION_DOMAIN_VIOLATION"||x.Code=="RELATION_RANGE_VIOLATION"{found=true}};if !found{t.Fatalf("expected domain/range violation, got %+v",v)}}
