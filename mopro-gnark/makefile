clean:
	rm *.bin *.json solidity/src/Verifier.sol solidity/test/Verifier.t.sol go.mod go.sum

run: 
	rm go.mod go.sum
	go mod init mopro-gnark
	go mod tidy
	mkdir -p solidity/src
	mkdir -p solidity/test
	go run secp256k1_Plonk.go