BIN=./bin

.PHONY: build

build:
	mkdir -p $(BIN)
	go build -o $(BIN)
