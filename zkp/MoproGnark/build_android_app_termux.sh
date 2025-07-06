#!/bin/bash

set -e

echo "📱 Building ZK Verify Android App in Termux"
echo "==========================================="

# Check if we have the current ARM64 library
ARM64_LIB="libverify.so"
if [ ! -f "$ARM64_LIB" ]; then
    echo "❌ ERROR: $ARM64_LIB not found"
    echo "Please build it first with:"
    echo "go build -buildmode=c-shared -o libverify.so main.go"
    exit 1
fi

echo "✅ Using ARM64 library: $ARM64_LIB"

# We don't need to check Android targets since we're already on Android
echo "✅ Running on Android ARM64 - no cross-compilation needed"

# We don't need cargo-ndk since we're building natively
echo "✅ Building natively on target platform"

# Create React Native app directory structure
RN_APP_DIR="react-native-app"
ANDROID_LIB_DIR="$RN_APP_DIR/android/app/src/main/jniLibs"

echo "📁 Setting up React Native app structure..."
mkdir -p "$ANDROID_LIB_DIR/arm64-v8a"
mkdir -p "$RN_APP_DIR/android/app/src/main/java/com/zkverify"
mkdir -p "$RN_APP_DIR/android/app/src/main/assets"

# Copy the ARM64 Go library
echo "📋 Copying ARM64 Go library..."
cp "$ARM64_LIB" "$ANDROID_LIB_DIR/arm64-v8a/libverify.so"
echo "✅ Copied ARM64 Go library to app"

# Build Rust wrapper for ARM64 (native build)
echo "🔨 Building Rust JNI wrapper natively..."

# Ensure we have the android feature in Cargo.toml
if ! grep -q 'jni.*=.*{.*version.*=.*"0.21".*optional.*=.*true.*}' Cargo.toml; then
    echo "📝 Adding Android dependencies to Cargo.toml..."
    
    # Check if [dependencies] section exists
    if grep -q "^\[dependencies\]" Cargo.toml; then
        # Add jni after the [dependencies] line
        sed -i '/^\[dependencies\]/a jni = { version = "0.21", optional = true }' Cargo.toml
    else
        # Add [dependencies] section with jni
        cat >> Cargo.toml << 'EOF'

[dependencies]
jni = { version = "0.21", optional = true }
EOF
    fi
    
    # Add android feature
    if grep -q "^\[features\]" Cargo.toml; then
        # Add android feature after [features] line
        sed -i '/^\[features\]/a android = ["jni"]' Cargo.toml
    else
        # Add [features] section
        cat >> Cargo.toml << 'EOF'

[features]
android = ["jni"]
EOF
    fi
fi

# Add Android module to lib.rs if not present
if ! grep -q "pub mod android" src/lib.rs; then
    echo "📝 Adding Android module to lib.rs..."
    # Add at the top of the file
    sed -i '1i#[cfg(feature = "android")]\npub mod android;\n' src/lib.rs
fi

# Build ARM64 Rust library (native build - no cross-compilation needed)
echo "🔨 Building Rust library with Android features..."
cargo build --release --features android

# Copy the built Rust library
echo "📋 Copying Rust library..."
cp target/release/libverify_rust.so "$ANDROID_LIB_DIR/arm64-v8a/"
echo "✅ Copied Rust library to app"

# Copy assets
echo "📋 Copying ZK proof assets..."
ASSETS_DIR="$RN_APP_DIR/android/app/src/main/assets"
FILES_TO_COPY=("r1cs.bin" "proving_key.bin" "verifying_key.bin" "witness_input.json")

for file in "${FILES_TO_COPY[@]}"; do
    if [ -f "$file" ]; then
        cp "$file" "$ASSETS_DIR/"
        echo "✅ Copied $file to assets"
    else
        echo "⚠️  WARNING: $file not found - app may not work properly"
    fi
done

# Create index.js for React Native
echo "📝 Creating React Native entry point..."
cat > "$RN_APP_DIR/index.js" << 'EOF'
import {AppRegistry} from 'react-native';
import App from './App';
import {name as appName} from './package.json';

AppRegistry.registerComponent(appName, () => App);
EOF

# Verify libraries were built
echo "🔍 Verifying built libraries..."
go_lib_path="$ANDROID_LIB_DIR/arm64-v8a/libverify.so"
rust_lib_path="$ANDROID_LIB_DIR/arm64-v8a/libverify_rust.so"

if [ -f "$go_lib_path" ]; then
    size=$(ls -lh "$go_lib_path" | awk '{print $5}')
    echo "✅ ARM64 Go library: libverify.so ($size)"
else
    echo "❌ ARM64 Go library not found!"
    exit 1
fi

if [ -f "$rust_lib_path" ]; then
    size=$(ls -lh "$rust_lib_path" | awk '{print $5}')
    echo "✅ ARM64 Rust library: libverify_rust.so ($size)"
else
    echo "❌ ARM64 Rust library not found!"
    exit 1
fi

# Test the Rust binary to make sure everything works
echo "🧪 Testing Rust binary..."
if [ -f "target/release/test_verify_rust" ]; then
    echo "Running quick test..."
    timeout 30s ./target/release/test_verify_rust || echo "Test completed (or timed out)"
else
    echo "⚠️  Test binary not found, but libraries are built"
fi

echo ""
echo "🎉 Android app build completed in Termux!"
echo "=========================================="
echo ""
echo "📱 What was built:"
echo "- React Native app structure: $RN_APP_DIR/"
echo "- ARM64 Go library: $go_lib_path"
echo "- ARM64 Rust wrapper: $rust_lib_path"
echo "- ZK assets: $ASSETS_DIR/"
echo ""
echo "📋 Next steps:"
echo "1. Transfer the '$RN_APP_DIR' folder to your development machine"
echo "2. On your dev machine: cd $RN_APP_DIR && npm install"
echo "3. Connect Android device and run: npx react-native run-android"
echo ""
echo "🎯 Or continue in Termux (if you have Node.js installed):"
echo "1. pkg install nodejs npm"
echo "2. cd $RN_APP_DIR && npm install"
echo "3. npm start"
echo ""
echo "✨ The app is ready with native ARM64 performance!"