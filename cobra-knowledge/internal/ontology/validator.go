package ontology

import (
	"fmt"
	"cobraknowledge.local/cobra-knowledge/internal/model"
)

type Violation struct { Code string `json:"code"`; Severity string `json:"severity"`; Subject string `json:"subject,omitempty"`; Message string `json:"message"` }
func ValidateOntology(o model.Ontology) []Violation {
	var out []Violation; classes:=map[string]model.OntologyClass{}; ids:=map[string]string{}
	for _,c:=range o.Classes { if c.ID==""||c.Label=="" { out=append(out,Violation{"ONTOLOGY_CLASS_REQUIRED","error",c.ID,"Class ID/Label 不能为空"}); continue }; if kind,ok:=ids[c.ID];ok{out=append(out,Violation{"ONTOLOGY_DUPLICATE_ID","error",c.ID,"ID 与 "+kind+" 重复"})}; ids[c.ID]="class"; classes[c.ID]=c }
	for _,c:=range o.Classes { for _,p:=range c.ParentIDs { if _,ok:=classes[p];!ok{out=append(out,Violation{"ONTOLOGY_PARENT_NOT_FOUND","error",c.ID,"父类不存在: "+p})} } }
	for _,p:=range o.Properties { if kind,ok:=ids[p.ID];ok{out=append(out,Violation{"ONTOLOGY_DUPLICATE_ID","error",p.ID,"ID 与 "+kind+" 重复"})}; ids[p.ID]="property"; for _,d:=range p.DomainIDs { if _,ok:=classes[d];!ok{out=append(out,Violation{"PROPERTY_DOMAIN_NOT_FOUND","error",p.ID,"属性 Domain 不存在: "+d})} } }
	for _,r:=range o.Relations { if kind,ok:=ids[r.ID];ok{out=append(out,Violation{"ONTOLOGY_DUPLICATE_ID","error",r.ID,"ID 与 "+kind+" 重复"})}; ids[r.ID]="relation"; if len(r.DomainIDs)==0||len(r.RangeIDs)==0{out=append(out,Violation{"RELATION_UNRESOLVED_DOMAIN_RANGE","warning",r.ID,"关系 Domain/Range 尚未稳定"})}; for _,d:=range r.DomainIDs { if _,ok:=classes[d];!ok{out=append(out,Violation{"RELATION_DOMAIN_NOT_FOUND","error",r.ID,"关系 Domain 不存在: "+d})} }; for _,d:=range r.RangeIDs { if _,ok:=classes[d];!ok{out=append(out,Violation{"RELATION_RANGE_NOT_FOUND","error",r.ID,"关系 Range 不存在: "+d})} } }
	return out
}
func ValidateGraph(graph model.GraphSnapshot,o model.Ontology) []Violation {
	classByLabel:=map[string]string{}; classSet:=map[string]struct{}{}; for _,c:=range o.Classes{classByLabel[c.Label]=c.ID;classSet[c.ID]=struct{}{}}
	relByLabel:=map[string]model.ObjectRelation{}; for _,r:=range o.Relations{relByLabel[r.Label]=r}; propByLabel:=map[string]model.DataProperty{}; for _,p:=range o.Properties{propByLabel[p.Label]=p}; entityType:=map[string]string{}; var out []Violation
	for _,e:=range graph.Entities { cid,ok:=classByLabel[e.BusinessType]; if !ok{out=append(out,Violation{"ENTITY_UNKNOWN_TYPE","error",e.ID,fmt.Sprintf("实体类型 %q 不在本体中",e.BusinessType)});continue}; entityType[e.ID]=cid; for key:=range e.Properties { p,ok:=propByLabel[key]; if !ok{out=append(out,Violation{"ENTITY_UNKNOWN_PROPERTY","warning",e.ID,fmt.Sprintf("属性 %q 不在本体中",key)});continue}; if len(p.DomainIDs)>0&&!contains(p.DomainIDs,cid){out=append(out,Violation{"PROPERTY_DOMAIN_VIOLATION","error",e.ID,fmt.Sprintf("属性 %q 不允许用于类型 %q",key,e.BusinessType)})} } }
	for _,r:=range graph.Relations { schema,ok:=relByLabel[r.BusinessType]; if !ok{out=append(out,Violation{"RELATION_UNKNOWN_TYPE","error",r.ID,fmt.Sprintf("关系 %q 不在本体中",r.BusinessType)});continue}; st,sok:=entityType[r.SourceID]; tt,tok:=entityType[r.TargetID]; if sok&&len(schema.DomainIDs)>0&&!contains(schema.DomainIDs,st){out=append(out,Violation{"RELATION_DOMAIN_VIOLATION","error",r.ID,"关系源实体类型不满足 Domain"})}; if tok&&len(schema.RangeIDs)>0&&!contains(schema.RangeIDs,tt){out=append(out,Violation{"RELATION_RANGE_VIOLATION","error",r.ID,"关系目标实体类型不满足 Range"})} }
	_=classSet; return out
}
func contains(in []string,target string) bool { for _,v:=range in{if v==target{return true}};return false }
