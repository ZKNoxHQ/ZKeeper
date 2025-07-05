install:
	@if [ ! -f go.mod ]; then \
		echo "Initializing Go module..."; \
		go mod init zkp; \
	else \
		echo "go.mod already exists."; \
	fi
	go mod tidy

run:
	mkdir -p data
	mkdir -p src
	go run main.go readwrite.go
	forge test -vvv