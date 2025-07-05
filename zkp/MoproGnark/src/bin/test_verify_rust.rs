use verify_rust::{verify_proof, verify_proof_simple};

fn main() {
    println!("🦀 Rust ZK Verification Test");
    println!("========================================");
    
    // Test the detailed verification function
    println!("\n📋 Running detailed verification...");
    match verify_proof() {
        Ok(result) => {
            println!("✅ Verification completed successfully!");
            println!("📄 Full result:");
            println!("{}", result);
        }
        Err(e) => {
            eprintln!("❌ Verification failed: {}", e);
            std::process::exit(1);
        }
    }
    
    println!("\n========================================");
    
    // Test the simple boolean verification function
    println!("\n🔍 Running simple verification...");
    let success = verify_proof_simple();
    
    if success {
        println!("🎉 All tests passed!");
        std::process::exit(0);
    } else {
        eprintln!("💥 Tests failed!");
        std::process::exit(1);
    }
}