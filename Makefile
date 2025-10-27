BINARY_NAME=gophkeeper
VERSION=1.0.0

build:
	go build -o $(BINARY_NAME) cmd/client/main.go

build-all:
	GOOS=linux GOARCH=amd64 go build -o bin/$(BINARY_NAME)-linux-amd64 cmd/client/main.go
	GOOS=windows GOARCH=amd64 go build -o bin/$(BINARY_NAME)-windows-amd64.exe cmd/client/main.go
	GOOS=darwin GOARCH=amd64 go build -o bin/$(BINARY_NAME)-darwin-amd64 cmd/client/main.go
	GOOS=darwin GOARCH=arm64 go build -o bin/$(BINARY_NAME)-darwin-arm64 cmd/client/main.go

clean:
	rm -f $(BINARY_NAME) bin/*

test:
	go test ./...

.PHONY: build build-all clean test