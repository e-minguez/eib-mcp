BINARY_NAME=eib-mcp
GO_FILES=$(shell find . -name '*.go')

.PHONY: all build clean test run

all: build

build:
	go build -o $(BINARY_NAME) .

clean:
	rm -f $(BINARY_NAME)

test:
	go test ./...

run: build
	./$(BINARY_NAME)
