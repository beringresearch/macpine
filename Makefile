BINARY_NAME := alpine

darwin: 
	@echo "Building ..."
	go clean
	go get
	@GOOS=darwin go build -ldflags=$(GO_LDFLAGS) -o bin/$(BINARY_NAME) *.go
	@echo "Installing ..."
	sudo rm /usr/local/bin/alpine
	sudo cp -f bin/alpine /usr/local/bin/
	@echo "macpine installed"