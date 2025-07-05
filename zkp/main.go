package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"time"

	cryptoecdsa "github.com/consensys/gnark-crypto/ecc/secp256k1/ecdsa"
	"github.com/ethereum/go-ethereum/common/hexutil"

	// Added for performance timing
	"github.com/consensys/gnark-crypto/ecc"

	"github.com/consensys/gnark-crypto/ecc/secp256k1/fp"

	"github.com/consensys/gnark/backend/plonk"
	plonk_bn254 "github.com/consensys/gnark/backend/plonk/bn254"
	"github.com/consensys/gnark/backend/witness"
	"github.com/consensys/gnark/constraint"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/scs"
	"github.com/consensys/gnark/std/algebra/emulated/sw_emulated"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/std/signature/ecdsa"
	"github.com/consensys/gnark/test/unsafekzg"
)

// EcdsaCircuit defines the circuit structure as provided by you.
type EcdsaCircuit[T, S emulated.FieldParams] struct {
	Sig ecdsa.Signature[S]    `gnark:",secret"` // secret input
	Msg emulated.Element[S]   `gnark:",public"` // Public input
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

func circuit_precomputation() (constraint.ConstraintSystem, plonk.ProvingKey, plonk.VerifyingKey) {
	// Compile the circuit and perform a SRS
	circuit := EcdsaCircuit[emulated.Secp256k1Fp, emulated.Secp256k1Fr]{}
	fmt.Printf("Compiling circuit...\n")
	r1cs, err := frontend.Compile(ecc.BN254.ScalarField(), scs.NewBuilder, &circuit)
	if err != nil {
		fmt.Printf("Error compiling ECDSA circuit: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("BN254 circuit compiled with %d constraints",
		r1cs.GetNbConstraints())

	// 4. Perform Groth16 setup
	fmt.Printf("Starting Groth16 setup...\n")

	A, B, _ := unsafekzg.NewSRS(r1cs)
	PK, VK, err := plonk.Setup(r1cs, A, B)
	if err != nil {
		fmt.Printf("Error during Groth16 setup for ECDSA: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Setup done.\n")

	// 6. Write outputs to files
	writeToFile("data/r1cs.bin", r1cs)
	writeToFile("data/proving_key.bin", PK)
	writeToFile("data/verifying_key.bin", VK)

	return r1cs, PK, VK
}

func loading_circuit_precomputation() (constraint.ConstraintSystem, plonk.ProvingKey, plonk.VerifyingKey) {
	// 1. Read back the compiled circuit
	cs := plonk.NewCS(ecc.BN254)
	err := readFromFile("data/r1cs.bin", cs)
	if err != nil {
		fmt.Printf("Error reading r1cs.bin: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Read r1cs.bin (Constraints: %d)\n", cs.GetNbConstraints())

	// 2. Read back the proving key
	PK := plonk.NewProvingKey(ecc.BN254)
	err = readFromFile("data/proving_key.bin", PK)
	if err != nil {
		fmt.Printf("Error reading proving_key.bin: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Read proving_key.bin")

	// 3. Read back the verifying key
	VK := plonk.NewVerifyingKey(ecc.BN254)
	err = readFromFile("data/verifying_key.bin", VK)
	if err != nil {
		fmt.Printf("Error reading verifying_key.bin: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Read verifying_key.bin")

	return cs, PK, VK
}

func circuit_inputs() (*big.Int, *big.Int, *big.Int, fp.Element, fp.Element) {
	fmt.Println("--- Generating ECDSA circuit inputs and performing compliance check ---")

	// 1. Off-circuit ECDSA signature generation (to get inputs for the circuit)
	privKey, _ := cryptoecdsa.GenerateKey(rand.Reader)
	publicKey := privKey.PublicKey

	msg := []byte("testing ECDSA (pre-hashed)")
	sigBin, _ := privKey.Sign(msg, nil)

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
	pkx_bytes := publicKey.A.X.Bytes()
	pky_bytes := publicKey.A.Y.Bytes()
	// Store the byte arrays in temporary variables first
	proveInput := ProveInputEcdsa{
		MsgHash: hex.EncodeToString(hash.Bytes()),
		R:       hex.EncodeToString(r.Bytes()),
		S:       hex.EncodeToString(s.Bytes()),
		PubX:    hex.EncodeToString(pkx_bytes[:]),
		PubY:    hex.EncodeToString(pky_bytes[:]),
	}

	proveInputJSON, err := json.MarshalIndent(proveInput, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling prove input JSON: %v\n", err)
		os.Exit(1)
	}
	writeToFile("data/witness_input.json", bytes.NewReader(proveInputJSON))

	return hash, r, s, publicKey.A.X, publicKey.A.Y

}

func generate_witness() (witness.Witness, witness.Witness) {
	hash, r, s, pkx, pky := circuit_inputs()
	// 5. Create the full witness for the circuit (includes private and public parts)
	witnessCircuit := EcdsaCircuit[emulated.Secp256k1Fp, emulated.Secp256k1Fr]{
		Sig: ecdsa.Signature[emulated.Secp256k1Fr]{
			R: emulated.ValueOf[emulated.Secp256k1Fr](r),
			S: emulated.ValueOf[emulated.Secp256k1Fr](s),
		},
		Msg: emulated.ValueOf[emulated.Secp256k1Fr](hash),
		Pub: ecdsa.PublicKey[emulated.Secp256k1Fp, emulated.Secp256k1Fr]{
			X: emulated.ValueOf[emulated.Secp256k1Fp](pkx),
			Y: emulated.ValueOf[emulated.Secp256k1Fp](pky),
		},
	}

	witnessFull, err := frontend.NewWitness(&witnessCircuit, ecc.BN254.ScalarField())
	if err != nil {
		fmt.Printf("Error creating full witness: %v\n", err)
		os.Exit(1)
	}
	publicWitness, err := witnessFull.Public() // Extract public parts for verification
	if err != nil {
		fmt.Printf("Error getting public witness: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("witness creation done.\n")

	return witnessFull, publicWitness
}

func load_witness(r1cs constraint.ConstraintSystem) (witness.Witness, witness.Witness) {

	// 4. Read back the prove input JSON
	var loadedProveInput ProveInputEcdsa
	err := readFromFile("data/witness_input.json", &loadedProveInput)
	if err != nil {
		fmt.Printf("Error reading witness_input.json: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Read witness_input.json")

	// Decode hex strings back to big.Int and byte slices for witness construction
	rBytes, err := hex.DecodeString(loadedProveInput.R)
	if err != nil {
		fmt.Printf("Error decoding R hex: %v\n", err)
		os.Exit(1)
	}
	sBytes, err := hex.DecodeString(loadedProveInput.S)
	if err != nil {
		fmt.Printf("Error decoding S hex: %v\n", err)
		os.Exit(1)
	}
	msgHashBytes, err := hex.DecodeString(loadedProveInput.MsgHash)
	if err != nil {
		fmt.Printf("Error decoding MsgHash hex: %v\n", err)
		os.Exit(1)
	}
	pubXBytes, err := hex.DecodeString(loadedProveInput.PubX)
	if err != nil {
		fmt.Printf("Error decoding PubX hex: %v\n", err)
		os.Exit(1)
	}
	pubYBytes, err := hex.DecodeString(loadedProveInput.PubY)
	if err != nil {
		fmt.Printf("Error decoding PubY hex: %v\n", err)
		os.Exit(1)
	}

	rLoaded := new(big.Int).SetBytes(rBytes)
	sLoaded := new(big.Int).SetBytes(sBytes)
	pubXLoaded := new(big.Int).SetBytes(pubXBytes)
	pubYLoaded := new(big.Int).SetBytes(pubYBytes)

	// 5. Create a new witness using the loaded input data
	witnessCircuitLoaded := EcdsaCircuit[emulated.Secp256k1Fp, emulated.Secp256k1Fr]{
		Sig: ecdsa.Signature[emulated.Secp256k1Fr]{
			R: emulated.ValueOf[emulated.Secp256k1Fr](rLoaded),
			S: emulated.ValueOf[emulated.Secp256k1Fr](sLoaded),
		},
		Msg: emulated.ValueOf[emulated.Secp256k1Fr](msgHashBytes),
		Pub: ecdsa.PublicKey[emulated.Secp256k1Fp, emulated.Secp256k1Fr]{
			X: emulated.ValueOf[emulated.Secp256k1Fp](pubXLoaded),
			Y: emulated.ValueOf[emulated.Secp256k1Fp](pubYLoaded),
		},
	}
	witnessFull, err := frontend.NewWitness(&witnessCircuitLoaded, ecc.BN254.ScalarField())
	if err != nil {
		fmt.Printf("Error creating full witness from loaded data: %v\n", err)
		os.Exit(1)
	}
	publicWitness, err := witnessFull.Public()
	if err != nil {
		fmt.Printf("Error getting public witness from loaded data: %v\n", err)
		os.Exit(1)
	}
	return witnessFull, publicWitness
}

func prove_and_verify(
	r1cs constraint.ConstraintSystem,
	PK plonk.ProvingKey,
	VK plonk.VerifyingKey,
	witnessFull witness.Witness,
	publicWitness witness.Witness,
) plonk.Proof {
	// 7. Perform a compliance check: Prove and Verify
	fmt.Println("\n--- Performing compliance check (Prove & Verify within generate_input.go) ---")

	// Prove
	startProve := time.Now()
	proof, err := plonk.Prove(r1cs, PK, witnessFull)
	if err != nil {
		fmt.Printf("Compliance check: Error generating proof: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Compliance check: Proof generated (%.1fms).\n", float64(time.Since(startProve).Milliseconds()))

	// Verify
	startVerify := time.Now()
	err = plonk.Verify(proof, VK, publicWitness)
	if err != nil {
		fmt.Printf("Compliance check: Verification FAILED: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Compliance check: Verification SUCCEEDED (%.1fms)!\n", float64(time.Since(startVerify).Milliseconds()))
	fmt.Println("Compliance check PASSED. Generated inputs are valid.")
	return proof
}

func export_to_solidity(VK plonk.VerifyingKey, proof plonk.Proof, publicWitness witness.Witness) {
	// 8. Export the Solidity verifier contract
	fmt.Println("\n--- Exporting Solidity Verifier ---")
	verifierFile, err := os.Create("src/Verifier.sol")
	if err != nil {
		fmt.Printf("Error creating Verifier.sol: %v\n", err)
		os.Exit(1)
	}
	defer verifierFile.Close()
	err = VK.ExportSolidity(verifierFile)
	if err != nil {
		fmt.Printf("Error exporting solidity verifier: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Successfully exported verifier.sol")

	// Proof
	fmt.Println(hexutil.Encode(proof.(*plonk_bn254.Proof).MarshalSolidity()))
	// Public inputs
	fmt.Println(publicWitness.Vector())
}

func main() {

	// Using precomputations
	// r1cs, PK, VK := loading_circuit_precomputation()
	// witnessFull, publicWitness := load_witness(r1cs)

	// Without precomputations
	r1cs, PK, VK := circuit_precomputation()
	witnessFull, publicWitness := generate_witness()

	// Proof and verify
	proof := prove_and_verify(r1cs, PK, VK, witnessFull, publicWitness)

	// Export to Solidity
	export_to_solidity(VK, proof, publicWitness)

}
