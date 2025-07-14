DIR:=$(patsubst %/,%,$(dir $(abspath $(lastword $(MAKEFILE_LIST)))))

LOCAL_BIN:=$(DIR)/bin
GEN_CLAY_BIN:=$(DIR)/bin/protoc-gen-goclay
export GEN_CLAY_BIN
GEN_GO_BIN:=$(DIR)/bin/protoc-gen-go
export GEN_GO_BIN
GEN_GO_GRPC_BIN:=$(DIR)/bin/protoc-gen-go-grpc
export GEN_GO_GRPC_BIN

export THIRD_PARTY_PROTO_PATH:=$(dir $(abspath $(lastword $(MAKEFILE_LIST))))../../integration/third_party/proto
$(info ${THIRD_PARTY_PROTO_PATH})

GREEN:=\033[0;32m
RED:=\033[0;31m
NC=:\033[0m

.protoc-build:
	$(info #Installing binary dependencies...)
	GOBIN=$(LOCAL_BIN) go install -mod=mod github.com/not-for-prod/clay/cmd/protoc-gen-goclay
	GOBIN=$(LOCAL_BIN) go install -mod=mod google.golang.org/protobuf/cmd/protoc-gen-go
	GOBIN=$(LOCAL_BIN) go install -mod=mod google.golang.org/grpc/cmd/protoc-gen-go-grpc

build: .protoc-build
	buf dep update

run-server:
	go run ./main.go

run-client-http:
	go run ./client-http/main.go

run-client-grpc:
	go run ./client-grpc/main.go

generate:
	buf dep update
	buf dep prune
	buf generate