BINARY_NAME := alpine
PREFIX := /usr/local
bindir = $(DESTDIR)$(PREFIX)/bin

linux:
	@echo "Building ..."
	go clean
	go get
	@GOOS=linux go build -ldflags=$(GO_LDFLAGS) -o bin/$(BINARY_NAME) *.go
	@echo "Installing ..."
ifneq ("$(wildcard $($(bindir)/alpine))","")
	sudo rm $(bindir)/alpine
endif
	sudo cp -f bin/alpine $(bindir)
	@echo "macpine installed"

darwin: 
	@echo "Building ..."
	go clean
	go get
	@GOOS=darwin go build -ldflags=$(GO_LDFLAGS) -o bin/$(BINARY_NAME) *.go
	@echo "Installing ..."
ifneq ("$(wildcard $($(bindir)/alpine))","")
	sudo rm $(bindir)/alpine
endif
	sudo cp -f bin/alpine $(destdir)
	@echo "macpine installed"
