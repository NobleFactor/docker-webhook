// SPDX-FileCopyrightText: 2016-2025 Noble Factor
// SPDX-License-Identifier: MIT
package sshremote

import (
	"fmt"
	"net"
	urlpkg "net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

// ParseSshDestination parses an SSH destination and returns (address, *ssh.ClientConfig, error)
func ParseSshDestination(destination string, configDirectory string) (string, *ssh.ClientConfig, error) {

	username, address, err := parseDestination(strings.TrimSpace(destination))
	if err != nil {
		return "", nil, err
	}

	// Load private key

	keyPath := filepath.Join(configDirectory, "ssh", "id_rsa")
	keyBytes, err := os.ReadFile(keyPath)

	if err != nil {
		return "", nil, fmt.Errorf("failed to read private key: %v", err)
	}

	signer, err := ssh.ParsePrivateKey(keyBytes)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	config := &ssh.ClientConfig{
		User:            username,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	return address, config, nil
}

// Parse destination with support for plain form ([user@]host or [user@]host:port) and URI form (ssh://[user@]host[:port]).
func parseDestination(destination string) (string, string, error) {

	destination = strings.TrimSpace(destination)
	if destination == "" {
		return "", "", fmt.Errorf("invalid ssh destination: empty")
	}

	var username, host, port string
	const defaultPort = "22"

	if strings.HasPrefix(destination, "ssh://") {

		// URI form: ssh://[user@]hostname[:port]

		url, err := urlpkg.Parse(destination)
		if err != nil {
			return "", "", fmt.Errorf("invalid ssh URI %q: %v", destination, err)
		}

		if url.User != nil && url.User.Username() != "" {
			username = url.User.Username()
		} else {
			currentUser, err := user.Current()
			if err != nil {
				return "", "", fmt.Errorf("unable to determine current user: %v", err)
			}
			username = currentUser.Username
		}

		host = url.Hostname()
		port = url.Port()

		if host == "" {
			return "", "", fmt.Errorf("invalid ssh URI: missing host")
		}

		if port == "" {
			port = defaultPort
		}

	} else {

		// Plain form: [user@]host

		atIndex := strings.Index(destination, "@")

		if atIndex != -1 {
			username = destination[:atIndex]
			host = destination[atIndex+1:]
			if username == "" {
				return "", "", fmt.Errorf("invalid ssh destination: empty username")
			}
			if host == "" {
				return "", "", fmt.Errorf("invalid ssh destination: empty host")
			}
		} else {
			currentUser, err := user.Current()
			if err != nil {
				return "", "", fmt.Errorf("unable to determine current user: %v", err)
			}
			username = currentUser.Username
			host = destination
			if host == "" {
				return "", "", fmt.Errorf("invalid ssh destination: empty host")
			}
		}

		port = defaultPort
	}

	return username, net.JoinHostPort(host, port), nil
}
