.PHONY: build test lint clean run fmt install

BINARY_NAME=kairo

ifeq ($(OS),Windows_NT)
	EXE=.exe
	RM=del /Q
	RUN=.\$(BINARY_NAME)$(EXE)
else
	EXE=
	RM=rm -f
	RUN=./$(BINARY_NAME)
endif

fmt:
	go fmt ./...

build:
	go build -trimpath -ldflags "-s -w" -o $(BINARY_NAME)$(EXE) ./cmd/kairo

test:
	go test ./...

lint:
	golangci-lint run

clean:
	go clean
	$(RM) $(BINARY_NAME)$(EXE)

run: build
	$(RUN)

install:
	go install ./cmd/kairo
	