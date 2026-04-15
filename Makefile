.PHONY: build test lint clean run

BINARY_NAME=kairo

build:
	go build -trimpath -ldflags "-s -w" -o $(BINARY_NAME) ./cmd/kairo

test:
	go test ./...

lint:
	golangci-lint run

clean:
	go clean
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_NAME).exe

run: build
	./$(BINARY_NAME)

install:
	go install ./cmd/kairo
