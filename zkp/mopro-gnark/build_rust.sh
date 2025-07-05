#!/bin/bash

set -e  # Exit on any error

echo "🦀 Building Rust bindings for CGO verify library"
echo "=================================================="

# Check if libverify.so exists
if [ ! -f "libverify.so" ]; then
    echo "❌ ERROR: libverify.so not found. Please run 'make build' first."
    exit 1
fi

# Check if required input files exist
echo "🔍 Checking for required input files..."
required_files=("r1cs.bin" "proving_key.bin" "verifying_key.bin" "witness_input.json")
for file in "${required_files[@]}"; do
    if [ ! -f "$file" ]; then
        echo "⚠️  WARNING: $file not found. The verification may fail."
    else
        echo "✅ Found $file"
    fi
done

# Initialize Rust project if Cargo.toml doesn't exist
if [ ! -f "Cargo.toml" ]; then
    echo "🚀 Initializing Rust project..."
    cargo init --name verify-rust --lib
    echo "📝 Replacing Cargo.toml with our configuration..."
    # The Cargo.toml artifact should be saved to the file
fi

# Create src directory structure
echo "📁 Setting up Rust project structure..."
mkdir -p src/bin

# Set library path for linking
export LIBRARY_PATH="$PWD:$LIBRARY_PATH"
export LD_LIBRARY_PATH="$PWD:$LD_LIBRARY_PATH"

# For macOS
if [[ "$OSTYPE" == "darwin"* ]]; then
    export DYLD_LIBRARY_PATH="$PWD:$DYLD_LIBRARY_PATH"
fi

echo "🔨 Building Rust library..."
cargo build

echo "🔨 Building Rust test binary..."
cargo build --bin test_verify_rust

echo "✅ Rust bindings built successfully!"
echo ""
echo "📋 Available commands:"
echo "  cargo test                    # Run Rust tests"
echo "  cargo run --bin test_verify_rust  # Run Rust test binary"
echo "  cargo build --release        # Build optimized release version"
echo ""
echo "🔧 Environment variables set:"
echo "  LIBRARY_PATH=$LIBRARY_PATH"
echo "  LD_LIBRARY_PATH=$LD_LIBRARY_PATH"
if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "  DYLD_LIBRARY_PATH=$DYLD_LIBRARY_PATH"
fi
echo ""
echo "🎯 To run the Rust test:"
echo "  ./target/debug/test_verify_rust"
echo ""
echo "💡 Or use cargo:"
echo "  cargo run --bin test_verify_rust"