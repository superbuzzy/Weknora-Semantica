package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"cobraknowledge.local/cobra-knowledge/internal/access"
	"cobraknowledge.local/cobra-knowledge/internal/graphview"
	"cobraknowledge.local/cobra-knowledge/internal/httpapi"
	ontsvc "cobraknowledge.local/cobra-knowledge/internal/ontology"
)

func main() {
	listen := flag.String("listen", envOr("COBRA_GRAPH_API_LISTEN", ":8090"), "listen address")
	registryRoot := flag.String("registry-root", envOr("COBRA_ONTOLOGY_REGISTRY_ROOT", "var/ontology-registry"), "ontology registry root")
	neo4jURL := flag.String("neo4j-url", envOr("COBRA_NEO4J_URL", "http://localhost:7474"), "Neo4j HTTP base URL")
	neo4jDB := flag.String("neo4j-database", envOr("COBRA_NEO4J_DATABASE", "neo4j"), "Neo4j database")
	neo4jUser := flag.String("neo4j-user", envOr("COBRA_NEO4J_USER", "neo4j"), "Neo4j username")
	neo4jPassword := flag.String("neo4j-password", os.Getenv("COBRA_NEO4J_PASSWORD"), "Neo4j password")
	weknoraURL := flag.String("weknora-url", os.Getenv("COBRA_WEKNORA_BASE_URL"), "WeKnora base URL for RBAC delegation")
	authMode := flag.String("auth", envOr("COBRA_GRAPH_AUTH_MODE", "weknora"), "authorization mode: weknora or off")
	registryAdminToken := flag.String("registry-admin-token", os.Getenv("COBRA_REGISTRY_ADMIN_TOKEN"), "admin token for ontology registry mutation/read APIs")
	allowedOrigin := flag.String("cors-origin", os.Getenv("COBRA_GRAPH_CORS_ORIGIN"), "optional allowed CORS origin; prefer same-origin reverse proxy")
	flag.Parse()

	registry := ontsvc.NewFSRegistry(*registryRoot)
	ontologySource := graphview.NewRegistryOntologySource(registry)
	entitySource := graphview.NewNeo4jHTTPSource(*neo4jURL, *neo4jDB, *neo4jUser, *neo4jPassword)

	var checker httpapi.AccessChecker
	switch strings.ToLower(strings.TrimSpace(*authMode)) {
	case "weknora":
		if strings.TrimSpace(*weknoraURL) == "" {
			log.Fatal("COBRA_WEKNORA_BASE_URL is required when auth mode is weknora")
		}
		checker = access.NewWeKnoraAccessChecker(*weknoraURL)
	case "off":
		log.Print("WARNING: graph API authorization is disabled; use only for local development")
	default:
		log.Fatalf("unsupported auth mode %q", *authMode)
	}
	if strings.TrimSpace(*registryAdminToken) == "" {
		log.Print("ontology registry admin API disabled: COBRA_REGISTRY_ADMIN_TOKEN is not configured")
	}

	server := &httpapi.Server{
		EntitySource:       entitySource,
		OntologySource:     ontologySource,
		AccessChecker:      checker,
		Registry:           registry,
		RegistryAdminToken: *registryAdminToken,
		AllowedOrigin:      *allowedOrigin,
	}

	httpServer := &http.Server{
		Addr:              *listen,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("CobraKnowledge API v0.6 listening on %s; registry=%s", *listen, *registryRoot)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func envOr(name, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(name)); value != "" {
		return value
	}
	return fallback
}
