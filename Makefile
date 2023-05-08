BINARY_NAME := alpine
BUILD_DIR = bin
PREFIX = /usr/local
XCOMP_TARGETS = amd64 arm64
bindir = $(DESTDIR)$(PREFIX)/bin
SRCS = $(wildcard */*.go)
MAIN = main.go

$(BUILD_DIR)/$(BINARY_NAME): $(MAIN) $(SRCS)
	@echo "Building ..."
	go clean
	go get
	go build -ldflags=$(GO_LDFLAGS) -o $@ $<

$(BUILD_DIR)/$(BINARY_NAME)_darwin_%: $(MAIN) $(SRCS)
	CGO_ENABLED=1 gox -ldflags "${LDFLAGS}" -output="$@" --osarch="darwin/$*"

.PHONY: install clean fmt agent xcompile
install: $(BUILD_DIR)/$(BINARY_NAME)
	@echo "Installing ..."
	install -d $(bindir)
	install -m 0755 $^ $(bindir)
	@echo "macpine installed"

clean:
	@echo "Cleaning ..."
	go clean
	$(RM) -r $(BUILD_DIR)

fmt:
	@echo "Formatting ..."
	@gofmt -e -l -s -w $(MAIN) $(SRCS)

agent:
	@echo "Installing agent ..."
	@if [ -f "$(shell which alpine)" ]; then \
		sed "s/{{ ALPINE_PATH }}/$(subst /,\\/,$(shell which -a alpine))/" utils/alpineDaemonLaunchAgent.plist > compiledAgent ; \
		mv compiledAgent ~/Library/LaunchAgents/alpineDaemonLaunchAgent.plist ; \
		echo "installed launch agent to ~/Library/LaunchAgents" ; \
	else \
		echo "error: macpine not found in PATH" >&2 ; exit 1 ; \
	fi

xcompile: $(patsubst %,$(BUILD_DIR)/$(BINARY_NAME)_darwin_%,$(XCOMP_TARGETS))
