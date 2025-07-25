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
	"zkp/circuit"
	"zkp/utils"

	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/constraint"
	"github.com/consensys/gnark/frontend"

	"github.com/consensys/gnark-crypto/ecc"
)

func main() {

	withSetup := false

	for _, arg := range os.Args[1:] {
		if arg == "setup" {
			withSetup = true
			break
		}
	}

	var R1CS constraint.ConstraintSystem
	var PK plonk.ProvingKey
	var VK plonk.VerifyingKey
	var err error

	if withSetup {
		// Trusted setup
		R1CS, PK, VK, err = circuit.ComputeTrustedSetup()
		if err != nil {
			panic(err)
		}

	} else {
		// Proof generation
		R1CS, PK, VK, err = circuit.LoadTrustedSetup()
		witnessCircuit, err := circuit.GenerateWitness()
		if err != nil {
			log.Fatalf("Failed to generate the witness: %v", err)
		}
		if err != nil {
			panic(err)
		}

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
}
