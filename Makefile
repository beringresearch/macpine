BINARY_NAME := alpine

all: 
	echo "Building ..."
	go clean
	go get
	go build -ldflags=$(GO_LDFLAGS) -o bin/$(BINARY_NAME) *.go
	echo "Installing ..."
	sudo cp -f bin/alpine /usr/local/bin/
	@echo "macpine installed"