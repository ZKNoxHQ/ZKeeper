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