package main

import (
	// "bytes"
	// "crypto/rand"
	// "encoding/hex"
	// "encoding/json"
	// "fmt"
	// "math/big"
	// "os"
	// "time"

	// "github.com/consensys/gnark-crypto/ecc"
	// cryptomimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
	// "github.com/consensys/gnark/backend/plonk"
	// "github.com/consensys/gnark/frontend"
	// "github.com/consensys/gnark/std/math/emulated"
	// "github.com/consensys/gnark/std/signature/ecdsa"

	"fmt"
	"log"
	"os"
	"time"

	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"

	"zkp/circuit"
	"zkp/utils"

	"github.com/consensys/gnark-crypto/ecc"
)

func main() {

	witnessCircuit, err := circuit.GenerateWitness()
	if err != nil {
		log.Fatalf("Failed to generate the witness: %v", err)
	}

	R1CS, PK, VK, err := circuit.ComputeTrustedSetup()

	witnessFull, err := frontend.NewWitness(&witnessCircuit, ecc.BN254.ScalarField())
	if err != nil {
		log.Fatalf("failed to create witness: %v", err)
	}

	publicWitness, err := witnessFull.Public()
	if err != nil {
		log.Fatalf("failed to get public witness: %v", err)
	}
	// 6. Perform a new proof and verification using the loaded artifacts
	fmt.Println("\n--- Proving with loaded setup ---")

	// Prove
	startProve := time.Now()
	proof, err := plonk.Prove(R1CS, PK, witnessFull)
	fmt.Printf("Proof GENERATED (%.1fms).\n", float64(time.Since(startProve).Milliseconds()))

	// Verify
	err = plonk.Verify(proof, VK, publicWitness)
	if err != nil {
		fmt.Printf("Error verifying the proof: %v\n", err)
		os.Exit(1)
	}

	utils.WriteSolidityTestFile(proof, publicWitness)

}
