BINARY_NAME := alpine
PREFIX = /usr/local
bindir = $(DESTDIR)$(PREFIX)/bin

bin/$(BINARY_NAME):
	@echo "Building ..."
	go clean
	go get
	go build -ldflags=$(GO_LDFLAGS) -o $@ *.go

.PHONY: install
install: bin/$(BINARY_NAME)
	@echo "Installing ..."
	install -d $(bindir)
	install -m 0755 $^ $(bindir)
	@echo "macpine installed"
