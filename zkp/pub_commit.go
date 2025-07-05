package main

import (
	"bytes"
	"io"

	"crypto/rand"

	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	// Added for performance timing

	// cryptoposeidon2 "github.com/consensys/gnark-crypto/ecc/bn254/fr/poseidon2"
	cryptomimc "github.com/consensys/gnark-crypto/ecc/bn254/fr/mimc"
)

// ProveInputEcdsa struct for JSON serialization of witness inputs.
type Input struct {
	PubX string `json:"pubX"` // Hex string of public key X
}

// ProveInputEcdsa struct for JSON serialization of witness inputs.
type InputWithCommit struct {
	MsgHash string `json:"msgHash"` // Hex string of msgHash
	R       string `json:"r"`       // Hex string of r
	S       string `json:"s"`       // Hex string of s
	PubX    string `json:"pubX"`    // Hex string of public key X
	PubY    string `json:"pubY"`    // Hex string of public key Y
	Address string `json:"address"` // Hex string of address
	Nonce   string `json:"nonce"`   // Hex string of nonce
	Com     string `json:"com"`     // Hex string of Com
}

func main() {

	var loadedInput Input
	err := readFromFile("pub_x.json", &loadedInput)
	if err != nil {
		fmt.Printf("Error reading pub_x.json: %v\n", err)
		os.Exit(1)
	}

	// Decode hex strings back to big.Int and byte slices for witness construction
	pubXBytes, err := hex.DecodeString(loadedInput.PubX)
	if err != nil {
		fmt.Printf("Error decoding PubX hex: %v\n", err)
		os.Exit(1)
	}

	// address-like derived from public key x-coordinate
	address := pubXBytes[:20]

	// 160 bits
	nonce := make([]byte, 20)
	_, err = rand.Read(nonce)
	if err != nil {
		panic(err)
	}

	// PK Commitment
	h := cryptomimc.NewMiMC()
	_, err = h.Write(address)
	if err != nil {
		panic(err)
	}
	_, err = h.Write(nonce)
	if err != nil {
		panic(err)
	}
	ComPK := h.Sum(nil)

	Output := InputWithCommit{
		MsgHash: "",
		R:       "",
		S:       "",
		PubX:    hex.EncodeToString(pubXBytes[:]),
		PubY:    "",
		Address: hex.EncodeToString(address[:]),
		Nonce:   hex.EncodeToString(nonce[:]),
		Com:     hex.EncodeToString(ComPK),
	}

	OutputJSON, err := json.MarshalIndent(Output, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling prove input JSON: %v\n", err)
		os.Exit(1)
	}

	writeToFile("witness_input.json", bytes.NewReader(OutputJSON))

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
	case *Input: // For the JSON input
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
