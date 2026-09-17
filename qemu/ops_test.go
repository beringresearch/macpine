package qemu

import (
	"net"
	"testing"
	"time"
)

func TestShouldShowResolveIPHint(t *testing.T) {
	cases := []struct {
		name         string
		sawCandidate bool
		hintShown    bool
		elapsed      time.Duration
		want         bool
	}{
		{"no candidate yet, still early boot", false, false, time.Hour, false},
		{"candidate seen but too soon", true, false, resolveIPHintDelay - time.Second, false},
		{"candidate seen and past the delay", true, false, resolveIPHintDelay + time.Second, true},
		{"already shown once", true, true, resolveIPHintDelay + time.Hour, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := shouldShowResolveIPHint(tc.sawCandidate, tc.hintShown, tc.elapsed)
			if got != tc.want {
				t.Errorf("shouldShowResolveIPHint(%v, %v, %v) = %v, want %v",
					tc.sawCandidate, tc.hintShown, tc.elapsed, got, tc.want)
			}
		})
	}
}

func TestReachable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test listener: %v", err)
	}
	defer ln.Close()

	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	_, openPort, err := net.SplitHostPort(ln.Addr().String())
	if err != nil {
		t.Fatalf("failed to parse listener address: %v", err)
	}

	c := MachineConfig{SSHPort: openPort}
	if !c.reachable("127.0.0.1") {
		t.Errorf("expected 127.0.0.1:%s to be reachable (a listener is running there)", openPort)
	}

	// Grab a port, close the listener, then dial it - nothing is
	// listening any more, so this should be reported unreachable and
	// return promptly (connection refused) rather than hang.
	closedLn, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to allocate a port to close: %v", err)
	}
	_, closedPort, _ := net.SplitHostPort(closedLn.Addr().String())
	closedLn.Close()

	cClosed := MachineConfig{SSHPort: closedPort}
	if cClosed.reachable("127.0.0.1") {
		t.Errorf("expected 127.0.0.1:%s to be unreachable (nothing is listening)", closedPort)
	}
}
