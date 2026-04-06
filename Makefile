.PHONY: build test lint fmt run install clean

BINARY=leetcli
VERSION?=dev
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BINARY) .

test:
	go test -v -race ./...

lint:
	golangci-lint run

fmt:
	gofmt -w .

run: build
	./$(BINARY)

install: build
	cp $(BINARY) ~/go/bin/

clean:
	rm -f $(BINARY)

release-dry:
	goreleaser release --snapshot --clean