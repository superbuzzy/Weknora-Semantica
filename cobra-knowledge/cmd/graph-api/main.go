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
)

func main() {
	listen := flag.String("listen", envOr("COBRA_GRAPH_API_LISTEN", ":8090"), "listen address")
	bindingsPath := flag.String("ontology-bindings", envOr("COBRA_ONTOLOGY_BINDINGS", "configs/ontology-bindings.json"), "knowledge base to ontology binding file")
	neo4jURL := flag.String("neo4j-url", envOr("COBRA_NEO4J_URL", "http://localhost:7474"), "Neo4j HTTP base URL")
	neo4jDB := flag.String("neo4j-database", envOr("COBRA_NEO4J_DATABASE", "neo4j"), "Neo4j database")
	neo4jUser := flag.String("neo4j-user", envOr("COBRA_NEO4J_USER", "neo4j"), "Neo4j username")
	neo4jPassword := flag.String("neo4j-password", os.Getenv("COBRA_NEO4J_PASSWORD"), "Neo4j password")
	weknoraURL := flag.String("weknora-url", os.Getenv("COBRA_WEKNORA_BASE_URL"), "WeKnora base URL for RBAC delegation")
	authMode := flag.String("auth", envOr("COBRA_GRAPH_AUTH_MODE", "weknora"), "authorization mode: weknora or off")
	allowedOrigin := flag.String("cors-origin", os.Getenv("COBRA_GRAPH_CORS_ORIGIN"), "optional allowed CORS origin; prefer same-origin reverse proxy")
	flag.Parse()

	ontologySource, err := graphview.LoadFileOntologyBindings(*bindingsPath)
	if err != nil {
		log.Fatalf("load ontology bindings: %v", err)
	}
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

	server := &httpapi.Server{
		EntitySource:   entitySource,
		OntologySource: ontologySource,
		AccessChecker:  checker,
		AllowedOrigin:  *allowedOrigin,
	}

	httpServer := &http.Server{
		Addr:              *listen,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("CobraKnowledge graph API v0.3 listening on %s", *listen)
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
