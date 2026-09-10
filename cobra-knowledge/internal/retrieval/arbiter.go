package retrieval

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"cobraknowledge.local/cobra-knowledge/internal/model"
)

type PredicatePolicy struct {
	SourcePriority       map[string]float64 `json:"source_priority,omitempty"`
	MaxAgeSeconds        int64              `json:"max_age_seconds,omitempty"`
	UnresolvedScoreDelta float64            `json:"unresolved_score_delta,omitempty"`
	PreferNewer          bool               `json:"prefer_newer,omitempty"`
}

type ArbitrationPolicy struct {
	DefaultSourcePriority map[string]float64         `json:"default_source_priority,omitempty"`
	Predicates            map[string]PredicatePolicy `json:"predicates,omitempty"`
	DefaultDelta          float64                    `json:"default_delta"`
}

type Arbiter struct{ Policy ArbitrationPolicy }

func NewArbiter(policy ArbitrationPolicy) *Arbiter {
	if policy.DefaultDelta == 0 {
		policy.DefaultDelta = 0.05
	}
	return &Arbiter{Policy: policy}
}

func (a *Arbiter) Arbitrate(assertions []model.Assertion, queryTime time.Time) []model.ArbitrationResult {
	if queryTime.IsZero() {
		queryTime = time.Now().UTC()
	}
	bySlot := map[string][]model.Assertion{}
	for _, x := range assertions {
		slot := x.SlotID
		if slot == "" {
			slot = x.SubjectID + "\x00" + x.PredicateID
		}
		bySlot[slot] = append(bySlot[slot], x)
	}
	slots := make([]string, 0, len(bySlot))
	for s := range bySlot { slots = append(slots, s) }
	sort.Strings(slots)
	out := make([]model.ArbitrationResult, 0, len(slots))
	for _, slot := range slots { out = append(out, a.arbitrateSlot(slot, bySlot[slot], queryTime)) }
	return out
}

func (a *Arbiter) arbitrateSlot(slot string, candidates []model.Assertion, qt time.Time) model.ArbitrationResult {
	var valid, stale []model.Assertion
	for _, x := range candidates {
		if !scopeCompatible(x.Scope, candidates[0].Scope) { continue }
		if !validAt(x, qt) || isStale(x, qt, a.predicatePolicy(x.PredicateID)) { stale = append(stale, x); continue }
		valid = append(valid, x)
	}
	if len(valid) == 0 { return model.ArbitrationResult{SlotID:slot, Status:"no_valid_assertion", Stale:stale, Reason:"没有满足有效时间/新鲜度策略的事实"} }
	sort.SliceStable(valid, func(i,j int)bool{return a.score(valid[i],qt)>a.score(valid[j],qt)})
	top:=valid[0]; var conflicts []model.Assertion
	for _,x:=range valid[1:]{if !equivalentValue(top.Value,x.Value)&&!isSupersededBy(x,top){conflicts=append(conflicts,x)}}
	if len(conflicts)>0{runner:=conflicts[0];delta:=a.score(top,qt)-a.score(runner,qt);policy:=a.predicatePolicy(top.PredicateID);threshold:=policy.UnresolvedScoreDelta;if threshold==0{threshold=a.Policy.DefaultDelta};if delta<=threshold{return model.ArbitrationResult{SlotID:slot,Status:"unresolved_conflict",Conflicts:append([]model.Assertion{top},conflicts...),Stale:stale,Reason:"同一事实槽位存在无决定性策略优势的冲突"}}}
	return model.ArbitrationResult{SlotID:slot,Status:map[bool]string{true:"accepted_with_conflict",false:"accepted"}[len(conflicts)>0],Accepted:&top,Conflicts:conflicts,Stale:stale,Reason:"按有效时间、来源优先级、置信度和观测时间确定"}
}
func (a *Arbiter) predicatePolicy(predicate string)PredicatePolicy{if p,ok:=a.Policy.Predicates[predicate];ok{return p};return PredicatePolicy{}}
func (a *Arbiter) score(x model.Assertion,qt time.Time)float64{p:=a.predicatePolicy(x.PredicateID);priority:=x.Authority;if v,ok:=p.SourcePriority[x.Source];ok{priority=v}else if v,ok:=a.Policy.DefaultSourcePriority[x.Source];ok{priority=v};score:=0.65*priority+0.35*x.Confidence;if p.PreferNewer&&x.ObservedAt!=nil{age:=qt.Sub(x.ObservedAt.UTC()).Hours();if age<0{age=0};score+=0.05/(1+age/24)};return score}
func validAt(x model.Assertion,t time.Time)bool{if x.ValidFrom!=nil&&t.Before(x.ValidFrom.UTC()){return false};if x.ValidTo!=nil&&t.After(x.ValidTo.UTC()){return false};return true}
func isStale(x model.Assertion,t time.Time,p PredicatePolicy)bool{if p.MaxAgeSeconds<=0||x.ObservedAt==nil{return false};return t.Sub(x.ObservedAt.UTC())>time.Duration(p.MaxAgeSeconds)*time.Second}
func equivalentValue(a,b interface{})bool{return normalizeValue(a)==normalizeValue(b)}
func normalizeValue(v interface{})string{return strings.ToLower(strings.Join(strings.Fields(fmt.Sprint(v)),""))}
func isSupersededBy(old,new model.Assertion)bool{for _,id:=range new.Supersedes{if id==old.ID{return true}};return false}
func scopeCompatible(a,b map[string]string)bool{if len(a)==0||len(b)==0{return true};for k,av:=range a{if bv,ok:=b[k];ok&&av!=bv{return false}};return true}
