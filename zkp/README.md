# Zero-knowledge proof of transaction

## Proof system
The ZK proof is computed using GNARK with PLONK proof system. It requires a trusted setup that has to be computed **once**.
The proof takes a few seconds to be generated and can be verified on-chain with our Solidity contract.

## Trusted setup
:warning: Defining a new trusted setup requires updating the on-chain contracts. This can be done using:
```
go run trusted_setup.go
```
It creates a file `r1cs.bin` containing the setup, and also the corresponding Solidity contract `solidity/src/Verifier.sol`.

## Witness generation
From a signed transaction `signed_transaction.json`, the witness is generated and output in a file `witness_input.json` using:
```
go run derive_pkcommit
```

## Proving
From the witness file `witness_input.json`, the zero-knowledge proof is computed using:
```
go run prove_blinded_k1
```
This creates a Solidity test file `solidity/test/Verifier.t.sol`.

## Verification
The proof can be verified using the solidity contract. It can be checked with:
```
cd solidty/
forge test -vvv
```

## Usage during the Hackathon
```bash
./private_proof <msgHash> <r> <s> <pubX> <pubY>
```
An example of working transaction:
```bash
./private_proof \
74657374696e6720454344534120287072652d68617368656429 \
847a2bb5c0b16efca1ceb70d79d4b9d4be0d29ecf4dad712f944953e6c33758d \
955e43f9edcf0f74cfa2b78adebcc8c8c8f30946d4820da362a7b490b321946a \
508e802faf338c15a571878f8be339e7442e582680fab0d0ad835672e0705471 \
d4a3fe56add0155c1ce79810a20e5c431488e79fbd3d2f425e20ecd0924eeaa4

```