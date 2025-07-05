install:
	go mod init zkp
	go mod tidy

run:
	mkdir -p data
	mkdir -p src
	go run main.go readwrite.go
	forge test -vvv