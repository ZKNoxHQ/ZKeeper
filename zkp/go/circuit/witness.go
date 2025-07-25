package circuit

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"zkp/utils"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/std/signature/ecdsa"
)

func GenerateWitness() (VerifCircuit[emulated.Secp256k1Fp, emulated.Secp256k1Fr], error) {
	// Load the signature, public key and message hash from the json file
	// Generate a nonce and the corresponding public key commitment
	// Save it to a file

	EmptyCircuit := VerifCircuit[emulated.Secp256k1Fp, emulated.Secp256k1Fr]{}

	// Read the JSON file
	var TransactionInput TransactionInput
	rBytes, sBytes, msgHashBytes, pubXBytes, pubYBytes, err := ReadTransactionFromFile("input/transaction_input.json", &TransactionInput)
	if err != nil {
		return EmptyCircuit, fmt.Errorf("Error while reading input/transaction_input.json")
	}

	// Nonce (160 bits)
	nonceBytes := make([]byte, 20)
	_, err = rand.Read(nonceBytes)
	if err != nil {
		return EmptyCircuit, fmt.Errorf("Error while creating the nonce: %w\n", err)
	}

	// compute the hash of the public key
	h := sha256.New()
	h.Write(pubXBytes)
	h.Write(pubYBytes)
	h.Write(nonceBytes)
	pubHash := h.Sum(nil)
	pubHashLo := pubHash[:16]
	pubHashHi := pubHash[16:32]

	// Save to a file
	Output := ProofWitness{
		MsgHash: hex.EncodeToString(msgHashBytes),
		R:       hex.EncodeToString(rBytes),
		S:       hex.EncodeToString(sBytes),
		PubX:    hex.EncodeToString(pubXBytes[:]),
		PubY:    hex.EncodeToString(pubYBytes[:]),
		Nonce:   hex.EncodeToString(nonceBytes[:]),
		Com:     hex.EncodeToString(pubHash[:]),
	}

	OutputJSON, err := json.MarshalIndent(Output, "", "  ")
	if err != nil {
		return EmptyCircuit, fmt.Errorf("Error marhsaling prove input JSON: %v\n", err)
	}

	utils.WriteToFile("output/witness.json", bytes.NewReader(OutputJSON))

	// now we prepare the witness for the circuit
	witnessCircuit := VerifCircuit[emulated.Secp256k1Fp, emulated.Secp256k1Fr]{
		// we splitted the public key hash into two 16-byte variables to fit into BN254 field
		PublicKeyHash: [2]frontend.Variable{pubHashLo, pubHashHi},
		// we construct the public key as non-native element. NB! this means that both X and Y coordinates are 4 limbs of 64 bytes each, so 8 limbs total
		PublicKey: ecdsa.PublicKey[emulated.Secp256k1Fp, emulated.Secp256k1Fr]{
			X: emulated.ValueOf[emulated.Secp256k1Fp](pubXBytes),
			Y: emulated.ValueOf[emulated.Secp256k1Fp](pubYBytes),
		},
		Signature: ecdsa.Signature[emulated.Secp256k1Fr]{
			R: emulated.ValueOf[emulated.Secp256k1Fr](rBytes),
			S: emulated.ValueOf[emulated.Secp256k1Fr](sBytes),
		},
		Msg:   emulated.ValueOf[emulated.Secp256k1Fr](msgHashBytes),
		Nonce: frontend.Variable(nonceBytes),
	}

	return witnessCircuit, nil
}
