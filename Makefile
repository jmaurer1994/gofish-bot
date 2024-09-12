.PHONY: compile generate build run clean protos templates

ifndef BINARY_NAME
BINARY_NAME=gofish-bot
endif

compile:
	@echo "Compiling $(BINARY_NAME) for all targets"
	GOARCH=amd64 GOOS=darwin go build -o ./bin/$(BINARY_NAME)-darwin ./cmd/$(BINARY_NAME)
	GOARCH=amd64 GOOS=linux go build -o ./bin/$(BINARY_NAME)-linux ./cmd/$(BINARY_NAME)
	GOARCH=amd64 GOOS=windows go build -o ./bin/$(BINARY_NAME).exe ./cmd/$(BINARY_NAME)

# Generates protobuf files and templ components 
generate: protos components

build: generate
	@echo "Building $(BINARY_NAME)"
	go build -o ./bin/$(BINARY_NAME) ./cmd/$(BINARY_NAME)

run: build
	@echo "Running $(BINARY_NAME)"
	./bin/$(BINARY_NAME)

dev: generate
	@echo "Running in dev mode"
	@templ generate --watch --cmd "go run ./cmd/overlay-dev"

clean:
	@echo "Cleaning"
	go clean
	-rm ./bin/$(BINARY_NAME)-darwin
	-rm ./bin/$(BINARY_NAME)-linux
	-rm ./bin/$(BINARY_NAME).exe

protos:
	@echo "Generating protobuf files"
	python -m grpc_tools.protoc \
		--python_out=./scripts/infer/protos \
		--grpc_python_out=./scripts/infer/protos \
		--proto_path=./protos \
		inferencetask.proto \
		
	protoc \
		--go_out=. \
		--go_opt=Minferencetask.proto=github.com/jmaurer1994/gofish-bot/internal/infer/protos \
		--go_opt=module=github.com/jmaurer1994/gofish-bot \
		--go-grpc_out=. \
		--go-grpc_opt=Minferencetask.proto=github.com/jmaurer1994/gofish-bot/internal/infer/protos \
		--go-grpc_opt=module=github.com/jmaurer1994/gofish-bot \
		--proto_path=./protos \
		inferencetask.proto \

components:
	@echo "Generating templates"
	@templ generate .
