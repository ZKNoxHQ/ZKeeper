package main

import (
	"bytes"
	"io"
	"time"

	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"os"

	// Added for performance timing
	"github.com/consensys/gnark-crypto/ecc"

	// cryptoposeidon2 "github.com/consensys/gnark-crypto/ecc/bn254/fr/poseidon2"

	"github.com/ethereum/go-ethereum/common/hexutil"

	"github.com/consensys/gnark/backend/plonk"
	plonk_bn254 "github.com/consensys/gnark/backend/plonk/bn254"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_emulated"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/std/signature/ecdsa"
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

// ProveInputEcdsa struct for JSON serialization of witness inputs.
type ProveInputEcdsa struct {
	MsgHash string `json:"msgHash"` // Hex string of the message hash
	R       string `json:"r"`       // Hex string of signature R
	S       string `json:"s"`       // Hex string of signature S
	PubX    string `json:"pubX"`    // Hex string of public key X
	PubY    string `json:"pubY"`    // Hex string of public key Y
	Address string `json:"address"` // Hex string of address
	Nonce   string `json:"nonce"`   // Hex string of nonce
	Com     string `json:"com"`     // Hex string of Com
}

func main() {

	// 8. Test the ReadFromFile functionality
	// 1. Read back the compiled circuit
	loadedR1CS := plonk.NewCS(ecc.BN254)
	err := readFromFile("r1cs.bin", loadedR1CS)
	fmt.Printf("Read r1cs.bin (Constraints: %d)\n", loadedR1CS.GetNbConstraints())

	// 2. Read back the proving key
	loadedPK := plonk.NewProvingKey(ecc.BN254)
	err = readFromFile("proving_key.bin", loadedPK)
	fmt.Println("Read proving_key.bin")

	// 3. Read back the verifying key
	loadedVK := plonk.NewVerifyingKey(ecc.BN254)
	err = readFromFile("verifying_key.bin", loadedVK)
	fmt.Println("Read verifying_key.bin")

	// 4. Read back the prove input JSON
	var loadedProveInput ProveInputEcdsa
	err = readFromFile("witness_input.json", &loadedProveInput)
	fmt.Println("Read witness_input.json")

	// Decode hex strings back to big.Int and byte slices for witness construction
	rBytes, err := hex.DecodeString(loadedProveInput.R)
	sBytes, err := hex.DecodeString(loadedProveInput.S)
	if err != nil {
		fmt.Printf("Error decoding S hex: %v\n", err)
		os.Exit(1)
	}
	msgHashBytes, err := hex.DecodeString(loadedProveInput.MsgHash)
	pubXBytes, err := hex.DecodeString(loadedProveInput.PubX)
	pubYBytes, err := hex.DecodeString(loadedProveInput.PubY)
	addressBytes, err := hex.DecodeString(loadedProveInput.Address)
	nonceBytes, err := hex.DecodeString(loadedProveInput.Nonce)
	comBytes, err := hex.DecodeString(loadedProveInput.Com)

	rLoaded := new(big.Int).SetBytes(rBytes)
	sLoaded := new(big.Int).SetBytes(sBytes)
	pubXLoaded := new(big.Int).SetBytes(pubXBytes)
	pubYLoaded := new(big.Int).SetBytes(pubYBytes)
	addressLoaded := new(big.Int).SetBytes(addressBytes)
	nonceLoaded := new(big.Int).SetBytes(nonceBytes)
	comLoaded := new(big.Int).SetBytes(comBytes)

	// 5. Create a new witness using the loaded input data
	witnessCircuitLoaded := Circuit[emulated.Secp256k1Fp, emulated.Secp256k1Fr]{
		Sig: ecdsa.Signature[emulated.Secp256k1Fr]{
			R: emulated.ValueOf[emulated.Secp256k1Fr](rLoaded),
			S: emulated.ValueOf[emulated.Secp256k1Fr](sLoaded),
		},
		Msg: emulated.ValueOf[emulated.Secp256k1Fr](msgHashBytes),
		Pub: ecdsa.PublicKey[emulated.Secp256k1Fp, emulated.Secp256k1Fr]{
			X: emulated.ValueOf[emulated.Secp256k1Fp](pubXLoaded),
			Y: emulated.ValueOf[emulated.Secp256k1Fp](pubYLoaded),
		},
		Address: addressLoaded,
		Nonce:   nonceLoaded,
		Com:     comLoaded,
	}
	witnessFullLoaded, err := frontend.NewWitness(&witnessCircuitLoaded, ecc.BN254.ScalarField())
	publicWitnessLoaded, err := witnessFullLoaded.Public()

	// 6. Perform a new proof and verification using the loaded artifacts
	fmt.Println("\n--- Proving with loaded setup ---")

	// Prove
	startProveLoaded := time.Now()
	proofLoaded, err := plonk.Prove(loadedR1CS, loadedPK, witnessFullLoaded)
	fmt.Printf("Proof GENERATED (%.1fms).\n", float64(time.Since(startProveLoaded).Milliseconds()))

	// Verify
	// err = plonk.Verify(proofLoaded, loadedVK, publicWitnessLoaded)

	// 9. Export the Solidity verifier test
	fmt.Println("\n--- Exporting Solidity Verifier Test ---")
	verifierTestFile, err := os.Create("solidity/test/Verifier.t.sol")
	defer verifierTestFile.Close()

	// header
	verifierTestFile.Write([]byte(`// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.25;

import {Test, console} from "forge-std/Test.sol";
import {PlonkVerifier} from "../src/Verifier.sol";

contract VerifierTest is Test {
    PlonkVerifier ZkK1;

    function setUp() public {
        ZkK1 = new PlonkVerifier();
    }

    function test_k1Plonk() public view {
`))

	Proof := proofLoaded.(*plonk_bn254.Proof)
	verifierTestFile.Write([]byte(`bytes memory proof = hex"` + hexutil.Encode(Proof.MarshalSolidity())[2:] + `";`))
	verifierTestFile.Write([]byte("\n"))

	PI := fmt.Sprintf("%v", publicWitnessLoaded.Vector())

	verifierTestFile.Write([]byte("uint256[5] memory public_inputs = " + PI + ";\n"))

	// footer
	verifierTestFile.Write([]byte(`
        uint256[] memory inputs = new uint256[](5);
        for (uint i = 0; i < 5; i++) inputs[i] = uint256(public_inputs[i]);

        bool res = ZkK1.Verify(proof, inputs);
        assertTrue(res);
        console.log(res);
    }
}
`))
	fmt.Println("Successfully exported solidty/test/Verifier.t.sol")

	fmt.Print("\n\n\n=======================\nPROOF and PUBLIC INPUTS\n=======================\n0x", hexutil.Encode(Proof.MarshalSolidity())[2:], " \"", publicWitnessLoaded.Vector(), "\"\n")
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

// readFromFile is a helper to deserialize and read gnark objects or JSON from files.
func readFromFile(filename string, data interface{}) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("error opening file %s: %w", filename, err)
	}
	defer file.Close()

	switch v := data.(type) {
	case io.ReaderFrom:
		_, err = v.ReadFrom(file)
		if err != nil && err != io.EOF { // io.EOF is expected if the file is empty or partially read
			return fmt.Errorf("error reading from file %s into io.ReaderFrom: %w", filename, err)
		}
	case *ProveInputEcdsa: // For the JSON input
		decoder := json.NewDecoder(file)
		err = decoder.Decode(v)
		if err != nil {
			return fmt.Errorf("error decoding JSON from file %s: %w", filename, err)
		}
	default:
		return fmt.Errorf("unsupported type for reading from file: %T", data)
	}

	return nil
}
