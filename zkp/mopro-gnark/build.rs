use std::env;
use std::path::PathBuf;

fn main() {
    // Get the current directory where libverify.so should be located
    let current_dir = env::current_dir().expect("Failed to get current directory");
    
    // Tell cargo to look for shared libraries in the current directory
    println!("cargo:rustc-link-search=native={}", current_dir.display());
    
    // Tell cargo to link against the verify library
    println!("cargo:rustc-link-lib=dylib=verify");
    
    // For macOS, we need to set the rpath
    if cfg!(target_os = "macos") {
        println!("cargo:rustc-link-arg=-Wl,-rpath,{}", current_dir.display());
    } else if cfg!(target_os = "linux") {
        println!("cargo:rustc-link-arg=-Wl,-rpath={}", current_dir.display());
    }
    
    // Tell cargo to rerun this build script if libverify.so changes
    let lib_path = current_dir.join("libverify.so");
    println!("cargo:rerun-if-changed={}", lib_path.display());
    
    // Check if the library exists
    if !lib_path.exists() {
        panic!("libverify.so not found at {}. Please run 'make build' first.", lib_path.display());
    }
    
    println!("cargo:warning=Using libverify.so from: {}", lib_path.display());
}