BINARY_NAME := alpine
BUILD_DIR = build
PREFIX = /usr/local
bindir = $(DESTDIR)$(PREFIX)/bin
SRCS = $(wildcard */*.go)
MAIN = main.go

$(BUILD_DIR)/$(BINARY_NAME): $(MAIN) $(SRCS)
	@echo "Building ..."
	go clean
	go get
	go build -ldflags=$(GO_LDFLAGS) -o $@ $<

.PHONY: install clean fmt
install: $(BUILD_DIR)/$(BINARY_NAME)
	@echo "Installing ..."
	install -d $(bindir)
	install -m 0755 $^ $(bindir)
	@sed "s/{{ ALPINE_PATH }}/$(subst /,\\/,$(bindir)/$(BINARY_NAME))/" utils/alpineDaemonLaunchAgent.plist > $(BUILD_DIR)/alpineDaemonLaunchAgent.plist
	@echo "macpine installed"

clean:
	@echo "Cleaning ..."
	go clean
	$(RM) -r $(BUILD_DIR)

fmt:
	@gofmt -e -l -s -w $(MAIN) $(SRCS)
