// SPDX-FileCopyrightText: 2016-2025 Noble Factor
// SPDX-License-Identifier: MIT
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
	keyVaultURL     = os.Getenv("WEBHOOK_KEYVAULT_URL") // Azure key vault URL
	location        = os.Getenv("WEBHOOK_LOCATION")     // Location of service deployment
	secretName      = os.Getenv("WEBHOOK_SECRET_NAME")  // Name of the secret storing JWT signing key
	configDirectory = os.Getenv("WEBHOOK_CONFIG")       // Service configuration directory (default: /usr/local/etc/webhook)
	jwtSecret       []byte                              // Cached secret retrieved from Azure key vault
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

	log.Printf("WEBHOOK_KEYVAULT_URL: %s", keyVaultURL)
	log.Printf("WEBHOOK_SECRET_NAME: %s", secretName)
	log.Printf("WEBHOOK_CONFIG: %s", os.Getenv("WEBHOOK_CONFIG"))
	log.Printf("AZURE_CLIENT_SECRET: %s", os.Getenv("AZURE_CLIENT_SECRET"))
	log.Printf("AZURE_CLIENT_ID: %s", os.Getenv("AZURE_CLIENT_ID"))
	log.Printf("AZURE_TENANT_ID: %s", os.Getenv("AZURE_TENANT_ID"))

	log.Printf("Arguments parsed successfully: destination=%s, command=%s, client-ips=%v", parsed.Destination, parsed.Command, parsed.ClientIps)

	destination := parsed.Destination
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

	if err := jwt.ValidateJWT(authHeader, string(jwtSecret), location); err != nil {
		log.Printf("[ERROR] JWT validation failed: %v", err)
		errorStr := "invalid JWT"
		outputJson(sshremote.Response{Error: &errorStr, Status: -1, Reason: "Executor Error", CorrelationId: correlationId})
		return
	}

	log.Printf("JWT validation successful") // Execute the remote command

	log.Printf("Executing remote SSH command: ssh %s %s", destination, command)

	destination, clientConfig, err := sshremote.ParseSshDestination(destination, configDirectory)
	if err != nil {
		log.Printf("[ERROR] SSH destination parsing failed: %v", err)
		errorStr := "invalid SSH destination"
		outputJson(sshremote.Response{Error: &errorStr, Status: -1, Reason: "Executor Error", CorrelationId: correlationId})
		return
	}

	response := sshremote.ExecuteRemoteCommand(destination, clientConfig, command)
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
