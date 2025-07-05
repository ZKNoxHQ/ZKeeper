#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "verify.h"

int main() {
    printf("Starting verification test...\n");
    printf("========================================\n");
    
    // Call the verify function from the Go shared library
    char* result = verify();
    
    if (result == NULL) {
        printf("ERROR: verify() returned NULL\n");
        return 1;
    }
    
    printf("Result from Go verify function:\n");
    printf("========================================\n");
    printf("%s\n", result);
    printf("========================================\n");
    
    // Check if verification was successful
    if (strstr(result, "SUCCESS") != NULL) {
        printf("✅ Verification completed successfully!\n");
        free(result); // Free the memory allocated by C.CString in Go
        return 0;
    } else {
        printf("❌ Verification failed or encountered errors.\n");
        free(result); // Free the memory allocated by C.CString in Go
        return 1;
    }
}