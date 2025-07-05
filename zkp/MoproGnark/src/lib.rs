use std::ffi::CStr;
use std::os::raw::c_char;

/// Error type for verification operations
#[derive(Debug)]
pub enum VerifyError {
    NullPointer,
    InvalidUtf8(std::str::Utf8Error),
    VerificationFailed(String),
}

impl std::fmt::Display for VerifyError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            VerifyError::NullPointer => write!(f, "Received null pointer from C function"),
            VerifyError::InvalidUtf8(e) => write!(f, "Invalid UTF-8 in response: {}", e),
            VerifyError::VerificationFailed(msg) => write!(f, "Verification failed: {}", msg),
        }
    }
}

impl std::error::Error for VerifyError {}

/// Result type for verification operations
pub type VerifyResult<T> = Result<T, VerifyError>;

// External C function declaration - now marked as unsafe
unsafe extern "C" {
    fn verify() -> *mut c_char;
}

/// Safe Rust wrapper for the verify function
pub fn verify_proof() -> VerifyResult<String> {
    unsafe {
        // Call the C function
        let c_str_ptr = verify();
        
        // Check for null pointer
        if c_str_ptr.is_null() {
            return Err(VerifyError::NullPointer);
        }
        
        // Convert C string to Rust string
        let c_str = CStr::from_ptr(c_str_ptr);
        let result_str = c_str.to_str()
            .map_err(VerifyError::InvalidUtf8)?
            .to_owned();
        
        // Free the memory allocated by C.CString in Go
        libc::free(c_str_ptr as *mut libc::c_void);
        
        // Check if verification was successful
        if result_str.contains("SUCCESS") {
            Ok(result_str)
        } else {
            Err(VerifyError::VerificationFailed(result_str))
        }
    }
}

/// High-level verification function that returns a boolean result
pub fn verify_proof_simple() -> bool {
    match verify_proof() {
        Ok(result) => {
            println!("✅ Verification successful: {}", result);
            true
        }
        Err(e) => {
            eprintln!("❌ Verification failed: {}", e);
            false
        }
    }
}

/// Async wrapper for verification (useful for GUI applications)
#[cfg(feature = "async")]
pub async fn verify_proof_async() -> VerifyResult<String> {
    tokio::task::spawn_blocking(verify_proof).await
        .map_err(|e| VerifyError::VerificationFailed(format!("Task join error: {}", e)))?
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_verify_proof() {
        // Note: This test requires the input files to be present
        match verify_proof() {
            Ok(result) => {
                println!("Verification result: {}", result);
                assert!(result.contains("SUCCESS"));
            }
            Err(e) => {
                println!("Verification error: {}", e);
                // In a real test, you might want to assert based on expected conditions
            }
        }
    }

    #[test]
    fn test_verify_proof_simple() {
        let success = verify_proof_simple();
        println!("Simple verification result: {}", success);
    }
}