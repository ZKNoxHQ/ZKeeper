package main

import (
	"bytes" 
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"os"
	"time"

	"github.com/consensys/gnark-crypto/ecc"
	cryptoecdsa "github.com/consensys/gnark-crypto/ecc/secp256k1/ecdsa"
	"github.com/consensys/gnark/backend/groth16"
	groth16_bn254 "github.com/consensys/gnark/backend/groth16/bn254"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
	"github.com/consensys/gnark/std/algebra/emulated/sw_emulated"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/std/signature/ecdsa"
)

// EcdsaCircuit defines the circuit structure as provided by you.
type EcdsaCircuit[T, S emulated.FieldParams] struct {
	Sig ecdsa.Signature[S] `gnark:",secret"` // secret input
	Msg emulated.Element[S] `gnark:",public"` // Public input
	Pub ecdsa.PublicKey[T, S] `gnark:",public"` // Public input
}

func (c *EcdsaCircuit[T, S]) Define(api frontend.API) error {
	curveParams := sw_emulated.GetCurveParams[T]()
	c.Pub.Verify(api, curveParams, &c.Msg, &c.Sig)
	return nil
}

// ProveInputEcdsa struct for JSON serialization of witness inputs.
type ProveInputEcdsa struct {
	MsgHash string `json:"msgHash"` // Hex string of the message hash
	R       string `json:"r"`       // Hex string of signature R
	S       string `json:"s"`       // Hex string of signature S
	PubX    string `json:"pubX"`    // Hex string of public key X
	PubY    string `json:"pubY"`    // Hex string of public key Y
}

func main() {
	fmt.Println("--- Generating ECDSA circuit inputs and performing compliance check ---")

	// 1. Off-circuit ECDSA signature generation (to get inputs for the circuit)
	privKey, _ := cryptoecdsa.GenerateKey(rand.Reader)
	publicKey := privKey.PublicKey

	msg := []byte("testing ECDSA (pre-hashed)")
	
	sigBin, _ := privKey.Sign( msg, nil)

	// unmarshal signature
	var sig cryptoecdsa.Signature
	sig.SetBytes(sigBin)
	r, s := new(big.Int), new(big.Int)
	r.SetBytes(sig.R[:32])
	s.SetBytes(sig.S[:32])

	hash := cryptoecdsa.HashToInt(msg)
	
	// check that the signature is correct
	flag, _ := publicKey.Verify(sigBin, msg, nil)
	if !flag {
		fmt.Printf("Invalid signature\n")
	}

	// 2. Prepare JSON input for proving
    xBytes := publicKey.A.X.Bytes()
    yBytes := publicKey.A.Y.Bytes()

    proveInput := ProveInputEcdsa{
        MsgHash: hex.EncodeToString(hash.Bytes()),
        R:       hex.EncodeToString(r.Bytes()),
        S:       hex.EncodeToString(s.Bytes()),
        PubX:    hex.EncodeToString(xBytes[:]),
        PubY:    hex.EncodeToString(yBytes[:]),
    }

	proveInputJSON, err := json.MarshalIndent(proveInput, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling prove input JSON: %v\n", err)
		os.Exit(1)
	}

	// 3. Compile the circuit
	circuit := EcdsaCircuit[emulated.Secp256k1Fp, emulated.Secp256k1Fr]{}
	fmt.Printf("Compiling circuit...\n")
	ecdsaR1CS, err := frontend.Compile(ecc.BN254.ScalarField(), r1cs.NewBuilder, &circuit)
	if err != nil {
		fmt.Printf("Error compiling ECDSA circuit: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("BN254 circuit compiled with %d constraints\n", ecdsaR1CS.GetNbConstraints())

	// Debug circuit info
	fmt.Printf("Circuit public variables: %d\n", ecdsaR1CS.GetNbPublicVariables())
	fmt.Printf("Circuit secret variables: %d\n", ecdsaR1CS.GetNbSecretVariables())

	// 4. Perform Groth16 setup
	fmt.Printf("Starting Groth16 setup...\n")
	ecdsaPK, ecdsaVK, err := groth16.Setup(ecdsaR1CS)
	if err != nil {
		fmt.Printf("Error during Groth16 setup for ECDSA: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Setup done.\n")

	// 5. Create the full witness for the circuit (includes private and public parts)
	witnessCircuit := EcdsaCircuit[emulated.Secp256k1Fp, emulated.Secp256k1Fr]{
		Sig: ecdsa.Signature[emulated.Secp256k1Fr]{
			R: emulated.ValueOf[emulated.Secp256k1Fr](r),
			S: emulated.ValueOf[emulated.Secp256k1Fr](s),
		},
		Msg: emulated.ValueOf[emulated.Secp256k1Fr](hash),
		Pub: ecdsa.PublicKey[emulated.Secp256k1Fp, emulated.Secp256k1Fr]{
			X: emulated.ValueOf[emulated.Secp256k1Fp](publicKey.A.X),
			Y: emulated.ValueOf[emulated.Secp256k1Fp](publicKey.A.Y),
		},
	}
	witnessFull, err := frontend.NewWitness(&witnessCircuit, ecc.BN254.ScalarField())
	if err != nil {
		fmt.Printf("Error creating full witness: %v\n", err)
		os.Exit(1)
	}
	publicWitness, err := witnessFull.Public()
	if err != nil {
		fmt.Printf("Error getting public witness: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("witness creation done.\n")

	// 6. Write outputs to files
	writeToFile("r1cs.bin", ecdsaR1CS)
	writeToFile("proving_key.bin", ecdsaPK)
	writeToFile("verifying_key.bin", ecdsaVK)
	writeToFile("witness_input.json", bytes.NewReader(proveInputJSON))

	// 7. Perform a compliance check: Prove and Verify
	fmt.Println("\n--- Performing compliance check (Prove & Verify within generate_input.go) ---")

	// Prove
	startProve := time.Now()
	proof, err := groth16.Prove(ecdsaR1CS, ecdsaPK, witnessFull)
	if err != nil {
		fmt.Printf("Compliance check: Error generating proof: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Compliance check: Proof generated (%.1fms).\n", float64(time.Since(startProve).Milliseconds()))

	// Verify
	startVerify := time.Now()
	err = groth16.Verify(proof, ecdsaVK, publicWitness)
	if err != nil {
		fmt.Printf("Compliance check: Verification FAILED: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Compliance check: Verification SUCCEEDED (%.1fms)!\n", float64(time.Since(startVerify).Milliseconds()))
	fmt.Println("Compliance check PASSED. Generated inputs are valid.")

	// 8. Export the Solidity verifier
	fmt.Println("\n--- Exporting Solidity Verifier ---")
	verifierFile, err := os.Create("verifier.sol")
	if err != nil {
		fmt.Printf("Error creating verifier.sol: %v\n", err)
		os.Exit(1)
	}
	defer verifierFile.Close()
	err = ecdsaVK.ExportSolidity(verifierFile)
	if err != nil {
		fmt.Printf("Error exporting solidity verifier: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully exported verifier.sol")

	// 9. Generate Solidity calldata
	generateSolidityCalldata(proof, publicWitness)
}

func generateSolidityCalldata(proof groth16.Proof, publicWitness frontend.Witness) {
	fmt.Println("\n--- Generating Solidity Calldata ---")

	Proof := proof.(*groth16_bn254.Proof)
	a := Proof.Ar
	b := Proof.Bs
	c := Proof.Krs

	// Solidity expects proof = [A.X, A.Y, B.X.A0, B.X.A1, B.Y.A0, B.Y.A1, C.X, C.Y]
	proofVals := []*big.Int{
		a.X.BigInt(new(big.Int)), a.Y.BigInt(new(big.Int)),
		b.X.A0.BigInt(new(big.Int)), b.X.A1.BigInt(new(big.Int)),
		b.Y.A0.BigInt(new(big.Int)), b.Y.A1.BigInt(new(big.Int)),
		c.X.BigInt(new(big.Int)), c.Y.BigInt(new(big.Int)),
	}

	fmt.Println("uint256[8] memory proof = [")
	for i, val := range proofVals {
		comma := ","
		if i == len(proofVals)-1 {
			comma = ""
		}
		fmt.Printf("    %s%s\n", val.Text(10), comma)
	}
	fmt.Println("];")

	// Handle commitments if they exist
	if len(Proof.Commitments) > 0 {
		fmt.Println("\n// Commitments found:")
		comm := Proof.Commitments[0]
		commX := comm.X.BigInt(new(big.Int))
		commY := comm.Y.BigInt(new(big.Int))
		fmt.Printf("uint256[2] memory commitments = [%s, %s];\n", commX.Text(10), commY.Text(10))

		pok := Proof.CommitmentPok
		pokX := pok.X.BigInt(new(big.Int))
		pokY := pok.Y.BigInt(new(big.Int))
		fmt.Printf("uint256[2] memory commitmentPok = [%s, %s];\n", pokX.Text(10), pokY.Text(10))
	} else {
		fmt.Println("\n// No commitments - using dummy values:")
		fmt.Println("uint256[2] memory commitments = [0, 0];")
		fmt.Println("uint256[2] memory commitmentPok = [0, 0];")
	}

	// Get public inputs - this will tell us the actual format
	fmt.Println("\n// Public inputs:")
	publicVector := publicWitness.Vector()
	fmt.Printf("// Number of public inputs: %d\n", len(publicVector))
	
	// Print all public inputs
	fmt.Printf("uint256[%d] memory input = [\n", len(publicVector))
	for i := 0; i < len(publicVector); i++ {
		comma := ","
		if i == len(publicVector)-1 {
			comma = ""
		}
		fmt.Printf("    %v%s\n", publicVector[i], comma)
	}
	fmt.Println("];")

	// Generate the Solidity test function
	fmt.Println("\n--- Solidity Test Function ---")
	fmt.Println("function test_EcdsaZKP() public {")
	
	fmt.Println("    uint256[8] memory proof = [")
	for i, val := range proofVals {
		comma := ","
		if i == len(proofVals)-1 {
			comma = ""
		}
		fmt.Printf("        %s%s\n", val.Text(10), comma)
	}
	fmt.Println("    ];")
	
	if len(Proof.Commitments) > 0 {
		comm := Proof.Commitments[0]
		commX := comm.X.BigInt(new(big.Int))
		commY := comm.Y.BigInt(new(big.Int))
		fmt.Printf("    uint256[2] memory commitments = [%s, %s];\n", commX.Text(10), commY.Text(10))

		pok := Proof.CommitmentPok
		pokX := pok.X.BigInt(new(big.Int))
		pokY := pok.Y.BigInt(new(big.Int))
		fmt.Printf("    uint256[2] memory commitmentPok = [%s, %s];\n", pokX.Text(10), pokY.Text(10))
	} else {
		fmt.Println("    uint256[2] memory commitments = [0, 0];")
		fmt.Println("    uint256[2] memory commitmentPok = [0, 0];")
	}
	
	fmt.Printf("    uint256[%d] memory input = [\n", len(publicVector))
	for i := 0; i < len(publicVector); i++ {
		comma := ","
		if i == len(publicVector)-1 {
			comma = ""
		}
		fmt.Printf("        %v%s\n", publicVector[i], comma)
	}
	fmt.Println("    ];")
	
	fmt.Println("    verifier.verifyProof(proof, commitments, commitmentPok, input);")
	fmt.Println("    console.log(\"ECDSA ZK proof verification successful!\");")
	fmt.Println("}")
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