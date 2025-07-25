package circuit

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_emulated"
	"github.com/consensys/gnark/std/conversion"
	"github.com/consensys/gnark/std/hash/sha2"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/std/math/emulated/emparams"
	"github.com/consensys/gnark/std/math/uints"
	"github.com/consensys/gnark/std/signature/ecdsa"
)

// VerifCircuit defines the circuit structure as provided by you.
type VerifCircuit[T, S emulated.FieldParams] struct {
	// PublicKeyHash is 32 bytes, but we split it into two 16-byte variables to fit into BN254 field
	PublicKeyHash [2]frontend.Variable  `gnark:",public"`
	Signature     ecdsa.Signature[S]    `gnark:",public"`
	Msg           emulated.Element[S]   // if tag is not set, then it is a private input
	PublicKey     ecdsa.PublicKey[T, S] // actual public key is also a private input
	Nonce         frontend.Variable
}

func (c *VerifCircuit[T, S]) Define(api frontend.API) error {

	// Convert the public key into X and Y bytes
	xbytes, err := conversion.EmulatedToBytes(api, &c.PublicKey.X)
	if err != nil {
		return fmt.Errorf("failed to convert PublicKey.X to bytes: %w", err)
	}
	ybytes, err := conversion.EmulatedToBytes(api, &c.PublicKey.Y)
	if err != nil {
		return fmt.Errorf("failed to convert PublicKey.Y to bytes: %w", err)
	}
	// Convert Nonce into bytes
	var noncebytes []uints.U8
	bts, err := conversion.NativeToBytes(api, c.Nonce)
	if err != nil {
		return fmt.Errorf("failed to convert Nonce to bytes: %w", err)
	}
	// NativeToBytes returns 32 bytes (MSB order), but we set only 20 bytes so take the last 20 bytes
	noncebytes = bts[12:]

	// Compute the SHA2 hash as a public key commitment
	h, err := sha2.New(api)
	if err != nil {
		return fmt.Errorf("failed to create SHA2 instance: %w", err)
	}
	h.Write(xbytes)
	h.Write(ybytes)
	h.Write(noncebytes)
	computedHash := h.Sum()

	// Check the hash against the public commitment
	// Initialize bytes gadget for comparison
	bapi, err := uints.NewBytes(api)
	if err != nil {
		return fmt.Errorf("failed to create bytes gadget: %w", err)
	}

	//  The hash is 256 bits, fitting in 2 frontend.Variable (as the field is 254 bit long)
	var hashpubkeybytes []uints.U8
	for i := range c.PublicKeyHash {
		bts, err := conversion.NativeToBytes(api, c.PublicKeyHash[i])
		if err != nil {
			return fmt.Errorf("failed to convert PublicKeyHash[%d] to bytes: %w", i, err)
		}
		// NativeToBytes returns 32 bytes (MSB order), but we set only 16 bytes so take the last 16 bytes
		hashpubkeybytes = append(hashpubkeybytes, bts[16:]...)
	}
	if len(hashpubkeybytes) != len(computedHash) {
		return fmt.Errorf("hashpubkeybytes and computedHash have different lengths: %d vs %d", len(hashpubkeybytes), len(computedHash))
	}
	// Checkgin byte by byte
	for i := range hashpubkeybytes {
		bapi.AssertIsEqual(hashpubkeybytes[i], computedHash[i])
	}

	// Verifying the ECDSA signature
	c.PublicKey.Verify(api, sw_emulated.GetCurveParams[emparams.Secp256k1Fp](), &c.Msg, &c.Signature)

	return nil
}

// Struct for JSON serialization of the transaction inputs.
type TransactionInput struct {
	MsgHash string `json:"msgHash"` // Hex string of the message hash
	R       string `json:"r"`       // Hex string of signature R
	S       string `json:"s"`       // Hex string of signature S
	PubX    string `json:"pubX"`    // Hex string of public key X
	PubY    string `json:"pubY"`    // Hex string of public key Y
}

// Struct for JSON serialization of the witnesses of the proof.
type ProofWitness struct {
	MsgHash string `json:"msgHash"` // Hex string of the message hash
	R       string `json:"r"`       // Hex string of signature R
	S       string `json:"s"`       // Hex string of signature S
	PubX    string `json:"pubX"`    // Hex string of public key X
	PubY    string `json:"pubY"`    // Hex string of public key Y
	Nonce   string `json:"Nonce"`   // Hex string of Nonce
	Com     string `json:"Com"`     // Hex string of Commitment
}

// readFromFile is a helper to deserialize and read gnark objects or JSON from files.
func ReadTransactionFromFile(filename string, data interface{}) ([]byte, []byte, []byte, []byte, []byte, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, nil, nil, nil, fmt.Errorf("error opening file %s: %w", filename, err)
	}
	defer file.Close()

	switch v := data.(type) {
	case *TransactionInput:
		decoder := json.NewDecoder(file)
		err = decoder.Decode(v)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("error decoding JSON from file %s: %w\n", filename, err)
		}
		// Decode hex strings back to big.Int and byte slices
		rBytes, err := hex.DecodeString(v.R)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("Error decoding R hex: %v\n", err)
		}
		sBytes, err := hex.DecodeString(v.S)
		if err != nil {
			return rBytes, nil, nil, nil, nil, fmt.Errorf("Error decoding S hex: %v\n", err)
		}
		msgHashBytes, err := hex.DecodeString(v.MsgHash)
		if err != nil {
			return rBytes, sBytes, nil, nil, nil, fmt.Errorf("Error decoding S hex: %v\n", err)
		}
		pubXBytes, err := hex.DecodeString(v.PubX)
		if err != nil {
			return rBytes, sBytes, msgHashBytes, nil, nil, fmt.Errorf("Error decoding S hex: %v\n", err)
		}
		pubYBytes, err := hex.DecodeString(v.PubY)
		if err != nil {
			return rBytes, sBytes, msgHashBytes, pubXBytes, nil, fmt.Errorf("Error decoding S hex: %v\n", err)
		}
		return rBytes, sBytes, msgHashBytes, pubXBytes, pubYBytes, nil

	default:
		return nil, nil, nil, nil, nil, fmt.Errorf("unsupported type for reading from file: %T", data)
	}

}
