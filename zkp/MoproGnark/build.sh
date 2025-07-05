#!/bin/bash

set -e  # Exit on any error

echo "Building CGO shared library and test executable..."
echo "=================================================="

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

# Build the shared library
echo "Building shared library..."
go build -buildmode=c-shared -o libverify.so main.go

# Verify the shared library was created
if [ ! -f "libverify.so" ]; then
    echo "❌ ERROR: Failed to build libverify.so"
    exit 1
fi

echo "✅ Shared library libverify.so created successfully"

# Verify the header file was created
if [ ! -f "libverify.h" ]; then
    echo "❌ ERROR: libverify.h was not generated"
    exit 1
fi

echo "✅ Header file libverify.h generated successfully"

# Compile the C test program
echo "Compiling C test program..."

# Detect OS and use appropriate linker flags
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    echo "Detected macOS - using macOS linker flags"
    gcc -o test_verify test_verify.c -L. -lverify -Wl,-rpath,.
elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
    # Linux
    echo "Detected Linux - using Linux linker flags"
    gcc -o test_verify test_verify.c -L. -lverify -Wl,-rpath=.
else
    # Try generic approach
    echo "Unknown OS - trying generic linker flags"
    gcc -o test_verify test_verify.c -L. -lverify
fi

# Verify the test executable was created
if [ ! -f "test_verify" ]; then
    echo "❌ ERROR: Failed to build test_verify executable"
    exit 1
fi

echo "✅ Test executable test_verify created successfully"
echo ""
echo "Build completed successfully!"
echo "=================================================="
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
echo "  - libverify.so (shared library)"
echo "  - libverify.h (auto-generated header)"
echo "  - test_verify (test executable)"