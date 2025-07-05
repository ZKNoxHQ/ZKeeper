clean:
	rm *.bin *.json solidity/src/Verifier.sol solidity/test/Verifier.t.sol go.mod go.sum

run: 
	@if [ -f go.mod ]; then rm go.mod; fi
	@if [ -f go.sum ]; then rm go.sum; fi
	go mod init mopro-gnark
	go mod tidy
	mkdir -p solidity/src
	mkdir -p solidity/test
	go run secp256k1_Plonk.go