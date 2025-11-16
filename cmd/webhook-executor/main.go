package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	// Azure Key Vault URL
	keyVaultURL = os.Getenv("WEBHOOK_KEYVAULT_URL")
	// Name of the secret storing JWT signing key
	secretName = os.Getenv("WEBHOOK_SECRET_NAME")
	// Cached secret
	jwtSecret []byte
)

// Response JSON
type Response struct {
	Stdout string `json:"stdout,omitempty"`
	Stderr string `json:"stderr,omitempty"`
	Error  string `json:"error,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		outputJSON(Response{Error: "no command provided"})
		os.Exit(1)
	}

	targetCmd := os.Args[1]
	args := os.Args[2:]

	// Optional: check if last argument is JWT
	var authHeader string
	if len(args) > 0 && strings.HasPrefix(args[len(args)-1], "Bearer ") {
		authHeader = args[len(args)-1]
		args = args[:len(args)-1] // remove from args
	}

	// Fetch JWT secret from Azure Key Vault (once)
	if len(jwtSecret) == 0 && authHeader != "" {
		var err error
		jwtSecret, err = fetchSecretFromKeyVault(keyVaultURL, secretName)
		if err != nil {
			outputJSON(Response{Error: fmt.Sprintf("failed to fetch JWT secret: %v", err)})
			os.Exit(1)
		}
	}

	// Validate JWT if present
	if authHeader != "" {
		if !validateJWT(authHeader) {
			outputJSON(Response{Error: "invalid JWT"})
			os.Exit(1)
		}
	}

	// Execute the target command
	cmd := exec.Command(targetCmd, args...)
	stdout, err := cmd.Output()

	resp := Response{}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			resp.Stderr = string(exitErr.Stderr)
			resp.Error = string(exitErr.Stderr)
		} else {
			resp.Error = err.Error()
		}
		outputJSON(resp)
		os.Exit(1)
	}

	resp.Stdout = string(stdout)
	outputJSON(resp)
}

// helper to output JSON
func outputJSON(resp Response) {
	b, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("failed to marshal JSON: %v", err)
	}
	fmt.Println(string(b))
}
