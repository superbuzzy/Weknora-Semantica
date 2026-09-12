package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"cobraknowledge.local/cobra-knowledge/internal/access"
	businessadapter "cobraknowledge.local/cobra-knowledge/internal/adapters/business"
	"cobraknowledge.local/cobra-knowledge/internal/graphview"
	"cobraknowledge.local/cobra-knowledge/internal/httpapi"
	ontsvc "cobraknowledge.local/cobra-knowledge/internal/ontology"
	"cobraknowledge.local/cobra-knowledge/internal/promotion"
	"cobraknowledge.local/cobra-knowledge/internal/promotionapi"
	"cobraknowledge.local/cobra-knowledge/internal/retrieval"
	"cobraknowledge.local/cobra-knowledge/internal/runtimecontext"
	"cobraknowledge.local/cobra-knowledge/internal/store"
)

func main() {
	listen := flag.String("listen", envOr("COBRA_GRAPH_API_LISTEN", ":8090"), "listen address")
	registryRoot := flag.String("registry-root", envOr("COBRA_ONTOLOGY_REGISTRY_ROOT", "var/ontology-registry"), "ontology registry root")
	promotionRoot := flag.String("promotion-root", envOr("LEECLAW_PROMOTION_ROOT", "var/promotion"), "promotion candidate store root")
	neo4jURL := flag.String("neo4j-url", envOr("COBRA_NEO4J_URL", "http://localhost:7474"), "Neo4j HTTP base URL")
	neo4jDB := flag.String("neo4j-database", envOr("COBRA_NEO4J_DATABASE", "neo4j"), "Neo4j database")
	neo4jUser := flag.String("neo4j-user", envOr("COBRA_NEO4J_USER", "neo4j"), "Neo4j username")
	neo4jPassword := flag.String("neo4j-password", os.Getenv("COBRA_NEO4J_PASSWORD"), "Neo4j password")
	weknoraURL := flag.String("weknora-url", os.Getenv("COBRA_WEKNORA_BASE_URL"), "WeKnora base URL for RBAC and Promotion")
	openvikingURL := flag.String("openviking-url", os.Getenv("OPENVIKING_BASE_URL"), "OpenViking base URL for Skill Promotion")
	openvikingAPIKey := flag.String("openviking-api-key", envOr("OPENVIKING_API_KEY", os.Getenv("OPENVIKING_ROOT_API_KEY")), "OpenViking service API key for Skill Promotion")
	businessURL := flag.String("business-gateway-url", os.Getenv("LEECLAW_BUSINESS_GATEWAY_URL"), "optional business data gateway base URL")
	businessToken := flag.String("business-gateway-token", os.Getenv("LEECLAW_BUSINESS_GATEWAY_TOKEN"), "optional business data gateway bearer token")
	authMode := flag.String("auth", envOr("COBRA_GRAPH_AUTH_MODE", "weknora"), "authorization mode: weknora or off")
	registryAdminToken := flag.String("registry-admin-token", os.Getenv("COBRA_REGISTRY_ADMIN_TOKEN"), "admin token for ontology registry mutation/read APIs")
	allowedOrigin := flag.String("cors-origin", os.Getenv("COBRA_GRAPH_CORS_ORIGIN"), "optional allowed CORS origin; prefer same-origin reverse proxy")
	runtimeToken := flag.String("runtime-token", os.Getenv("LEECLAW_RUNTIME_TOKEN"), "LeeClaw internal runtime API bearer token")
	promotionToken := flag.String("promotion-token", os.Getenv("LEECLAW_PROMOTION_TOKEN"), "LeeClaw internal Promotion API bearer token")
	catalogOverlayPath := flag.String("catalog-overlay", os.Getenv("COBRA_CATALOG_OVERLAY_FILE"), "optional semantic catalog overlay")
	flag.Parse()

	registry := ontsvc.NewFSRegistry(*registryRoot)
	ontologySource := graphview.NewRegistryOntologySource(registry)
	entitySource := graphview.NewNeo4jHTTPSource(*neo4jURL, *neo4jDB, *neo4jUser, *neo4jPassword)

	var checker httpapi.AccessChecker
	switch strings.ToLower(strings.TrimSpace(*authMode)) {
	case "weknora":
		if strings.TrimSpace(*weknoraURL) == "" { log.Fatal("COBRA_WEKNORA_BASE_URL is required when auth mode is weknora") }
		checker = access.NewWeKnoraAccessChecker(*weknoraURL)
	case "off": log.Print("WARNING: graph API authorization is disabled; use only for local development")
	default: log.Fatalf("unsupported auth mode %q", *authMode)
	}
	if strings.TrimSpace(*registryAdminToken) == "" { log.Print("ontology registry admin API disabled: COBRA_REGISTRY_ADMIN_TOKEN is not configured") }

	var overlay retrieval.CatalogOverlay
	if strings.TrimSpace(*catalogOverlayPath) != "" { if err := store.ReadJSON(*catalogOverlayPath, &overlay); err != nil { log.Fatalf("load catalog overlay: %v", err) } }
	runtimeService := &runtimecontext.Service{DefaultWeKnoraBaseURL:*weknoraURL, Registry:registry, Overlay:overlay, EntitySource:entitySource}
	if strings.TrimSpace(*businessURL) != "" { runtimeService.BusinessGateway = businessadapter.NewHTTPGateway(*businessURL,*businessToken); log.Printf("business data retriever enabled: %s", strings.TrimSpace(*businessURL)) }
	if strings.TrimSpace(*runtimeToken) == "" { log.Print("runtime context API disabled: LEECLAW_RUNTIME_TOKEN is not configured") }

	promotionService := promotion.NewService(promotion.NewFSStore(*promotionRoot), promotion.NewWeKnoraBackend(*weknoraURL), promotion.NewOpenVikingBackend(*openvikingURL,*openvikingAPIKey))
	promotionHTTP := &promotionapi.Handler{Service:promotionService, Token:*promotionToken}
	if strings.TrimSpace(*promotionToken) == "" { log.Print("Promotion API disabled: LEECLAW_PROMOTION_TOKEN is not configured") }

	server := &httpapi.Server{RuntimeContext:runtimeService, RuntimeToken:*runtimeToken, EntitySource:entitySource, OntologySource:ontologySource, AccessChecker:checker, Registry:registry, RegistryAdminToken:*registryAdminToken, AllowedOrigin:*allowedOrigin}
	httpServer := &http.Server{Addr:*listen, Handler:promotionHTTP.Wrap(server.Handler()), ReadHeaderTimeout:5*time.Second, ReadTimeout:20*time.Second, WriteTimeout:30*time.Second, IdleTimeout:60*time.Second}
	log.Printf("LeeClaw Core API v0.10 listening on %s; registry=%s; promotion=%s", *listen,*registryRoot,*promotionRoot)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed { log.Fatal(err) }
}

func envOr(name,fallback string) string { if value:=strings.TrimSpace(os.Getenv(name)); value!="" { return value }; return fallback }
