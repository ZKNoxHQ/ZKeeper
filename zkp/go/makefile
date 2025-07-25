clean:
	rm -f output/*.bin output/*.json input/*.json go.mod go.sum

init: 
	@if [ -f go.mod ]; then rm go.mod; fi
	@if [ -f go.sum ]; then rm go.sum; fi
	go mod init zkp
	go get github.com/consensys/gnark@master
	go mod tidy
	mkdir -p ../solidity/src
	mkdir -p ../solidity/test

setup:
	go run main.go setup
proof:
	go run main.go