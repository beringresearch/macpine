BINARY_NAME := alpine
BUILD_DIR = bin
PREFIX = /usr/local
bindir = $(DESTDIR)$(PREFIX)/bin
SRCS = $(wildcard */*.go)
MAIN = main.go

$(BUILD_DIR)/$(BINARY_NAME): $(MAIN) $(SRCS)
	@echo "Building ..."
	go clean
	go get
	go build -ldflags=$(GO_LDFLAGS) -o $@ $<

.PHONY: install clean fmt agent
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
	@gofmt -e -l -s -w $(MAIN) $(SRCS)

agent:
	@if [ -f "$(shell which alpine)" ]; then \
		sed "s/{{ ALPINE_PATH }}/$(subst /,\\/,$(shell which -a alpine))/" utils/alpineDaemonLaunchAgent.plist > compiledAgent ; \
		mv compiledAgent ~/Library/LaunchAgents/alpineDaemonLaunchAgent.plist ; \
		echo "installed launch agent to ~/Library/LaunchAgents" ; \
	else \
		echo "error: macpine not found in PATH" >&2 ; exit 1 ; \
	fi
