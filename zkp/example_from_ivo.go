package main

import (
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"math/big"

	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/emulated/sw_emulated"

	"github.com/consensys/gnark/std/conversion"
	"github.com/consensys/gnark/std/hash/sha2"
	"github.com/consensys/gnark/std/math/emulated"
	"github.com/consensys/gnark/std/math/emulated/emparams"
	"github.com/consensys/gnark/std/math/uints"
	"github.com/consensys/gnark/std/signature/ecdsa"
	"github.com/consensys/gnark/test"

	"github.com/consensys/gnark-crypto/ecc"
	p256_ecdsa "github.com/consensys/gnark-crypto/ecc/secp256k1/ecdsa"
)

// we check that the hash of the public key is correctly computed using SHA2
type PubKeySHA2 struct {
	// PublicKeyHash is 32 bytes, but we split it into two 16-byte variables to fit into BN254 field
	PublicKeyHash [2]frontend.Variable                                        `gnark:",public"`
	Signature     ecdsa.Signature[emparams.Secp256k1Fr]                       `gnark:",public"`
	Msg           emulated.Element[emparams.Secp256k1Fr]                      // if tag is not set, then it is a private input
	PublicKey     ecdsa.PublicKey[emparams.Secp256k1Fp, emparams.Secp256k1Fr] // actual public key is also a private input
}

func (c *PubKeySHA2) Define(api frontend.API) error {
	// -- hash the given public key
	//  - first we convert the public key coordinates to bytes
	xbytes, err := conversion.EmulatedToBytes(api, &c.PublicKey.X)
	if err != nil {
		return fmt.Errorf("failed to convert PublicKey.X to bytes: %w", err)
	}
	ybytes, err := conversion.EmulatedToBytes(api, &c.PublicKey.Y)
	if err != nil {
		return fmt.Errorf("failed to convert PublicKey.Y to bytes: %w", err)
	}
	//  - now we compute the SHA2 hash of the concatenated bytes
	h, err := sha2.New(api)
	if err != nil {
		return fmt.Errorf("failed to create SHA2 instance: %w", err)
	}
	h.Write(xbytes)
	h.Write(ybytes)
	//  - and compute the hash
	computedHash := h.Sum()
	// -- now we check that the computed hash matches the expected hash
	//  - first, we used [2]frontend.Variable to store the hash so that we wouldn't be using too much public inputs and we want the parts to fit into BN254 field, so 16-byte chunks
	//    we convert it back to bytes
	var hashpubkeybytes []uints.U8
	for i := range c.PublicKeyHash {
		bts, err := conversion.NativeToBytes(api, c.PublicKeyHash[i])
		if err != nil {
			return fmt.Errorf("failed to convert PublicKeyHash[%d] to bytes: %w", i, err)
		}
		// NativeToBytes returns 32 bytes (MSB order), but we set only 16 bytes so take the last 16 bytes
		hashpubkeybytes = append(hashpubkeybytes, bts[16:]...)
	}
	//  - now we need to initialize bytes gadget for comparison
	bapi, err := uints.NewBytes(api)
	if err != nil {
		return fmt.Errorf("failed to create bytes gadget: %w", err)
	}
	if len(hashpubkeybytes) != len(computedHash) {
		return fmt.Errorf("hashpubkeybytes and computedHash have different lengths: %d vs %d", len(hashpubkeybytes), len(computedHash))
	}
	//  - finally we check that the computed hash matches the expected hash
	for i := range hashpubkeybytes {
		bapi.AssertIsEqual(hashpubkeybytes[i], computedHash[i])
	}

	// -- now we check that the signature is valid
	c.PublicKey.Verify(api, sw_emulated.GetCurveParams[emparams.Secp256k1Fp](), &c.Msg, &c.Signature)

	return nil
}

func main() {

	// generate random key pair
	sk, err := p256_ecdsa.GenerateKey(rand.Reader)
	// assert.NoError(err, "failed to generate key")
	fmt.Println("err", err)
	pubkey := sk.PublicKey

	// compute the hash of the public key
	h := sha256.New()
	h.Write(pubkey.Bytes())
	pubHash := h.Sum(nil)
	pubHashLo := pubHash[:16]
	pubHashHi := pubHash[16:32]

	msg := []byte("this is a test message for pubkey hashing!")
	// obtain the signature
	sig, err := sk.Sign(msg, sha256.New())
	// assert.NoError(err, "failed to sign message")
	// sanity check
	ok, err := pubkey.Verify(sig, msg, sha256.New())
	// assert.NoError(err, "failed to verify signature")
	// assert.True(ok, "signature verification failed")
	fmt.Println("err", err)
	fmt.Println("OK", ok)

	// the signature has concatenated R and S values. Lets unwrap them
	var sigT p256_ecdsa.Signature
	_, err = sigT.SetBytes(sig)
	// assert.NoError(err, "failed to set bytes for signature")
	fmt.Println("err", err)
	r, s := new(big.Int), new(big.Int)
	r.SetBytes(sigT.R[:32])
	s.SetBytes(sigT.S[:32])

	// compute the hash of the message as an integer
	mshHash := sha256.Sum256(msg)
	msgHashInt := p256_ecdsa.HashToInt(mshHash[:])

	// now we prepare the witness for the circuit
	assignment := &PubKeySHA2{
		// we splitted the public key hash into two 16-byte variables to fit into BN254 field
		PublicKeyHash: [2]frontend.Variable{pubHashLo, pubHashHi},
		// we construct the public key as non-native element. NB! this means that both X and Y coordinates are 4 limbs of 64 bytes each, so 8 limbs total
		PublicKey: ecdsa.PublicKey[emparams.Secp256k1Fp, emparams.Secp256k1Fr]{
			X: emulated.ValueOf[emulated.Secp256k1Fp](pubkey.A.X),
			Y: emulated.ValueOf[emulated.Secp256k1Fp](pubkey.A.Y),
		},
		Signature: ecdsa.Signature[emparams.Secp256k1Fr]{
			R: emulated.ValueOf[emparams.Secp256k1Fr](r),
			S: emulated.ValueOf[emparams.Secp256k1Fr](s),
		},
		Msg: emulated.ValueOf[emparams.Secp256k1Fr](msgHashInt),
	}

	err = test.IsSolved(&PubKeySHA2{}, assignment, ecc.BN254.ScalarField())
	// assert.NoError(err, "failed to solve the circuit")
	fmt.Println("err", err)
}
