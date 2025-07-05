use std::ffi::CStr;
use std::os::raw::c_char;

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

pub type VerifyResult<T> = Result<T, VerifyError>;

unsafe extern "C" {
    fn verify() -> *mut c_char;
}

pub fn verify_proof() -> VerifyResult<String> {
    unsafe {
        let c_str_ptr = verify();
        
        if c_str_ptr.is_null() {
            return Err(VerifyError::NullPointer);
        }
        
        let c_str = CStr::from_ptr(c_str_ptr);
        let result_str = c_str.to_str()
            .map_err(VerifyError::InvalidUtf8)?
            .to_owned();
        
        libc::free(c_str_ptr as *mut libc::c_void);
        
        if result_str.contains("SUCCESS") {
            Ok(result_str)
        } else {
            Err(VerifyError::VerificationFailed(result_str))
        }
    }
}

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