package retrieval

import (
	"testing"
	"time"
	"cobraknowledge.local/cobra-knowledge/internal/model"
)
func tm(s string)*time.Time{t,_:=time.Parse(time.RFC3339,s);return &t}
func TestArbiterAuthorityWinsConflict(t *testing.T){a:=NewArbiter(ArbitrationPolicy{DefaultSourcePriority:map[string]float64{"GIS":1,"PMS":0.7},DefaultDelta:0.05});in:=[]model.Assertion{{ID:"a",SlotID:"s",PredicateID:"p",Value:"不可转供",Source:"GIS",Confidence:.9},{ID:"b",SlotID:"s",PredicateID:"p",Value:"可转供",Source:"PMS",Confidence:.95}};r:=a.Arbitrate(in,time.Now().UTC());if r[0].Status!="accepted_with_conflict"||r[0].Accepted.Source!="GIS"{t.Fatalf("result=%+v",r)}}
func TestArbiterUnresolvedWhenNoPolicyAdvantage(t *testing.T){a:=NewArbiter(ArbitrationPolicy{DefaultDelta:.05});in:=[]model.Assertion{{ID:"a",SlotID:"s",PredicateID:"p",Value:"A",Source:"x",Authority:.8,Confidence:.8},{ID:"b",SlotID:"s",PredicateID:"p",Value:"B",Source:"y",Authority:.8,Confidence:.8}};r:=a.Arbitrate(in,time.Now().UTC());if r[0].Status!="unresolved_conflict"{t.Fatalf("result=%+v",r)}}
func TestArbiterHistoricalValidTime(t *testing.T){a:=NewArbiter(ArbitrationPolicy{DefaultDelta:.05});in:=[]model.Assertion{{ID:"old",SlotID:"s",PredicateID:"p",Value:"旧",Source:"doc",Authority:.7,Confidence:.9,ValidFrom:tm("2024-01-01T00:00:00Z"),ValidTo:tm("2024-12-31T23:59:59Z")},{ID:"new",SlotID:"s",PredicateID:"p",Value:"新",Source:"doc",Authority:.7,Confidence:.9,ValidFrom:tm("2025-01-01T00:00:00Z")}};r:=a.Arbitrate(in,*tm("2024-06-01T00:00:00Z"));if r[0].Accepted==nil||r[0].Accepted.ID!="old"{t.Fatalf("result=%+v",r)}}
func TestArbiterFreshness(t *testing.T){a:=NewArbiter(ArbitrationPolicy{Predicates:map[string]PredicatePolicy{"p":{MaxAgeSeconds:3600}},DefaultDelta:.05});in:=[]model.Assertion{{ID:"a",SlotID:"s",PredicateID:"p",Value:1,Source:"api",Authority:1,Confidence:1,ObservedAt:tm("2026-09-10T00:00:00Z")}};r:=a.Arbitrate(in,*tm("2026-09-10T03:00:00Z"));if r[0].Status!="no_valid_assertion"||len(r[0].Stale)!=1{t.Fatalf("result=%+v",r)}}
