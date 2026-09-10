package ontology

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

type FileRegistry struct{ Root string }
func NewFileRegistry(root string) *FileRegistry { return &FileRegistry{Root: root} }
func (r *FileRegistry) SaveCandidate(o model.Ontology) (string,error) { if o.Status != model.StatusCandidate { return "", fmt.Errorf("only candidate ontology can be saved as candidate") }; return r.write(o,false) }
func (r *FileRegistry) Approve(o model.Ontology, version, reviewer string) (model.Ontology,string,error) {
	for _, item := range o.ReviewQueue { if item.Risk == "high" && item.Status != "approved" { return model.Ontology{},"",fmt.Errorf("high-risk review item %s is not approved", item.ID) } }
	if strings.TrimSpace(version)=="" { return model.Ontology{},"",fmt.Errorf("approved version is required") }
	approved:=o; approved.Version=version; approved.Status=model.StatusApproved; approved.CreatedAt=time.Now().UTC(); if approved.Metadata==nil { approved.Metadata=map[string]string{} }; approved.Metadata["reviewer"]=reviewer
	path,err:=r.write(approved,true); return approved,path,err
}
func (r *FileRegistry) write(o model.Ontology, immutable bool) (string,error) {
	domain:=sanitize(o.Domain); version:=sanitize(o.Version); dir:=filepath.Join(r.Root,domain); if err:=os.MkdirAll(dir,0o755); err!=nil{return "",err}; path:=filepath.Join(dir,version+".json")
	if immutable { if _,err:=os.Stat(path); err==nil { return "",fmt.Errorf("immutable ontology version already exists: %s",path) } }
	data,err:=json.MarshalIndent(o,"","  "); if err!=nil{return "",err}; if err:=os.WriteFile(path,data,0o644); err!=nil{return "",err}; return path,nil
}
func sanitize(s string) string { s=strings.TrimSpace(s); replacer:=strings.NewReplacer("/","_","\\","_"," ","_"); if s=="" { return "default" }; return replacer.Replace(s) }
