# Makefile for CGO verify library, C test, and Rust bindings

.PHONY: all clean build test run help rust-build rust-test rust-run rust-clean

# Default target
all: build

# Build the shared library and test executable
build:
	@echo "Building CGO shared library and test executable..."
	@chmod +x build.sh
	@./build.sh

# Run the C test
test: build
	@echo "Running C verification test..."
	@./test_verify

# Run the C test (alias for test)
run: test

# Build Rust bindings
rust-build: build
	@echo "Building Rust bindings..."
	@chmod +x build_rust.sh
	@./build_rust.sh

# Run Rust tests
rust-test: rust-build
	@echo "Running Rust tests..."
	@export LIBRARY_PATH="$PWD:$LIBRARY_PATH"; \
	 export LD_LIBRARY_PATH="$PWD:$LD_LIBRARY_PATH"; \
	 if [[ "$OSTYPE" == "darwin"* ]]; then \
	   export DYLD_LIBRARY_PATH="$PWD:$DYLD_LIBRARY_PATH"; \
	 fi; \
	 cargo test --lib

# Run Rust example with JSON features
rust-example: rust-build
	@echo "Running Rust example with JSON features..."
	@export LIBRARY_PATH="$PWD:$LIBRARY_PATH"; \
	 export LD_LIBRARY_PATH="$PWD:$LD_LIBRARY_PATH"; \
	 if [[ "$OSTYPE" == "darwin"* ]]; then \
	   export DYLD_LIBRARY_PATH="$PWD:$DYLD_LIBRARY_PATH"; \
	 fi; \
	 cargo run --example simple_app --features json

# Run basic example
rust-basic: rust-build
	@echo "Running basic Rust example..."
	@export LIBRARY_PATH="$PWD:$LIBRARY_PATH"; \
	 export LD_LIBRARY_PATH="$PWD:$LD_LIBRARY_PATH"; \
	 if [[ "$OSTYPE" == "darwin"* ]]; then \
	   export DYLD_LIBRARY_PATH="$PWD:$DYLD_LIBRARY_PATH"; \
	 fi; \
	 cargo run --example basic_usage

# Run Rust test binary
rust-run: rust-build
	@echo "Running Rust test binary..."
	@cd . && cargo run --bin test_verify_rust

# Run both C and Rust tests
test-all: test rust-test rust-run
	@echo "✅ All tests completed!"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -f libverify.so libverify.h test_verify
	@rm -rf solidity/
	@echo "C artifacts cleaned."

# Clean Rust artifacts
rust-clean:
	@echo "Cleaning Rust artifacts..."
	@if [ -f "Cargo.toml" ]; then cargo clean; fi
	@rm -rf target/
	@echo "Rust artifacts cleaned."

# Clean everything
clean-all: clean rust-clean
	@echo "All artifacts cleaned."

# Display help
help:
	@echo "Available targets:"
	@echo "  all        - Build the shared library and test executable (default)"
	@echo "  build      - Build the shared library and test executable"
	@echo "  test       - Build and run the C verification test"
	@echo "  run        - Alias for test"
	@echo ""
	@echo "Rust targets:"
	@echo "  rust-build   - Build Rust bindings"
	@echo "  rust-test    - Build and run Rust tests"
	@echo "  rust-run     - Build and run Rust test binary"
	@echo "  rust-basic   - Run basic Rust example"
	@echo "  rust-example - Run advanced Rust example with JSON"
	@echo "  test-all     - Run both C and Rust tests"
	@echo ""
	@echo "Cleanup targets:"
	@echo "  clean      - Remove C build artifacts"
	@echo "  rust-clean - Remove Rust build artifacts"
	@echo "  clean-all  - Remove all build artifacts"
	@echo "  help       - Display this help message"
	@echo ""
	@echo "Prerequisites:"
	@echo "  - Go 1.21+ installed"
	@echo "  - GCC compiler"
	@echo "  - Rust toolchain (cargo, rustc)"
	@echo "  - Required input files: r1cs.bin, proving_key.bin, verifying_key.bin, witness_input.json"

# Build only the shared library
lib:
	@echo "Building shared library only..."
	@go build -buildmode=c-shared -o libverify.so main.go
	@echo "✅ Shared library libverify.so created"

# Build only the test executable (requires existing shared library)
testbin: libverify.so
	@echo "Building test executable..."
	@if [[ "$$OSTYPE" == "darwin"* ]]; then \
		gcc -o test_verify test_verify.c -L. -lverify -Wl,-rpath,.; \
	elif [[ "$$OSTYPE" == "linux-gnu"* ]]; then \
		gcc -o test_verify test_verify.c -L. -lverify -Wl,-rpath=.; \
	else \
		gcc -o test_verify test_verify.c -L. -lverify; \
	fi
	@echo "✅ Test executable test_verify created"