// SPDX-FileCopyrightText: 2016-2025 Noble Factor
// SPDX-License-Identifier: MIT
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "os"
    "net/url"
    "strings"
    "time"

    "github.com/NobleFactor/docker-webhook/cmd/internal/argparse"
    "github.com/NobleFactor/docker-webhook/cmd/internal/azure"
    "github.com/NobleFactor/docker-webhook/cmd/internal/jwt"
    "github.com/NobleFactor/docker-webhook/cmd/internal/sshremote"
    "github.com/google/uuid"
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

    // Validate environment early

    keyVaultURL, err := getKeyVaultURL()
    if err != nil {
        message := err.Error()
        log.Printf("[ERROR] %s", message)
        outputJson(sshremote.Response{Error: &message, Status: -1, Reason: "Executor Error", CorrelationId: correlationId})
        return
    }

    secretName, err := getSecretName()
    if err != nil {
        message := err.Error()
        log.Printf("[ERROR] %s", message)
        outputJson(sshremote.Response{Error: &message, Status: -1, Reason: "Executor Error", CorrelationId: correlationId})
        return
    }

    configDirectory, err := getConfigDirectory()
    if err != nil {
        message := err.Error()
        log.Printf("[ERROR] %s", message)
        outputJson(sshremote.Response{Error: &message, Status: -1, Reason: "Executor Error", CorrelationId: correlationId})
        return
    }

    location, _ := getLocation()

    tokenTtl, err := getTokenTtl()
    if err != nil {
        message := err.Error()
        log.Printf("[ERROR] %s", message)
        outputJson(sshremote.Response{Error: &message, Status: -1, Reason: "Executor Error", CorrelationId: correlationId})
        return
    }

    tokenRefreshWindow, err := getTokenRefreshWindow()
    if err != nil {
        message := err.Error()
        log.Printf("[ERROR] %s", message)
        outputJson(sshremote.Response{Error: &message, Status: -1, Reason: "Executor Error", CorrelationId: correlationId})
        return
    }

    log.Printf("WEBHOOK_KEYVAULT_URL         : %s", keyVaultURL)
    log.Printf("WEBHOOK_SECRET_NAME          : %s", secretName)
    log.Printf("WEBHOOK_CONFIG               : %s", configDirectory)
    log.Printf("WEBHOOK_LOCATION             : %s", location)
    log.Printf("WEBHOOK_TOKEN_TTL            : %s", tokenTtl)
    log.Printf("WEBHOOK_TOKEN_REFRESH_WINDOW : %s", tokenRefreshWindow)

    destination := parsed.Destination
    command := parsed.Command
    authHeader := parsed.AuthHeader // Fetch JWT secret from Azure Key Vault (once)

    var jwtSecret []byte

    if authHeader != "" {
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

    // Validate JWT (required) — returns parsed token for reuse by refresh

    tokenStr, parsedToken, err := jwt.ValidateJWT(authHeader, string(jwtSecret), location)
    if err != nil {
        log.Printf("[ERROR] JWT validation failed: %v", err)
        errorStr := "invalid JWT"
        outputJson(sshremote.Response{Error: &errorStr, Status: -1, Reason: "Executor Error", CorrelationId: correlationId})
        return
    }

    log.Printf("JWT validation successful")

    // Attempt to refresh the token. Field name in response: `authToken` (string)
    
    var refreshedToken string
    
    if authHeader != "" && len(jwtSecret) > 0 {
        // Refresh if token is within configured window; new TTL = configured value
        newTok, refreshed, err := jwt.RefreshJWT(parsedToken, tokenStr, string(jwtSecret), location, tokenRefreshWindow, tokenTtl)
        if err != nil {
            log.Printf("[WARN] token refresh attempt failed: %v", err)
        } else {
            // always capture the token returned (either refreshed or the original)
            refreshedToken = newTok
            if refreshed {
                log.Printf("JWT was refreshed for subject %s", location)
            }
        }
    }

    // Execute the remote command

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
    if refreshedToken != "" {
        response.AuthToken = &refreshedToken
    }
    log.Printf("Remote SSH command execution completed")
    outputJson(response)
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Environment validators
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Validates the value of WEBHOOK_KEYVAULT_URL
func getKeyVaultURL() (string, error) {
    u := getenvOrDefault("WEBHOOK_KEYVAULT_URL", "")
    if strings.TrimSpace(u) == "" {
        return "", fmt.Errorf("WEBHOOK_KEYVAULT_URL is required")
    }
    if p, err := url.Parse(u); err != nil || p.Scheme != "https" || !strings.HasSuffix(p.Host, ".vault.azure.net") {
        return "", fmt.Errorf("invalid WEBHOOK_KEYVAULT_URL: %s", u)
    }
    return u, nil
}

// Validates the value of WEBHOOK_SECRET_NAME
func getSecretName() (string, error) {
    s := getenvOrDefault("WEBHOOK_SECRET_NAME", "")
    if strings.TrimSpace(s) == "" {
        return "", fmt.Errorf("WEBHOOK_SECRET_NAME is required")
    }
    return s, nil
}

// Validates the value of WEBHOOK_CONFIG
func getConfigDirectory() (string, error) {
    p := getenvOrDefault("WEBHOOK_CONFIG", "")
    if strings.TrimSpace(p) == "" {
        return "", fmt.Errorf("WEBHOOK_CONFIG is required")
    }
    if fi, err := os.Stat(p); err != nil || !fi.IsDir() {
        return "", fmt.Errorf("WEBHOOK_CONFIG does not exist or is not a directory: %s", p)
    }
    return p, nil
}

// Validates the value of WEBHOOK_LOCATION
func getLocation() (string, error) { 
    return getenvOrDefault("WEBHOOK_LOCATION", ""), nil 
}

// Validates the value of WEBHOOK_TOKEN_TTL
func getTokenTtl() (time.Duration, error) { 
    return parseDurationEnv("WEBHOOK_TOKEN_TTL", "24h") 
}

// Validates the value of WEBHOOK_TOKEN_REFRESH_WINDOW
func getTokenRefreshWindow() (time.Duration, error) { 
    return parseDurationEnv("WEBHOOK_TOKEN_REFRESH_WINDOW", "5m") 
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// HELPERS
////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Get the value of an environment variable or a default if it's blank
func getenvOrDefault(key, def string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return def
}

// Convert an SSH remote response to its JSON representation
func outputJson(resp sshremote.Response) {
    b, err := json.Marshal(resp)
    if err != nil {
        log.Fatalf("failed to marshal JSON: %v", err)
    }
    fmt.Println(string(b))
}

func parseDurationEnv(key, def string) (time.Duration, error) {
    s := getenvOrDefault(key, def)
    d, err := time.ParseDuration(s)
    if err != nil {
        return 0, fmt.Errorf("invalid %s: %v", key, err)
    }
    return d, nil
}
