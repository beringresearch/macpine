package qemu

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"syscall"
	"testing"
)

func TestDialHint(t *testing.T) {
	err := &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connect", syscall.EHOSTUNREACH)}
	got := dialHint(fmt.Errorf("wrapped: %w", err))
	if runtime.GOOS == "darwin" && !strings.Contains(got, "Local Network") {
		t.Errorf("expected Local Network hint on darwin, got %q", got)
	}
	if got := dialHint(syscall.ECONNREFUSED); got != "" {
		t.Errorf("unexpected hint for ECONNREFUSED: %q", got)
	}
}
