#!/bin/bash

set -e  # Exit on any error

echo "Building optimized CGO shared library and test executable..."
echo "============================================================="

# Clean up previous builds
echo "Cleaning up previous builds..."
rm -f libverify.so libverify.h test_verify

# Initialize Go module if go.mod doesn't exist
if [ ! -f "go.mod" ]; then
    echo "Initializing Go module..."
    go mod init verify
fi

# Download dependencies
echo "Downloading Go dependencies..."
go mod tidy

# Build the shared library with maximum optimizations
echo "Building optimized shared library..."

# Set optimization environment variables
export CGO_ENABLED=1
export CGO_CFLAGS="-O3 -march=native -mtune=native -flto"
export CGO_LDFLAGS="-O3 -flto"

# Build with Go optimizations
go build \
    -buildmode=c-shared \
    -ldflags="-s -w" \
    -gcflags="-B" \
    -a \
    -o libverify.so \
    main.go

# Verify the shared library was created
if [ ! -f "libverify.so" ]; then
    echo "❌ ERROR: Failed to build libverify.so"
    exit 1
fi

echo "✅ Optimized shared library libverify.so created successfully"

# Verify the header file was created
if [ ! -f "libverify.h" ]; then
    echo "❌ ERROR: libverify.h was not generated"
    exit 1
fi

echo "✅ Header file libverify.h generated successfully"

# Compile the C test program with optimizations
echo "Compiling optimized C test program..."

# Detect OS and use appropriate linker flags with optimizations
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    echo "Detected macOS - using optimized macOS linker flags"
    gcc -O3 -march=native -mtune=native -flto \
        -o test_verify test_verify.c \
        -L. -lverify -Wl,-rpath,.
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    echo "Detected Linux - using optimized Linux linker flags"
    gcc -O3 -march=native -mtune=native -flto \
        -o test_verify test_verify.c \
        -L. -lverify -Wl,-rpath=.
else
    # Try generic approach with optimizations
    echo "Unknown OS - trying optimized generic linker flags"
    gcc -O3 -o test_verify test_verify.c -L. -lverify
fi

# Verify the test executable was created
if [ ! -f "test_verify" ]; then
    echo "❌ ERROR: Failed to build test_verify executable"
    exit 1
fi

echo "✅ Optimized test executable test_verify created successfully"

# Optional: Strip symbols for smaller binaries (comment out if you need debugging)
echo "Stripping debug symbols for smaller binaries..."
strip libverify.so 2>/dev/null || echo "Note: strip command not available or failed"
strip test_verify 2>/dev/null || echo "Note: strip command not available or failed"

echo ""
echo "Optimized build completed successfully!"
echo "======================================"
echo ""
echo "Optimization flags used:"
echo "  Go: -ldflags='-s -w' -gcflags='-B' -a"
echo "  CGO: -O3 -march=native -mtune=native -flto"
echo "  GCC: -O3 -march=native -mtune=native -flto"
echo ""
echo "To run the test:"
echo "  ./test_verify"
echo ""
echo "Make sure the following files are in the current directory:"
echo "  - r1cs.bin"
echo "  - proving_key.bin"
echo "  - verifying_key.bin"
echo "  - witness_input.json"
echo ""
echo "Files created:"
echo "  - libverify.so (optimized shared library)"
echo "  - libverify.h (auto-generated header)"
echo "  - test_verify (optimized test executable)"