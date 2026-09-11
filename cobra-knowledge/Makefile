.PHONY: test vet build plugin-test verify demo clean

test:
	go test ./...

vet:
	go vet ./...

plugin-test:
	node --test integrations/openclaw/knowledge-plugin/test/*.test.js
	node --test integrations/openclaw/openviking-plugin/test/*.test.js

verify:
	./scripts/verify-v0.7.sh

build:
	mkdir -p bin
	go build -o bin/cobra-knowledge ./cmd/cobra
	go build -o bin/cobra-context-mcp ./cmd/context-mcp
	go build -o bin/cobra-graph-api ./cmd/graph-api

demo: build
	bin/cobra-knowledge bootstrap -in examples/weknora-graph.json -domain distribution_network -out out/bootstrap
	bin/cobra-knowledge plan -ontology out/bootstrap/candidate-ontology.json -overlay configs/semantic-overlay.json -query '金牛线为什么转供能力不足？' -out out/retrieval-plan.json

clean:
	rm -rf bin out
