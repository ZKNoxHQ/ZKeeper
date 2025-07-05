package main

import (
	"bytes"
	"io"

	"fmt"
	"os"

	// Added for performance timing
	"github.com/consensys/gnark-crypto/ecc"

	// cryptoposeidon2 "github.com/consensys/gnark-crypto/ecc/bn254/fr/poseidon2"

	"github.com/consensys/gnark/backend/plonk"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/std/algebra/emulated/sw_emulated"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/std/signature/ecdsa"

	"github.com/consensys/gnark/test/unsafekzg"
)

// Circuit defines the circuit structure as provided by you.
type Circuit[T, S emulated.FieldParams] struct {
	Sig     ecdsa.Signature[S]    `gnark:",secret"` // signature
	Msg     emulated.Element[S]   `gnark:",public"` // message
	Pub     ecdsa.PublicKey[T, S] `gnark:",secret"` // now secret
	Address frontend.Variable     `gnark:",secret"` // secret address
	Nonce   frontend.Variable     `gnark:",secret"` // secret nonce
	Com     frontend.Variable     `gnark:",public"` // public commitment
}

func (c *Circuit[T, S]) Define(api frontend.API) error {
	curveParams := sw_emulated.GetCurveParams[T]()
	c.Pub.Verify(api, curveParams, &c.Msg, &c.Sig)

	mimc, _ := mimc.NewMiMC(api)

	// specify constraints
	// mimc(preImage) == hash
	mimc.Write(c.Address)
	mimc.Write(c.Nonce)
	api.AssertIsEqual(c.Com, mimc.Sum())
	return nil
}

func main() {
	fmt.Println("--- Generating ECDSA circuit inputs and performing compliance check ---")

	// 1. Compile the circuit
	circuit := Circuit[emulated.Secp256k1Fp, emulated.Secp256k1Fr]{}
	fmt.Printf("Compiling circuit...\n")
	R1CS, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, &circuit)
	if err != nil {
		fmt.Printf("Error compiling ECDSA circuit: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("BN254 circuit compiled with %d constraints",
		R1CS.GetNbConstraints())

	// 2. Perform Groth16 setup
	fmt.Printf("Starting Plonk setup...\n")
	A, B, _ := unsafekzg.NewSRS(R1CS)
	PK, VK, err := plonk.Setup(R1CS, A, B)
	if err != nil {
		fmt.Printf("Error during Plonk setup for ECDSA: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Setup done.\n")

	// 3. Save to bin files
	writeToFile("r1cs.bin", R1CS)
	writeToFile("proving_key.bin", PK)
	writeToFile("verifying_key.bin", VK)

	fmt.Println("\nAll input files generated successfully for CGO wrapper.")

	// 4. Export the Solidity verifier contract
	fmt.Println("\n--- Exporting Solidity Verifier ---")
	verifierFile, err := os.Create("solidity/src/Verifier.sol")
	if err != nil {
		fmt.Printf("Error creating solidity/src/Verifier.sol: %v\n", err)
		os.Exit(1)
	}
	defer verifierFile.Close()
	err = VK.ExportSolidity(verifierFile)
	if err != nil {
		fmt.Printf("Error exporting solidity verifier: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully exported solidty/src/Verifier.sol")

}

// writeToFile is a helper to serialize and write gnark objects or byte readers to files.
func writeToFile(filename string, data interface{}) {
	file, err := os.Create(filename)
	if err != nil {
		fmt.Printf("Error creating file %s: %v\n", filename, err)
		os.Exit(1)
	}
	defer file.Close()

	switch v := data.(type) {
	case io.WriterTo:
		_, err = v.WriteTo(file)
	case *bytes.Reader: // For the JSON input
		_, err = v.WriteTo(file)
	default:
		err = fmt.Errorf("unsupported type for writing to file")
	}

	if err != nil {
		fmt.Printf("Error writing to file %s: %v\n", filename, err)
		os.Exit(1)
	}
	fmt.Printf("Wrote %s\n", filename)
}
