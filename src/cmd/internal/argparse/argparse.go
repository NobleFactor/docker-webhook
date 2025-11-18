// Package argparse provides command line argument parsing for webhook-executor
package argparse

import (
	"flag"
	"fmt"
)

// ParsedArgs holds the parsed command line arguments
type ParsedArgs struct {
	Destination   string
	Command       string
	AuthHeader    string
	CorrelationId string
}

// ParseArguments parses command line flags and returns the values.
// Returns: ParsedArgs, error
func ParseArguments(args []string) (ParsedArgs, error) {

	var destination = flag.String("destination", "", "SSH destination (e.g., user@host or host)")
	var command = flag.String("command", "", "Command to execute on the remote host")
	var authorization = flag.String("authorization", "", "JWT token from Authorization Bearer header")
	var correlationId = flag.String("correlation-id", "", "Correlation ID for traceability (auto-generated if not provided)")
	var help = flag.Bool("help", false, "Show help message")

	flagSet := flag.NewFlagSet("webhook-executor", flag.ContinueOnError)
	flagSet.StringVar(destination, "destination", "", "SSH destination (e.g., user@host or host)")
	flagSet.StringVar(command, "command", "", "Command to execute on the remote host")
	flagSet.StringVar(authorization, "authorization", "", "JWT token from Authorization Bearer header")
	flagSet.StringVar(correlationId, "correlation-id", "", "Correlation ID for traceability (auto-generated if not provided)")
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
		CorrelationId: *correlationId,
	}, nil
}
