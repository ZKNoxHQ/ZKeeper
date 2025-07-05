clean:
	rm *.bin *.json solidity/src/Verifier.sol solidity/test/Verifier.t.sol go.mod go.sum

run:
	go mod init mopro-gnark
	go mod tidy
	go run secp256k1_Plonk.go