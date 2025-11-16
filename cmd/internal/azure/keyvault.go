package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
)

// fetchSecretFromKeyVault retrieves the secret value from Azure Key Vault
func fetchSecretFromKeyVault(vaultUrl, secretName string) ([]byte, error) {
	if vaultUrl == "" || secretName == "" {
		return nil, fmt.Errorf("WEBHOOK_KEYVAULT_URL or WEBHOOK_SECRET_NAME not set")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure credential: %w", err)
	}

	client, err := azsecrets.NewClient(vaultUrl, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Key Vault client: %w", err)
	}

	resp, err := client.GetSecret(context.Background(), secretName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	return []byte(*resp.Value), nil
}
