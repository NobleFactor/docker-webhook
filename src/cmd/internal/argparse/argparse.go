// Package argparse provides command line argument parsing for webhook-executor
// SPDX-FileCopyrightText: 2016-2025 Noble Factor
// SPDX-License-Identifier: MIT
package argparse

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strings"
)

// parseClientIps parses the X-Forwarded-For header value into a slice of valid net.IP addresses
func parseClientIps(value string) []net.IP {

	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	invalid := make([]string, 0)
	clientIps := make([]net.IP, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			ip := net.ParseIP(part)
			if ip != nil {
				clientIps = append(clientIps, ip)
			} else {
				// record invalid entries for later warning (do not error)
				invalid = append(invalid, part)
			}
		}
	}

	if len(invalid) > 0 {
		log.Printf("[WARN] parseClientIps: parsed valid IPs: %v; ignored invalid client IP entries: %v", clientIps, invalid)
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
	var xForwardedFor = flag.String("X-Forwarded-For", "", "Client IP chain from X-Forwarded-For header (optional)")
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

	// After parsing flags, treat remaining args as positional values when a corresponding named flag was not supplied.
	// Positional order:
	//   1) destination (required)
	//   2) command (required)
	//   3) auth-token (required)
	//   4) correlation-id (optional)
	//   5) X-Forwarded-For (optional)

	pos := flagSet.Args()
	index := 0

	takePos := func(named *string) {
		if *named == "" && index < len(pos) {
			*named = pos[index]
			index++
		}
	}

	takePos(destination)
	takePos(command)
	takePos(authorization)
	takePos(correlationId)
	takePos(xForwardedFor)

	// Now validation for required params

	if *destination == "" {
		return ParsedArgs{}, fmt.Errorf("--destination is required (or provide as 1st positional)")
	}
	if *command == "" {
		return ParsedArgs{}, fmt.Errorf("--command is required (or provide as 2nd positional)")
	}
	if *authorization == "" {
		return ParsedArgs{}, fmt.Errorf("--authorization is required (or provide as 3rd positional)")
	}

	clientIps := *xForwardedFor

	return ParsedArgs{
		Destination:   *destination,
		Command:       *command,
		AuthHeader:    *authorization,
		ClientIps:     parseClientIps(clientIps),
		CorrelationId: *correlationId,
	}, nil
}
