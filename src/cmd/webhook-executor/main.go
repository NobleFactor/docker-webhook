package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/NobleFactor/docker-webhook/cmd/internal/argparse"
	"github.com/NobleFactor/docker-webhook/cmd/internal/azure"
	"github.com/NobleFactor/docker-webhook/cmd/internal/jwt"
	"github.com/NobleFactor/docker-webhook/cmd/internal/sshremote"
	"github.com/google/uuid"
)

var (
	keyVaultURL = os.Getenv("WEBHOOK_KEYVAULT_URL") // Azure key vault URL
	secretName  = os.Getenv("WEBHOOK_SECRET_NAME")  // Name of the secret storing JWT signing key
	location    = os.Getenv("WEBHOOK_LOCATION")     // Expected location for JWT subject
	jwtSecret   []byte                              // Cached secret retrieved from Azure key vault
)

func main() {
	log.SetPrefix("[webhook-executor] ")

	parsed, err := argparse.ParseArguments(os.Args[1:])
	if err != nil {
		log.Printf("[ERROR] %v", err)
		errorStr := err.Error()
		outputJson(sshremote.Response{Error: &errorStr, Status: -1, Reason: "Executor Error", CorrelationId: ""})
		return
	}

	correlationId := parsed.CorrelationId
	if correlationId == "" {
		correlationId = uuid.New().String()
	}

	log.SetPrefix(fmt.Sprintf("[%s] ", correlationId))

	log.Printf("Arguments parsed successfully: destination=%s, command=%s, client-ips=%v", parsed.Destination, parsed.Command, parsed.ClientIps)

	targetCmd := parsed.Destination
	command := parsed.Command
	authHeader := parsed.AuthHeader // Fetch JWT secret from Azure Key Vault (once)

	if len(jwtSecret) == 0 && authHeader != "" {
		var err error
		jwtSecret, err = azure.FetchSecretFromKeyVault(keyVaultURL, secretName)
		if err != nil {
			log.Printf("[ERROR] Failed to fetch JWT secret from Azure Key Vault: %v", err)
			errorStr := fmt.Sprintf("failed to fetch JWT secret: %v", err)
			outputJson(sshremote.Response{Error: &errorStr, Status: -1, Reason: "Executor Error", CorrelationId: correlationId})
			return
		}
		log.Printf("JWT secret fetched successfully from Azure Key Vault")
	}

	// Validate JWT (required)

	if !jwt.ValidateJWT(authHeader, string(jwtSecret), location) {
		prefix := authHeader
		if len(authHeader) > 10 {
			prefix = authHeader[:10]
		}
		log.Printf("[ERROR] JWT validation failed for token starting with: %s...", prefix)
		errorStr := "invalid JWT"
		outputJson(sshremote.Response{Error: &errorStr, Status: -1, Reason: "Executor Error", CorrelationId: correlationId})
		return
	}

	log.Printf("JWT validation successful") // Execute the remote command

	log.Printf("Executing remote SSH command: ssh %s %s", targetCmd, command)
	response := sshremote.ExecuteRemoteCommand(targetCmd, command)
	response.CorrelationId = correlationId
	log.Printf("Remote SSH command execution completed")
	outputJson(response)
}

// helper to output JSON
func outputJson(resp sshremote.Response) {
	b, err := json.Marshal(resp)
	if err != nil {
		log.Fatalf("failed to marshal JSON: %v", err)
	}
	fmt.Println(string(b))
}
