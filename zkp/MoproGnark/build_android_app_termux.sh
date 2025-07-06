#!/bin/bash

set -e

echo "📱 Building ZK Verify Android App"
echo "=================================="

# Check if we have the Termux-compiled ARM64 library
ARM64_LIB="libverify_arm64.so"
if [ ! -f "$ARM64_LIB" ]; then
    echo "⚠️  WARNING: $ARM64_LIB not found"
    echo "Using local libverify.so (make sure it's ARM64 compatible)"
    ARM64_LIB="libverify.so"
fi

if [ ! -f "$ARM64_LIB" ]; then
    echo "❌ ERROR: No ARM64 library found"
    echo "Please ensure you have either:"
    echo "  - libverify_arm64.so (from Termux compilation)"
    echo "  - libverify.so (ARM64 compatible)"
    exit 1
fi

echo "✅ Using ARM64 library: $ARM64_LIB"

# Check for Android targets
echo "🔍 Checking Android Rust targets..."
ANDROID_TARGETS=("aarch64-linux-android")

for target in "${ANDROID_TARGETS[@]}"; do
    if ! rustup target list --installed | grep -q "$target"; then
        echo "📦 Installing Android target: $target"
        rustup target add "$target"
    else
        echo "✅ Android target already installed: $target"
    fi
done

# Check NDK
if [ -z "$ANDROID_NDK_ROOT" ] && [ -z "$NDK_HOME" ]; then
    echo "❌ ERROR: ANDROID_NDK_ROOT or NDK_HOME not set"
    echo "Please set your Android NDK path:"
    echo "export ANDROID_NDK_ROOT=/path/to/android-ndk"
    exit 1
fi

# Install cargo-ndk if needed
if ! command -v cargo-ndk &> /dev/null; then
    echo "📦 Installing cargo-ndk..."
    cargo install cargo-ndk
fi

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

# Build Rust wrapper for ARM64
echo "🔨 Building Rust JNI wrapper for ARM64..."

# First, ensure we have the android feature in Cargo.toml
if ! grep -q 'jni.*=.*{.*version.*=.*"0.21".*optional.*=.*true.*}' Cargo.toml; then
    echo "📝 Adding Android dependencies to Cargo.toml..."
    cat >> Cargo.toml << 'EOF'

# Android JNI dependencies
jni = { version = "0.21", optional = true }

[features]
android = ["jni"]
EOF
fi

# Add Android module to lib.rs if not present
if ! grep -q "pub mod android" src/lib.rs; then
    echo "📝 Adding Android module to lib.rs..."
    sed -i '1i#[cfg(feature = "android")]\npub mod android;\n' src/lib.rs
fi

# Build ARM64 Rust library
cargo ndk -t aarch64-linux-android -o "$ANDROID_LIB_DIR/arm64-v8a" build --release --features android

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
lib_path="$ANDROID_LIB_DIR/arm64-v8a/libverify_rust.so"
if [ -f "$lib_path" ]; then
    size=$(ls -lh "$lib_path" | awk '{print $5}')
    echo "✅ ARM64 Rust library: libverify_rust.so ($size)"
else
    echo "❌ ARM64 Rust library not found!"
    exit 1
fi

echo ""
echo "🎉 Android app build completed!"
echo "==============================="
echo ""
echo "📱 Next steps:"
echo "1. cd $RN_APP_DIR"
echo "2. npm install"
echo "3. npx react-native run-android"
echo ""
echo "📋 Make sure you have:"
echo "- Android device connected or emulator running"
echo "- Android Studio and SDK installed"
echo ""
echo "🎯 The app will have:"
echo "- Single 'Generate Proof' button"
echo "- Real ZK verification on ARM64 devices"
echo "- Clean, modern UI with result display"
echo ""
echo "🔧 Built components:"
echo "- React Native UI: $RN_APP_DIR/"
echo "- ARM64 Go library: $ANDROID_LIB_DIR/arm64-v8a/libverify.so"
echo "- ARM64 Rust wrapper: $ANDROID_LIB_DIR/arm64-v8a/libverify_rust.so"
echo "- ZK assets: $ASSETS_DIR/"