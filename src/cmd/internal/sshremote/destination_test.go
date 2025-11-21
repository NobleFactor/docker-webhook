package sshremote

import (
    "net"
    "testing"
)

func TestParseDestination(t *testing.T) {
    tests := []struct {
        name             string
        in               string
        wantUserNonEmpty bool
        wantUser         string
        wantAddr         string
        wantErr          bool
    }{
        {name: "ssh URI no port", in: "ssh://alice@example.com", wantUser: "alice", wantAddr: net.JoinHostPort("example.com", "22")},
        {name: "ssh URI with port", in: "ssh://bob@host:2222", wantUser: "bob", wantAddr: net.JoinHostPort("host", "2222")},
        {name: "plain user@host", in: "carol@host.example", wantUser: "carol", wantAddr: net.JoinHostPort("host.example", "22")},
        {name: "plain host", in: "host.local", wantUserNonEmpty: true, wantAddr: net.JoinHostPort("host.local", "22")},
        {name: "plain host with colon accepted", in: "host:2222", wantUserNonEmpty: true, wantAddr: net.JoinHostPort("host:2222", "22")},
        {name: "ambiguous unbracketed ipv6 accepted", in: "fe80::1:2222", wantUserNonEmpty: true, wantAddr: net.JoinHostPort("fe80::1:2222", "22")},
        {name: "bracketed ipv6 via uri", in: "ssh://dave@[fe80::1]:2222", wantUser: "dave", wantAddr: net.JoinHostPort("fe80::1", "2222")},
        {name: "empty input", in: "", wantErr: true},
        {name: "empty username", in: "@host", wantErr: true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            user, addr, err := parseDestination(tt.in)
            if tt.wantErr {
                if err == nil {
                    t.Fatalf("expected error, got nil (user=%q addr=%q)", user, addr)
                }
                return
            }
            if err != nil {
                t.Fatalf("unexpected error: %v", err)
            }
            if tt.wantUser != "" {
                if user != tt.wantUser {
                    t.Fatalf("user: want %q, got %q", tt.wantUser, user)
                }
            }
            if tt.wantUserNonEmpty {
                if user == "" {
                    t.Fatalf("user expected to be non-empty")
                }
            }
            if addr != tt.wantAddr {
                t.Fatalf("addr: want %q, got %q", tt.wantAddr, addr)
            }
        })
    }
}
