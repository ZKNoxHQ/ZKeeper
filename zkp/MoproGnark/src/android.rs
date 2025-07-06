use jni::objects::{JClass, JString};
use jni::sys::jstring;
use jni::JNIEnv;

use crate::{verify_proof, VerifyError};

/// JNI function called from Android/Java
/// Performs ZK proof verification and returns the result as a string
#[no_mangle]
pub extern "system" fn Java_com_zkverify_MainActivity_verifyProof(
    mut env: JNIEnv,
    _class: JClass,
) -> jstring {
    // Perform the verification
    let result = match verify_proof() {
        Ok(result_msg) => {
            format!("SUCCESS: {}", result_msg)
        }
        Err(e) => {
            let error_msg = match e {
                VerifyError::NullPointer => "Critical error: Received null pointer from verification".to_string(),
                VerifyError::InvalidUtf8(err) => format!("Error: Invalid response format: {}", err),
                VerifyError::VerificationFailed(msg) => format!("Verification failed: {}", msg),
            };
            format!("ERROR: {}", error_msg)
        }
    };

    // Convert Rust string to Java string
    let output = env
        .new_string(&result)
        .expect("Couldn't create Java string!");

    output.into_raw()
}

/// JNI function to test connectivity
#[no_mangle]
pub extern "system" fn Java_com_zkverify_MainActivity_testConnection(
    mut env: JNIEnv,
    _class: JClass,
) -> jstring {
    let result = "🦀 Rust library connection successful!\n✅ JNI bridge working\n✅ Ready for ZK verification";
    let output = env
        .new_string(result)
        .expect("Couldn't create Java string!");

    output.into_raw()
}