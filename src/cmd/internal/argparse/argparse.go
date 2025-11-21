// Package argparse provides command line argument parsing for webhook-executor
// SPDX-FileCopyrightText: 2016-2025 Noble Factor
// SPDX-License-Identifier: MIT
package argparse

import (
	"flag"
	"fmt"
	"net"
	"strings"
)

// parseClientIps parses the X-Forwarded-For header value into a slice of valid net.IP addresses
func parseClientIps(headerValue string) []net.IP {
	if headerValue == "" {
		return nil
	}
	parts := strings.Split(headerValue, ",")
	clientIps := make([]net.IP, 0, len(parts))
	for _, part := range parts {
		ip := net.ParseIP(strings.TrimSpace(part))
		if ip != nil {
			clientIps = append(clientIps, ip)
		}
	}
	return clientIps
}

// ParsedArgs holds the parsed command line arguments
type ParsedArgs struct {
	Destination   string
	Command       string
	AuthHeader    string
	ClientIps     []net.IP
	CorrelationId string
}

// ParseArguments parses command line flags and returns the values.
// Returns: ParsedArgs, error
func ParseArguments(args []string) (ParsedArgs, error) {

	var destination = flag.String("destination", "", "SSH destination (e.g., user@host or host)")
	var command = flag.String("command", "", "Command to execute on the remote host")
	var authorization = flag.String("authorization", "", "JWT token from Authorization Bearer header")
	var correlationId = flag.String("correlation-id", "", "Correlation ID for traceability (auto-generated if not provided)")
	var xForwardedFor = flag.String("X-Forwarded-For", "", "Client IP chain from X-Forwarded-For header")
	var help = flag.Bool("help", false, "Show help message")

	flagSet := flag.NewFlagSet("webhook-executor", flag.ContinueOnError)
	flagSet.StringVar(destination, "destination", "", "SSH destination (e.g., user@host or host)")
	flagSet.StringVar(command, "command", "", "Command to execute on the remote host")
	flagSet.StringVar(authorization, "authorization", "", "JWT token from Authorization Bearer header")
	flagSet.StringVar(correlationId, "correlation-id", "", "Correlation ID for traceability (auto-generated if not provided)")
	flagSet.StringVar(xForwardedFor, "X-Forwarded-For", "", "Client IP chain from X-Forwarded-For header")
	flagSet.BoolVar(help, "help", false, "Show help message")

	err := flagSet.Parse(args)
	if err != nil {
		return ParsedArgs{}, fmt.Errorf("unable to parse arguments: %v", err)
	}

	if *help {

		return ParsedArgs{}, fmt.Errorf("help requested")
	}

	// Basic validation

	if *destination == "" {
		return ParsedArgs{}, fmt.Errorf("--destination is required")
	}
	if *command == "" {
		return ParsedArgs{}, fmt.Errorf("--command is required")
	}
	if *authorization == "" {
		return ParsedArgs{}, fmt.Errorf("--authorization is required")
	}

	return ParsedArgs{
		Destination:   *destination,
		Command:       *command,
		AuthHeader:    *authorization,
		ClientIps:     parseClientIps(*xForwardedFor),
		CorrelationId: *correlationId,
	}, nil
}
