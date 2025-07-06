# 🔐 MoproGnark Compilation Chain

What has been done during the hackathon is integrate the secp256k1 ecdsa verification in gnark and compile it with rust-ffi in order to augment the mopro framework:

The proof is generated using the mopro-gnark rust prover, then verify it on chain as displayed here:
https://explorer.garfield-testnet.zircuit.com/address/0x1D97Feb682eb3B0Ab3467790395C13BE37ec9DEf




Here is described the complete compilation chain for building a Zero-Knowledge proof verification system that works across desktop and Android platforms.

## 📋 **Overview**

The project creates a multi-layered verification system:
- **Go CGO Library**: Core ZK proof verification logic
- **Rust Wrapper**: Safe bindings and cross-platform compatibility  
- **Android App**: React Native mobile interface with JNI integration

## 🏗️ **Architecture**

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Go (CGO)      │───▶│   Rust (FFI)    │───▶│  Android (JNI)  │
│ ZK Verification │    │ Safe Bindings   │    │ React Native UI │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 🖥️ **Desktop Compilation (macOS/Linux)**

### **Prerequisites**
- Go 1.21+
- Rust toolchain
- GCC/Clang compiler

### **Build Steps**

1. **Build Go CGO Library**:
   ```bash
   go build -buildmode=c-shared -o libverify.so main.go
   ```

2. **Build Rust Wrapper**:
   ```bash
   cargo add libc #this is for Termux compilation
   cargo build --release
   ```

3. **Run Tests**:
   ```bash
   cargo run --bin test_verify_rust
   ./test_verify  # C test
   ```

### **Files Generated**
- `libverify.so` - Go shared library
- `libverify.h` - Auto-generated C header
- `target/release/libverify_rust.so` - Rust library

## 📱 **Android Compilation Chain**

### **Challenge: Cross-Architecture Compatibility**

The main challenge is that Go CGO libraries compiled on desktop (x86_64/ARM64 macOS) are **not compatible** with Android ARM64. This requires a different approach.

### **Solution: Termux Compilation**

**Termux** (Android Linux environment) allows compiling native ARM64 libraries directly on Android devices.

#### **Phase 1: ARM64 Library via Termux**

1. **Install Termux on Android device**:
   ```bash
   # In Termux app
   pkg update && pkg upgrade
   pkg install golang clang make git
   ```

2. **Set up build environment**:
   ```bash
   export CGO_ENABLED=1
   export CC=clang
   export CXX=clang++
   ```

3. **Transfer files to Android**:
   - Copy `main.go`, `*.bin`, `*.json` to Termux directory

4. **Build ARM64 CGO library**:
   ```bash
   # In Termux
   go build -buildmode=c-shared -o libverify.so main.go
   ```

5. **Build Rust wrapper**:
   ```bash
   # Create simplified Cargo.toml and src/lib.rs
   cargo build --lib
   cargo build --bin test_verify_rust
   ./target/debug/test_verify_rust
   ```

#### **Phase 2: Android App Integration**

1. **Transfer ARM64 library**:
   ```bash
   # Copy from Termux to development machine
   cp libverify.so libverify_arm64.so
   ```

2. **Build React Native app**:
   ```bash
   ./build_android_rust.sh
   cd react-native-app
   npm install
   npx react-native run-android
   ```

## 📁 **File Structure**

```
project/
├── main.go                     # Go ZK verification logic
├── libverify.so               # Compiled Go library
├── Cargo.toml                 # Rust project configuration
├── build.rs                   # Rust build script
├── src/
│   ├── lib.rs                 # Rust FFI bindings
│   ├── android.rs             # Android JNI interface
│   └── bin/
│       └── test_verify_rust.rs # Rust test binary
├── react-native-app/
│   ├── App.tsx                # React Native UI
│   ├── android/
│   │   └── app/src/main/
│   │       ├── java/com/zkverify/ # Java JNI modules
│   │       ├── jniLibs/           # Native libraries
│   │       └── assets/            # ZK proof files
│   └── package.json
├── *.bin                      # ZK circuit files
├── *.json                     # Witness input data
└── README.md
```

## 🔧 **Key Technical Details**

### **CGO to Rust FFI**
```rust
// Rust side - unsafe extern block
unsafe extern "C" {
    fn verify() -> *mut c_char;
    fn free(ptr: *mut c_char);
}

// Safe wrapper
pub fn verify_proof() -> Result<String, VerifyError> {
    unsafe {
        let c_str_ptr = verify();
        // ... string conversion and memory management
        free(c_str_ptr);
    }
}
```

### **Rust to Android JNI**
```rust
// JNI export for Android
#[no_mangle]
pub extern "system" fn Java_com_zkverify_MainActivity_verifyProof(
    mut env: JNIEnv,
    _class: JClass,
) -> jstring {
    let result = verify_proof().unwrap_or_else(|e| format!("Error: {}", e));
    env.new_string(&result).unwrap().into_raw()
}
```

### **Android Java Integration**
```java
// Java side - load native library
static {
    System.loadLibrary("verify_rust");
}

// Native method declarations
public static native String verifyProof();
```

## 🎯 **Platform Support**

| Platform | Go Library | Rust Wrapper | App Interface |
|----------|------------|--------------|---------------|
| **macOS Desktop** | ✅ Native | ✅ Native | ❌ N/A |
| **Linux Desktop** | ✅ Native | ✅ Native | ❌ N/A |
| **Android ARM64** | ✅ Termux | ✅ Cross-compiled | ✅ React Native |
| **Android x86** | ❌ Mock only | ✅ Cross-compiled | ✅ React Native |
| **iOS** | 🔄 Possible | 🔄 Possible | 🔄 Possible |

## 🚀 **Quick Start Commands**

### **Desktop Development**:
```bash
# Build and test everything
make build
make rust-test
```

### **Android Development**:
```bash
# 1. Build ARM64 library in Termux (on Android device)
go build -buildmode=c-shared -o libverify.so main.go

# 2. Build Android app (on development machine)
./build_android_rust.sh
cd react-native-app && npm install && npx react-native run-android
```

## 🛠️ **Troubleshooting**

### **Common Issues**:

1. **Library not found**: Ensure `libverify.so` exists and is in the correct path
2. **Architecture mismatch**: Use Termux for ARM64 Android compatibility
3. **JNI linking errors**: Check that `System.loadLibrary("verify_rust")` matches the actual library name
4. **Memory management**: Ensure proper `free()` calls for CGO-allocated strings

### **Debug Commands**:
```bash
# Check library architecture
file libverify.so

# View JNI logs
adb logcat | grep -E "(ZKVerify|verify_rust)"

# Test indiv