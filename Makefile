GOBIN ?= .
GO111MODULE=on
ifneq ($(OS),Windows_NT)
EXE =
else
EXE = .exe
endif
PKG = $(shell go env GOOS)_$(shell go env GOARCH)
TAGS ?=

all: ${GOBIN}/hercules${EXE}

# Run all tests with CGO disabled (for cross-platform compatibility)
test: all
	CGO_ENABLED=0 go test ./...

# Run all tests with CGO disabled and verbose output
testv: all
	CGO_ENABLED=0 go test ./... -v

${GOBIN}/protoc-gen-gogo${EXE}:
	go build github.com/gogo/protobuf/protoc-gen-gogo

ifneq ($(OS),Windows_NT)
internal/pb/pb.pb.go: internal/pb/pb.proto ${GOBIN}/protoc-gen-gogo
	PATH="${PATH}:${GOBIN}" protoc --gogo_out=internal/pb --proto_path=internal/pb internal/pb/pb.proto

internal/pb/hercules.pb.go: internal/pb/hercules.proto
	protoc --go_out=internal/pb --go_opt=paths=source_relative \
		--go-grpc_out=internal/pb --go-grpc_opt=paths=source_relative \
		--proto_path=internal/pb internal/pb/hercules.proto
else
internal/pb/pb.pb.go: internal/pb/pb.proto ${GOBIN}/protoc-gen-gogo.exe
	export PATH="${PATH};${GOBIN}" && \
	protoc --gogo_out=internal/pb --proto_path=internal/pb internal/pb/pb.proto

internal/pb/hercules.pb.go: internal/pb/hercules.proto
	protoc --go_out=internal/pb --go_opt=paths=source_relative \
		--go-grpc_out=internal/pb --go-grpc_opt=paths=source_relative \
		--proto_path=internal/pb internal/pb/hercules.proto
endif

cmd/hercules/plugin_template_source.go: cmd/hercules/plugin.template
	cd cmd/hercules && go generate

${GOBIN}/hercules${EXE}: *.go */*.go */*/*.go */*/*/*.go internal/pb/pb.pb.go internal/pb/hercules.pb.go cmd/hercules/plugin_template_source.go
	LDFLAGS="-X gopkg.in/src-d/hercules.v10.BinaryGitHash=$(shell git rev-parse HEAD)"; \
	CGO_ENABLED=0 go build -tags "$(TAGS)" -ldflags "$$LDFLAGS" -o ${GOBIN}/hercules${EXE} ./cmd/hercules
