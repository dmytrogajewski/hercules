GOBIN ?= build/bin
GO111MODULE=on
ifneq ($(OS),Windows_NT)
EXE =
else
EXE = .exe
endif
PKG = $(shell go env GOOS)_$(shell go env GOARCH)
TAGS ?=

all: precompile ${GOBIN}/uast${EXE} ${GOBIN}/herr${EXE}

# Pre-compile UAST mappings for faster startup
precompile:
	@echo "Pre-compiling UAST mappings..."
	@mkdir -p ${GOBIN}
	@go run ./build/scripts/precompgen/precompile.go -o pkg/uast/embedded_mappings.gen.go

# Generate UAST mappings for all languages
uastmaps-gen:
	@echo "Generating UAST mappings..."
	@python3 build/scripts/uastmapsgen/gen_uastmaps.py

# Install binaries to system PATH
install: all
	@echo "Installing uast and herr binaries..."
	@if [ -w /usr/local/bin ]; then \
		cp ${GOBIN}/uast${EXE} /usr/local/bin/; \
		cp ${GOBIN}/herr${EXE} /usr/local/bin/; \
		echo "Installed to /usr/local/bin"; \
	elif [ -w $(shell go env GOPATH)/bin ]; then \
		cp ${GOBIN}/uast${EXE} $(shell go env GOPATH)/bin/; \
		cp ${GOBIN}/herr${EXE} $(shell go env GOPATH)/bin/; \
		echo "Installed to $(shell go env GOPATH)/bin"; \
	else \
		echo "Error: Cannot write to /usr/local/bin or $(shell go env GOPATH)/bin"; \
		echo "Please run with sudo or ensure GOPATH/bin is in your PATH"; \
		exit 1; \
	fi

# Run all tests with CGO disabled (for cross-platform compatibility)
test: all
	CGO_ENABLED=1 go test ./...

# Run all tests with CGO disabled and verbose output
testv: all
	CGO_ENABLED=1 go test ./... -v

# Run UAST performance benchmarks (comprehensive suite with organized results)
bench: all
	python3 build/scripts/benchmark/benchmark_runner.py

# Run basic Go benchmarks directly (no organization)
bench-basic: all
	CGO_ENABLED=1 go test -run="^$$" -bench=. -benchmem ./pkg/uast

# Run UAST performance benchmarks with verbose output
benchv: all
	CGO_ENABLED=1 go test -run="^$$" -bench=. -benchmem -v ./pkg/uast

# Run UAST performance benchmarks with CPU profiling
benchcpu: all
	CGO_ENABLED=1 go test -run="^$$" -bench=. -benchmem -cpuprofile=cpu.prof ./pkg/uast

# Run UAST performance benchmarks with memory profiling
benchmem: all
	CGO_ENABLED=1 go test -run="^$$" -bench=. -benchmem -memprofile=mem.prof ./pkg/uast

# Run UAST performance benchmarks with both CPU and memory profiling
benchprofile: all
	CGO_ENABLED=1 go test -run="^$$" -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./pkg/uast

# Run benchmarks and generate performance plots
benchplot: all
	CGO_ENABLED=1 go test -run="^$$" -bench=. -benchmem ./pkg/uast > test/benchmarks/benchmark_results.txt 2>&1
	python3 build/scripts/benchmark/benchmark_plot.py test/benchmarks/benchmark_results.txt

# Run benchmarks with verbose output and generate plots
benchplotv: all
	CGO_ENABLED=1 go test -run="^$$" -bench=. -benchmem -v ./pkg/uast > test/benchmarks/benchmark_results.txt 2>&1
	python3 build/scripts/benchmark/benchmark_plot.py test/benchmarks/benchmark_results.txt

# Run comprehensive benchmark suite with organized results
bench-suite: bench

# Run benchmarks without plots
bench-no-plots: all
	python3 build/scripts/benchmark/benchmark_runner.py --no-plots

# Generate report for latest benchmark run
report:
	python3 build/scripts/benchmark/benchmark_report.py

# Generate report for specific run
report-run:
	python3 build/scripts/benchmark/benchmark_report.py $(RUN_NAME)

# List all benchmark runs
bench-list:
	python3 build/scripts/benchmark/benchmark_report.py --list

# Compare latest run with previous
compare:
	python3 build/scripts/benchmark/benchmark_comparison.py $(shell python3 build/scripts/benchmark/benchmark_report.py --list | head -1)

# Compare specific runs
compare-runs:
	python3 build/scripts/benchmark/benchmark_comparison.py $(CURRENT_RUN) --baseline $(BASELINE_RUN)

# Compare last N benchmark runs (usage: make compare-last N=3)
compare-last:
	@N=$${N:-2}; \
	echo "Comparing last $$N benchmark runs..."; \
	if [ $$N -eq 2 ]; then \
		LATEST=$$(python3 build/scripts/benchmark/benchmark_report.py --list | grep -v "Available benchmark runs:" | head -1 | xargs); \
		PREVIOUS=$$(python3 build/scripts/benchmark/benchmark_report.py --list | grep -v "Available benchmark runs:" | head -2 | tail -1 | xargs); \
		echo "Latest: $$LATEST"; \
		echo "Previous: $$PREVIOUS"; \
		python3 build/scripts/benchmark/benchmark_comparison.py "$$LATEST" --baseline "$$PREVIOUS"; \
	else \
		RUNS=$$(python3 build/scripts/benchmark/benchmark_report.py --list | grep -v "Available benchmark runs:" | head -$$N | tr '\n' ' '); \
		echo "Comparing $$N runs:"; \
		echo "$$RUNS" | nl; \
		python3 build/scripts/benchmark/benchmark_comparison_multi.py "$$RUNS"; \
	fi

# Run benchmarks with profiling and generate plots
benchplotprofile: all
	CGO_ENABLED=1 go test -run="^$$" -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof ./pkg/uast > test/benchmarks/benchmark_results.txt 2>&1
	python3 build/scripts/benchmark/benchmark_plot.py test/benchmarks/benchmark_results.txt

# Run specific benchmark and generate plots (usage: make benchplot-simple BENCH=BenchmarkParse)
benchplot-simple: all
	CGO_ENABLED=1 go test -run="^$$" -bench=$(BENCH) -benchmem ./pkg/uast > test/benchmarks/benchmark_results.txt 2>&1
	python3 build/scripts/benchmark/benchmark_plot.py test/benchmarks/benchmark_results.txt

# Run benchmarks with timeout and generate plots (usage: make benchplot-timeout TIMEOUT=30s)
benchplot-timeout: all
	CGO_ENABLED=1 go test -run="^$$" -bench=. -benchmem -timeout=$(TIMEOUT) ./pkg/uast > test/benchmarks/benchmark_results.txt 2>&1
	python3 build/scripts/benchmark/benchmark_plot.py test/benchmarks/benchmark_results.txt

clean:
	rm -f ./hercules
	rm -f ./uast
	rm -f ./herr
	rm -rf ${GOBIN}/
	rm -f *.prof
	rm -f test/benchmarks/benchmark_results.txt
	rm -rf benchmark_plots/

${GOBIN}/protoc-gen-gogo${EXE}:
	go build -o ${GOBIN}/protoc-gen-gogo${EXE} github.com/gogo/protobuf/protoc-gen-gogo

ifneq ($(OS),Windows_NT)
api/proto/pb.pb.go: api/proto/pb.proto ${GOBIN}/protoc-gen-gogo
	PATH="${PATH}:${GOBIN}" protoc --gogo_out=api/proto --proto_path=api/proto api/proto/pb.proto

api/proto/hercules.pb.go: api/proto/hercules.proto
	protoc --go_out=api/proto --go_opt=paths=source_relative \
		--go-grpc_out=api/proto --go-grpc_opt=paths=source_relative \
		--proto_path=api/proto api/proto/hercules.proto
else
api/proto/pb.pb.go: api/proto/pb.proto ${GOBIN}/protoc-gen-gogo.exe
	export PATH="${PATH};${GOBIN}" && \
	protoc --gogo_out=api/proto --proto_path=api/proto api/proto/pb.proto

api/proto/hercules.pb.go: api/proto/hercules.proto
	protoc --go_out=api/proto --go_opt=paths=source_relative \
		--go-grpc_out=api/proto --go-grpc_opt=paths=source_relative \
		--proto_path=api/proto api/proto/hercules.proto
endif

${GOBIN}/uast${EXE}: cmd/uast/*.go pkg/uast/*.go pkg/uast/*/*.go pkg/uast/*/*/*.go internal/pkg/*.go internal/pkg/*/*.go internal/pkg/*/*/*.go
	LDFLAGS="-X github.com/dmytrogajewski/hercules.BinaryGitHash=$(shell git rev-parse HEAD)"; \
	CGO_ENABLED=1 go build -tags "$(TAGS)" -ldflags "$$LDFLAGS" -o ${GOBIN}/uast${EXE} ./cmd/uast

${GOBIN}/herr${EXE}: cmd/herr/*.go pkg/analyzers/*/*.go internal/pkg/*.go internal/pkg/*/*.go internal/pkg/*/*/*.go
	LDFLAGS="-X github.com/dmytrogajewski/hercules.BinaryGitHash=$(shell git rev-parse HEAD)"; \
	CGO_ENABLED=1 go build -tags "$(TAGS)" -ldflags "$$LDFLAGS" -o ${GOBIN}/herr${EXE} ./cmd/herr
