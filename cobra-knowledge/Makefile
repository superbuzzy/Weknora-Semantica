.PHONY: test build demo clean

test:
	go test ./...

build:
	mkdir -p bin
	go build -o bin/cobra-knowledge ./cmd/cobra
	go build -o bin/cobra-context-mcp ./cmd/context-mcp

demo: build
	bin/cobra-knowledge bootstrap -in examples/weknora-graph.json -domain distribution_network -out out/bootstrap
	bin/cobra-knowledge plan -ontology out/bootstrap/candidate-ontology.json -overlay configs/semantic-overlay.json -query '金牛线为什么转供能力不足？' -out out/retrieval-plan.json

clean:
	rm -rf bin out
