BINARY_NAME := alpine

linux:
	@echo "Building ..."
	go clean
	go get
	@GOOS=linux go build -ldflags=$(GO_LDFLAGS) -o bin/$(BINARY_NAME) *.go
	@echo "Installing ..."
ifneq ("$(wildcard $(/usr/local/bin/alpine))","")
	sudo rm /usr/local/bin/alpine
endif
	sudo cp -f bin/alpine /usr/local/bin/
	@echo "macpine installed"

darwin: 
	@echo "Building ..."
	go clean
	go get
	@GOOS=darwin go build -ldflags=$(GO_LDFLAGS) -o bin/$(BINARY_NAME) *.go
	@echo "Installing ..."
ifneq ("$(wildcard $(/usr/local/bin/alpine))","")
	sudo rm /usr/local/bin/alpine
endif
	sudo cp -f bin/alpine /usr/local/bin/
	@echo "macpine installed"
