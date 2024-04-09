NAME=dagster
BINARY=terraform-provider-${NAME}

default: build
.PHONY: default

build: generate
	mkdir -p build/
	go build -o build/$(BINARY)
.PHONY: build

clean:
	go mod tidy
	rm -vrf build/
.PHONY: clean

generate:
	cd internal/client/schema && go run github.com/Khan/genqlient
.PHONY: generate

docs:
	mkdir -p docs
	rm -rf ./docs/images
	go generate ./...
.PHONY: docs

reflex:
	@go install github.com/cespare/reflex@latest
	reflex -r "\.graphql$$" -s -- sh -c "make build"
.PHONY: reflex