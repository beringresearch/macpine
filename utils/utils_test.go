package utils

import (
	"fmt"
	"testing"
	"time"
)

func leaseHex(expiresIn time.Duration) string {
	return fmt.Sprintf("0x%x", time.Now().Add(expiresIn).Unix())
}

func TestMatchHwAddressCandidates(t *testing.T) {
	targetMAC := "56:84:17:14:59:9d"

	data := []DhcpData{
		// Expired lease for the target MAC - must be excluded entirely.
		{IpAddress: "192.168.2.4", HwAddress: targetMAC, Lease: leaseHex(-time.Hour)},
		// A leftover-but-still-unexpired lease from an earlier session,
		// with a *later* expiry than the fresh one below (the scenario
		// that used to make the stale IP win outright).
		{IpAddress: "192.168.2.5", HwAddress: targetMAC, Lease: leaseHex(55 * time.Minute)},
		// The genuinely fresh lease for the current session, expiring
		// sooner only because it was issued more recently.
		{IpAddress: "192.168.105.2", HwAddress: targetMAC, Lease: leaseHex(50 * time.Minute)},
		// A different MAC entirely - must never be returned.
		{IpAddress: "192.168.2.6", HwAddress: "56:a3:de:96:79:7f", Lease: leaseHex(time.Hour)},
	}

	candidates := MatchHwAddressCandidates(data, targetMAC)

	if len(candidates) != 2 {
		t.Fatalf("expected 2 unexpired candidates for %s, got %d: %+v", targetMAC, len(candidates), candidates)
	}

	if candidates[0].IpAddress != "192.168.2.5" {
		t.Errorf("expected the later-expiry lease (192.168.2.5) first, got %s", candidates[0].IpAddress)
	}
	if candidates[1].IpAddress != "192.168.105.2" {
		t.Errorf("expected the fresher-but-shorter-remaining lease (192.168.105.2) second, got %s", candidates[1].IpAddress)
	}

	for _, c := range candidates {
		if c.IpAddress == "192.168.2.4" {
			t.Errorf("expired lease 192.168.2.4 must not be returned")
		}
		if c.HwAddress != targetMAC {
			t.Errorf("candidate for wrong hardware address returned: %+v", c)
		}
	}
}

func TestMatchHwAddressCandidatesNoMatch(t *testing.T) {
	data := []DhcpData{
		{IpAddress: "192.168.2.6", HwAddress: "56:a3:de:96:79:7f", Lease: leaseHex(time.Hour)},
	}

	candidates := MatchHwAddressCandidates(data, "56:84:17:14:59:9d")
	if len(candidates) != 0 {
		t.Fatalf("expected no candidates, got %+v", candidates)
	}
}
