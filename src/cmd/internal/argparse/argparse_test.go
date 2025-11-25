package argparse

import (
	"bytes"
	"log"
	"net"
	"strings"
	"testing"
)

func TestParseClientIps_MixedValidAndInvalid(t *testing.T) {
	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)

	input := "1.2.3.4, bad, 5.6.7.8"
	got := parseClientIps(input)

	if len(got) != 2 {
		t.Fatalf("expected 2 parsed IPs, got %d", len(got))
	}

	if !got[0].Equal(net.ParseIP("1.2.3.4")) || !got[1].Equal(net.ParseIP("5.6.7.8")) {
		t.Fatalf("parsed IPs do not match expected values: %v", got)
	}

	out := buf.String()
	if !strings.Contains(out, "ignored invalid client IP entries") || !strings.Contains(out, "bad") {
		t.Fatalf("expected warning containing invalid token 'bad', got: %s", out)
	}
	if !strings.Contains(out, "parsed valid IPs") || !strings.Contains(out, "1.2.3.4") || !strings.Contains(out, "5.6.7.8") {
		t.Fatalf("expected warning to include valid parsed IPs, got: %s", out)
	}
}

func TestParseClientIps_AllInvalid(t *testing.T) {
	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)

	input := "not-an-ip"
	got := parseClientIps(input)

	if len(got) != 0 {
		t.Fatalf("expected 0 parsed IPs, got %d", len(got))
	}

	out := buf.String()
	if !strings.Contains(out, "not-an-ip") {
		t.Fatalf("expected warning containing invalid token 'not-an-ip', got: %s", out)
	}
	if !strings.Contains(out, "parsed valid IPs") {
		t.Fatalf("expected warning to include a parsed valid IPs section, got: %s", out)
	}
}

func TestParseClientIps_Empty(t *testing.T) {
	var buf bytes.Buffer
	old := log.Writer()
	log.SetOutput(&buf)
	defer log.SetOutput(old)

	input := ""
	got := parseClientIps(input)

	if got != nil && len(got) != 0 {
		t.Fatalf("expected nil/empty slice for empty input, got: %v", got)
	}

	out := buf.String()
	if out != "" {
		t.Fatalf("expected no warnings for empty input, got: %s", out)
	}
}
