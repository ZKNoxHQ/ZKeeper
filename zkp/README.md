# Zero-knowledge proof of transaction

## How to use

## Public key commitment
In order to obtain the commitment to the public key:
```bash
./pub_commit <pubX>
```
For example,
```bash
./pub_commit 508e802faf338c15a571878f8be339e7442e582680fab0d0ad835672e0705471
```
This creates a file pub_commit.json

## Proof computation
It is possible to compute a ZK proof from a signed transaction:
```bash
./private_proof <msgHash> <r> <s> <pubX> <pubY>
```
An example with a working transaction:
```bash
./private_proof \
74657374696e6720454344534120287072652d68617368656429 \
847a2bb5c0b16efca1ceb70d79d4b9d4be0d29ecf4dad712f944953e6c33758d \
955e43f9edcf0f74cfa2b78adebcc8c8c8f30946d4820da362a7b490b321946a \
508e802faf338c15a571878f8be339e7442e582680fab0d0ad835672e0705471 \
d4a3fe56add0155c1ce79810a20e5c431488e79fbd3d2f425e20ecd0924eeaa4
```
This outputs a proof together with the public inputs. Here is an example:
```
=======================
PROOF and PUBLIC INPUTS
=======================
0x2baccb18071685ce1d48fa8b9f7e84fe6b80e45222ca10eadd83d517e3e4e14f19ec400816f295c3cb4b940bfd880d7bf29c80e597b632c0d16b64546f171bd119b422530d0a00e274ae1f979c63f1e187ff312e666b4c22c3c5121e0642e9e621f9a1020da8ee53ed5c0248a9d4a7b5042f15e89f4be59d2d4bce6a925354ec10d3e3c350f9ec062d288e76c2bba88ccd152cbba88521e3895e36bb2d1331792dff1b2e12b2083f1db9bcec80b338a6c73b4893060a6dd485762f9b43069f41079fa3a6abf9e2d1ad5bd3b13857bb4e8e87cf6e1f7f7e23faa30b730e2a634419345c8312a175ca9ecc8f612ced59ab551b0307a8b76ebe14d7b8fe56d696201d23e2f2f12afefc55bc053b567188c11c3fa2294be0ab89d51ffb3e695e12631c77c9ba9e69923ceb08eb85d6bc6f4784441a3b0525f76e15991ace0d757aad2ae7d349f298a59df6424a440b1920da7623404d128cbdbf22896bcb9f28c04214ad26b2a582daa495a5b4ec6f2b2e3f5893ae8fb530f351f14a35616a0a5d4f2b325f1d9cae5f5bdef2e10547a7549df45496f7360288f52c85a0302578095c10f06a4c532df6faa20985560552b392b5726ba231852513a107dd693ad5820515cab6759dbd41b9cbe0306d7e7a700de88ee69fdbf1afdd8be749c3991aa7e52f6257e7870509e72e82a4cdce0be8a9551b4041ac0c31dcee8e76d51a82a04d21407ceca3e9a63c8af515ec63168775b4ecbac51f47ed7d410ab6097437c6c72e8645fbc94f50da735193fa30b1ce688b6c25772fc7f7d1775f678c032788d21284628d8f4ad1c658a905806a7e27d82bc107aac737bf2dbe61e19d9fb29629143b5a27dcec376cca6f8b16432f2b75fc1fb19ec69829927c95d5f54188ef4127cf390f2cdafdd2b24d5e2862240591f64470f3eba9a3166616e188b7a65f200aa853b65fe737b1bd6ef9f14f7daae972a737e368b56bf3edb9d1f55ff6fd4b1a6fcb7efc45c1af1c756a9d36f94b1ad9a368f9ba0a588a2f58c0677c157b292a9dcf5cf0a89116ebe6e2a131e435790f434b7b721ea4c852152dcc8a18b8841c8bcd0dc473877971ebc9824629c905c8dcd140193279851d346a0b7853cfca232773a25982f65e5661f632b673d6d66fbbe94526f122cca0a15538bc232af417f5cc0c2eade521728e84bb863f69a5d0773879b9ee9adf40a05d5a8f488a50 "[3271972277585273897,4923350424019300965,8319390334557635907,29797,14780897900014725779528081348356307274037329420277504568859839333622488768077]"
```

## Details of the proof system
The ZK proof is computed using GNARK with PLONK proof system. It requires a trusted setup that has to be computed **once**.
The proof takes a few seconds to be generated and can be verified on-chain with our Solidity contract.

### Trusted setup
:warning: Defining a new trusted setup requires updating the on-chain contracts. This can be done using:
```
go run trusted_setup.go
```
It creates a file `r1cs.bin` containing the setup, and also the corresponding Solidity contract `solidity/src/Verifier.sol`.

### Witness generation
From a signed transaction `signed_transaction.json`, the witness is generated and output in a file `witness_input.json` using:
```
go run derive_pkcommit
```

### Proving
From the witness file `witness_input.json`, the zero-knowledge proof is computed using:
```
go run prove_blinded_k1
```
This creates a Solidity test file `solidity/test/Verifier.t.sol`.

### Verification
The proof can be verified using the solidity contract. It can be checked with:
```
cd solidty/
forge test -vvv
```
