package retrieval

import (
	"sort"
	"strings"
	"cobraknowledge.local/cobra-knowledge/internal/model"
)
type CatalogTerm struct{ID string `json:"id"`;Label string `json:"label"`;Aliases []string `json:"aliases,omitempty"`;Kind string `json:"kind"`;PreferredSources []model.RetrievalSource `json:"preferred_sources,omitempty"`}
type SemanticCatalog struct{Domain string `json:"domain"`;Classes []CatalogTerm `json:"classes"`;Properties []CatalogTerm `json:"properties"`;Relations []CatalogTerm `json:"relations"`}
type TermOverlay struct{Aliases []string `json:"aliases,omitempty"`;PreferredSources []string `json:"preferred_sources,omitempty"`}
type CatalogOverlay struct{Classes map[string]TermOverlay `json:"classes,omitempty"`;Properties map[string]TermOverlay `json:"properties,omitempty"`;Relations map[string]TermOverlay `json:"relations,omitempty"`}
func CompileCatalog(o model.Ontology)SemanticCatalog{return CompileCatalogWithOverlay(o,CatalogOverlay{})}
func CompileCatalogWithOverlay(o model.Ontology,overlay CatalogOverlay)SemanticCatalog{c:=SemanticCatalog{Domain:o.Domain};for _,cls:=range o.Classes{ov:=overlay.Classes[cls.Label];c.Classes=append(c.Classes,CatalogTerm{ID:cls.ID,Label:cls.Label,Aliases:mergeStrings(cls.Aliases,ov.Aliases),Kind:"class",PreferredSources:overlaySources(ov.PreferredSources,[]model.RetrievalSource{model.SourceEntityGraph})})};for _,prop:=range o.Properties{sources:=[]model.RetrievalSource{model.SourceEntityGraph};if prop.RetrievalPolicy!=nil&&len(prop.RetrievalPolicy.PreferredSources)>0{sources=nil;for _,s:=range prop.RetrievalPolicy.PreferredSources{sources=append(sources,model.RetrievalSource(s))}};ov:=overlay.Properties[prop.Label];c.Properties=append(c.Properties,CatalogTerm{ID:prop.ID,Label:prop.Label,Aliases:mergeStrings(prop.Aliases,ov.Aliases),Kind:"property",PreferredSources:overlaySources(ov.PreferredSources,sources)})};for _,rel:=range o.Relations{ov:=overlay.Relations[rel.Label];c.Relations=append(c.Relations,CatalogTerm{ID:rel.ID,Label:rel.Label,Aliases:mergeStrings(rel.Aliases,ov.Aliases),Kind:"relation",PreferredSources:overlaySources(ov.PreferredSources,[]model.RetrievalSource{model.SourceEntityGraph})})};sortTerms(c.Classes);sortTerms(c.Properties);sortTerms(c.Relations);return c}
func sortTerms(items []CatalogTerm){sort.Slice(items,func(i,j int)bool{return items[i].Label<items[j].Label})}
func termMatches(query string,t CatalogTerm)bool{q:=strings.ToLower(query);if t.Label!=""&&strings.Contains(q,strings.ToLower(t.Label)){return true};for _,a:=range t.Aliases{if a!=""&&strings.Contains(q,strings.ToLower(a)){return true}};return false}
func overlaySources(in []string,fallback []model.RetrievalSource)[]model.RetrievalSource{if len(in)==0{return fallback};out:=make([]model.RetrievalSource,0,len(in));for _,s:=range in{if strings.TrimSpace(s)!=""{out=append(out,model.RetrievalSource(strings.TrimSpace(s)))}};return out}
func mergeStrings(a,b []string)[]string{seen:=map[string]struct{}{};var out []string;for _,s:=range append(append([]string(nil),a...),b...){s=strings.TrimSpace(s);if s==""{continue};if _,ok:=seen[s];ok{continue};seen[s]=struct{}{};out=append(out,s)};sort.Strings(out);return out}
