package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	wk "cobraknowledge.local/cobra-knowledge/internal/adapters/weknora"
	ctxsvc "cobraknowledge.local/cobra-knowledge/internal/context"
	"cobraknowledge.local/cobra-knowledge/internal/model"
	ontsvc "cobraknowledge.local/cobra-knowledge/internal/ontology"
	retrievalsvc "cobraknowledge.local/cobra-knowledge/internal/retrieval"
	"cobraknowledge.local/cobra-knowledge/internal/store"
)

type Server struct {
	ContextService *ctxsvc.Service
	WeKnora        *wk.SearchClient
}

type request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}
type response struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id,omitempty"`
	Result  interface{} `json:"result,omitempty"`
	Error   *rpcError   `json:"error,omitempty"`
}
type rpcError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func NewFromEnv() (*Server, error) {
	var onto model.Ontology
	ontologyLoaded := false
	registryRoot := strings.TrimSpace(os.Getenv("COBRA_ONTOLOGY_REGISTRY_ROOT"))
	registryKBID := strings.TrimSpace(os.Getenv("COBRA_ONTOLOGY_KB_ID"))
	if registryRoot != "" && registryKBID != "" {
		resolution, err := ontsvc.NewFSRegistry(registryRoot).ResolveForKnowledgeBase(context.Background(), registryKBID)
		if err != nil {
			return nil, fmt.Errorf("resolve ontology from registry for %s: %w", registryKBID, err)
		}
		onto = resolution.Ontology
		ontologyLoaded = true
	}
	ontoPath := strings.TrimSpace(os.Getenv("COBRA_ONTOLOGY_FILE"))
	if !ontologyLoaded && ontoPath != "" {
		if err := store.ReadJSON(ontoPath, &onto); err != nil {
			return nil, err
		}
		ontologyLoaded = true
	}
	overlay := retrievalsvc.CatalogOverlay{}
	if overlayPath := os.Getenv("COBRA_CATALOG_OVERLAY_FILE"); overlayPath != "" {
		if err := store.ReadJSON(overlayPath, &overlay); err != nil {
			return nil, err
		}
	}
	planner := retrievalsvc.NewPlanner(retrievalsvc.CompileCatalogWithOverlay(onto, overlay))
	arbiterPolicy := retrievalsvc.ArbitrationPolicyFromOntology(onto, map[string]float64{"business_data": 1.0, "entity_graph": 0.8, "weknora": 0.6})
	if policyPath := os.Getenv("COBRA_ARBITRATION_POLICY_FILE"); policyPath != "" {
		if err := store.ReadJSON(policyPath, &arbiterPolicy); err != nil {
			return nil, err
		}
	}
	arbiter := retrievalsvc.NewArbiter(arbiterPolicy)
	service := ctxsvc.NewService(planner, arbiter, ctxsvc.NewAssembler())
	if ontologyLoaded {
		service.Register(&ctxsvc.StaticOntologyRetriever{Ontology: onto})
	}
	graphPath := os.Getenv("COBRA_ENTITY_GRAPH_FILE")
	if graphPath != "" {
		var g model.GraphSnapshot
		if err := store.ReadJSON(graphPath, &g); err != nil {
			return nil, err
		}
		service.Register(ctxsvc.NewStaticEntityGraphRetriever(g))
	}
	var wkClient *wk.SearchClient
	if base := os.Getenv("WEKNORA_BASE_URL"); base != "" {
		wkClient = wk.NewSearchClient(base, os.Getenv("WEKNORA_API_KEY"))
		service.Register(&ctxsvc.WeKnoraRAGRetriever{Client: wkClient, KnowledgeBaseIDs: splitEnvList(os.Getenv("WEKNORA_KB_IDS")), KnowledgeIDs: splitEnvList(os.Getenv("WEKNORA_KNOWLEDGE_IDS"))})
	}
	return &Server{ContextService: service, WeKnora: wkClient}, nil
}

func (s *Server) Serve(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 4096), 4*1024*1024)
	enc := json.NewEncoder(out)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(strings.TrimSpace(string(line))) == 0 {
			continue
		}
		var req request
		if err := json.Unmarshal(line, &req); err != nil {
			_ = enc.Encode(response{JSONRPC: "2.0", Error: &rpcError{Code: -32700, Message: "parse error"}})
			continue
		}
		resp := s.handle(ctx, req)
		if req.ID != nil {
			if err := enc.Encode(resp); err != nil {
				return err
			}
		}
	}
	return scanner.Err()
}

func (s *Server) handle(ctx context.Context, req request) response {
	resp := response{JSONRPC: "2.0", ID: req.ID}
	switch req.Method {
	case "initialize":
		resp.Result = map[string]interface{}{"protocolVersion": "2025-06-18", "capabilities": map[string]interface{}{"tools": map[string]interface{}{}}, "serverInfo": map[string]string{"name": "cobra-knowledge", "version": "0.5.0"}}
	case "notifications/initialized":
		resp.Result = map[string]interface{}{}
	case "tools/list":
		resp.Result = map[string]interface{}{"tools": toolDefinitions()}
	case "tools/call":
		var p struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &p); err != nil {
			resp.Error = &rpcError{Code: -32602, Message: err.Error()}
			break
		}
		result, err := s.callTool(ctx, p.Name, p.Arguments)
		if err != nil {
			resp.Result = toolError(err)
		} else {
			resp.Result = toolResult(result)
		}
	default:
		resp.Error = &rpcError{Code: -32601, Message: "method not found"}
	}
	return resp
}

func (s *Server) callTool(ctx context.Context, name string, args json.RawMessage) (interface{}, error) {
	switch name {
	case "context.retrieve":
		var req model.QueryRequest
		if err := json.Unmarshal(args, &req); err != nil {
			return nil, err
		}
		return s.ContextService.Retrieve(ctx, req)
	case "context.get_evidence":
		if s.WeKnora == nil {
			return nil, fmt.Errorf("WEKNORA_BASE_URL is not configured")
		}
		var p struct {
			ChunkID string `json:"chunk_id"`
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return nil, err
		}
		if p.ChunkID == "" {
			return nil, fmt.Errorf("chunk_id is required")
		}
		return s.WeKnora.GetChunk(ctx, p.ChunkID)
	case "retrieval.plan":
		var p struct {
			Ontology model.Ontology     `json:"ontology"`
			Request  model.QueryRequest `json:"request"`
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return nil, err
		}
		return retrievalsvc.NewPlanner(retrievalsvc.CompileCatalog(p.Ontology)).Plan(p.Request), nil
	case "ontology.discover":
		var p struct {
			Graph model.GraphSnapshot `json:"graph"`
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return nil, err
		}
		o, report, err := ontsvc.NewDiscoveryService(nil).Discover(ctx, p.Graph)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{"ontology": o, "pattern_report": report}, nil
	case "ontology.compile_weknora":
		var p struct {
			Ontology model.Ontology `json:"ontology"`
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return nil, err
		}
		return ontsvc.CompileWeKnora(p.Ontology), nil
	case "knowledge.arbitrate":
		var p struct {
			Assertions []model.Assertion               `json:"assertions"`
			Policy     retrievalsvc.ArbitrationPolicy `json:"policy"`
			At         *time.Time                      `json:"at,omitempty"`
		}
		if err := json.Unmarshal(args, &p); err != nil {
			return nil, err
		}
		qt := time.Now().UTC()
		if p.At != nil {
			qt = p.At.UTC()
		}
		return retrievalsvc.NewArbiter(p.Policy).Arbitrate(p.Assertions, qt), nil
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func toolDefinitions() []map[string]interface{} {
	return []map[string]interface{}{
		{"name": "context.retrieve", "description": "统一检索入口。根据本体语义和检索策略规划 Wiki/RAG、实体图、本体图或业务数据检索，返回 Context Pack。", "inputSchema": objSchema(map[string]interface{}{"query": map[string]string{"type": "string"}, "domain": map[string]string{"type": "string"}, "task": map[string]string{"type": "string"}, "scope": map[string]string{"type": "object"}}, []string{"query"})},
		{"name": "context.get_evidence", "description": "根据 chunk_id 读取 WeKnora 原始 Chunk 证据。", "inputSchema": objSchema(map[string]interface{}{"chunk_id": map[string]string{"type": "string"}}, []string{"chunk_id"})},
		{"name": "retrieval.plan", "description": "开发/审计工具：根据 Ontology 和问题生成可审计的 RetrievalPlan。", "inputSchema": objSchema(map[string]interface{}{"ontology": map[string]string{"type": "object"}, "request": map[string]string{"type": "object"}}, []string{"ontology", "request"})},
		{"name": "ontology.discover", "description": "从规范化实体图进行 bottom-up 模式归纳，生成候选本体与模式报告。", "inputSchema": objSchema(map[string]interface{}{"graph": map[string]string{"type": "object"}}, []string{"graph"})},
		{"name": "ontology.compile_weknora", "description": "把 CobraKnowledge Ontology 编译为 WeKnora ExtractConfig 与中文抽取约束。", "inputSchema": objSchema(map[string]interface{}{"ontology": map[string]string{"type": "object"}}, []string{"ontology"})},
		{"name": "knowledge.arbitrate", "description": "按有效时间、来源策略、置信度和观测时间裁决 Assertion 冲突。", "inputSchema": objSchema(map[string]interface{}{"assertions": map[string]string{"type": "array"}, "policy": map[string]string{"type": "object"}}, []string{"assertions"})},
	}
}
func objSchema(props interface{}, required []string) map[string]interface{} {
	return map[string]interface{}{"type": "object", "properties": props, "required": required}
}
func toolResult(v interface{}) map[string]interface{} {
	raw, _ := json.MarshalIndent(v, "", "  ")
	return map[string]interface{}{"content": []map[string]string{{"type": "text", "text": string(raw)}}, "structuredContent": v}
}
func toolError(err error) map[string]interface{} {
	return map[string]interface{}{"isError": true, "content": []map[string]string{{"type": "text", "text": err.Error()}}}
}
func splitEnvList(v string) []string {
	var out []string
	for _, x := range strings.Split(v, ",") {
		x = strings.TrimSpace(x)
		if x != "" {
			out = append(out, x)
		}
	}
	return out
}
