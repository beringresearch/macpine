BINARY_NAME := alpine
PREFIX = /usr/local
bindir = $(DESTDIR)$(PREFIX)/bin
SRCS = $(wildcard cmd/*.go host/*.go qemu/*.go utils/*.go)
MAIN = main.go

bin/$(BINARY_NAME): $(MAIN) $(SRCS)
	@echo "Building ..."
	go clean
	go get
	go build -ldflags=$(GO_LDFLAGS) -o $@ $<

.PHONY: install clean
install: bin/$(BINARY_NAME)
	@echo "Installing ..."
	install -d $(bindir)
	install -m 0755 $^ $(bindir)
	@echo "macpine installed"

clean:
	@echo "Cleaning ..."
	go clean
	$(RM) -r bin/
