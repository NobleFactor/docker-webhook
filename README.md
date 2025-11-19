
# Docker Webhook

Deploy secure, containerized webhook services that execute commands in response to HTTP requests. This project provides production-ready Docker images with built-in authentication, configuration management, and deployment tooling to get your webhook infrastructure running quickly and reliably.

## Deployment Overview

The `webhook.config/` directory contains your deployment configuration:

- **`hooks.json`**: Defines webhook rules and command execution logic
- **`hooks.env`**: Environment variables for webhook configuration
- **`command/`**: Custom scripts executed by webhooks
- **`ssl-certificates/`**: TLS certificates for HTTPS
- **`ssh/`**: SSH keys for secure operations

### Deployment Workflow

1. **Generate Location**: Run `make Prepare-WebhookDeployment` to create compose files and certificates
2. **Configure Hooks**: Create `webhook.config/$(LOCATION)/hooks.json` with your webhook rules
3. **Set Environment**: Define location-specific settings in `webhook-$(LOCATION).env`
4. **Build Container**: Execute `make New-WebhookContainer` to prepare the deployment
5. **Start Service**: Use `make Start-Webhook` to launch the webhook service

For secure deployments with JWT-based HMAC authentication using Azure Key Vault:

1. Create an Azure service principal with Reader role: `az ad sp create-for-rbac --name "webhook-executor-sp" --role reader --scopes /subscriptions/<subscription-id>`
2. Set up authentication credentials for Key Vault access: `make New-WebhookAzureAuth LOCATION=<location>`
3. Generate and store JWT token in Key Vault: `make New-WebhookExecutorToken`
4. Start the service with client secret: `make Start-Webhook AZURE_CLIENT_SECRET="<secret>"`

## Features

- **HTTP Webhook Server**: Receive and process webhook requests with configurable rules
- **Command Execution**: Run shell commands or scripts in response to webhooks
- **HMAC Authentication**: Secure webhook validation using JWT tokens retrieved from Azure Key Vault, with Azure service principal authentication for Key Vault access
- **Docker Ready**: Pre-built images with s6-overlay for reliable container lifecycle management
- **Hot Reload**: Automatically reload hook configurations without restarting
- **TLS Support**: Optional HTTPS with custom certificates
- **JSON Configuration**: Define webhook rules in JSON format
- **Multi-Platform**: Supports AMD64 and ARM64 architectures
- **Deployment Tooling**: Makefile targets for complete deployment lifecycle management

## Quick Start

Deploy a webhook service in seconds:

```bash
# Run with default configuration
docker run -d -p 9000:9000 --name webhook noblefactor/docker-webhook

# Access the service
curl http://localhost:9000/hooks/my-hook
```

For a more complete setup with custom hooks:

```bash
# Create a hooks configuration
cat > hooks.json << 'EOF'
[
  {
    "id": "hello",
    "execute-command": "/bin/echo",
    "command-working-directory": "/",
    "pass-arguments-to-command": [
      {"source": "string", "name": "Hello World!"}
    ]
  }
]
EOF

# Run with mounted configuration
docker run -d -p 9000:9000 \
  -v $(pwd)/hooks.json:/etc/webhook/hooks.json \
  --name webhook \
  noblefactor/docker-webhook \
  -hooks=/etc/webhook/hooks.json \
  -verbose
```

## Configuration

### Prepare-WebhookDeployment (brief)

- **Usage:** `Prepare-WebhookDeployment [--env-file <file>] [--role <role> --location <location> ...]`
- **Outputs:** `ROLE-<LOCATION>.env`, `ROLE-<LOCATION>.yaml`, `<ROLE>.config/<LOCATION>/ssl-certificates/certificate-request.conf`
- **Requires:** `bash`, `envsubst` (from `gettext`)

Example env-file (`examples/us-wa.env`):

```env
ROLE=webhook
LOCATION=us-wa
DOMAIN_NAME=example.com
EMAIL_ADDRESS=ops@example.com
```

Invocation:

```bash
Prepare-WebhookDeployment --env-file ./examples/us-wa.env
```

### Hook Configuration

Webhooks are defined in a JSON file (typically `hooks.json`). Each hook specifies:

- **id**: Unique identifier for the hook
- **execute-command**: Command to run when triggered
- **trigger-rule**: Conditions for triggering (method, headers, payload)
- **pass-arguments-to-command**: How to pass request data to the command

Example `hooks.json`:

```json
[
  {
    "id": "deploy",
    "execute-command": "/usr/local/bin/deploy.sh",
    "command-working-directory": "/app",
    "trigger-rule": {
      "match": {
        "type": "value",
        "value": "production",
        "parameter": {
          "source": "payload",
          "name": "environment"
        }
      }
    },
    "pass-arguments-to-command": [
      {
        "source": "payload",
        "name": "branch"
      }
    ]
  }
]
```

### webhook-executor API Reference

This section documents the API for hooks configured with `webhook-executor` as the `execute-command` in `hooks.json`. These hooks enable secure command execution with HMAC authentication using JWT tokens from Azure Key Vault. Note that this API format is specific to `webhook-executor` calls and does not apply to other webhook configurations.

#### Request Format

Webhook-executor endpoints are triggered via HTTP GET requests to `/hooks/{hook-name}`, where `{hook-name}` matches a hook defined in `hooks.json` with `"execute-command": "webhook-executor"`. The request supports the following query parameters:

- `hostname` (required): The target hostname for command execution
- `command` (required): The command to execute on the target host

Example request:

```http
GET /hooks/remote-mac?hostname=example.com&command=uptime
```

#### Response Schema

All webhook-executor responses are returned as JSON objects with the following structure:

- `exit_code` (integer, required): The exit code of the executed command (0 for success, non-zero for failure)
- `reason` (string, required): A description of the command execution result
- `error` (string or null, optional): Error details if the command failed, or `null` if successful
- `stdout` (string, optional): The standard output from the executed command
- `stderr` (string, optional): The standard error output from the executed command

Example successful response:

```json
{
  "exit_code": 0,
  "reason": "Command executed successfully",
  "error": null,
  "stdout": " 14:32:15 up  5:23,  1 user,  load average: 0.00, 0.00, 0.00\n",
  "stderr": ""
}
```

Example error response:

```json
{
  "exit_code": 1,
  "reason": "Command failed",
  "error": "uptime: command not found",
  "stdout": "",
  "stderr": "bash: uptime: command not found\n"
}
```

A test script `test/Test-WebhookExecutor` is provided for validating webhook-executor API responses.

### Environment Variables

- `WEBHOOK_PORT`: Port to listen on (default: 9000)
- `WEBHOOK_TLS_CERT`: Path to TLS certificate file
- `WEBHOOK_TLS_KEY`: Path to TLS private key file

### Command Line Options

- `-hooks`: Path to hooks JSON file
- `-hotreload`: Reload hooks on file change
- `-verbose`: Enable verbose logging
- `-port`: Override default port

## Deployment

### Docker Compose

```yaml
version: '3.8'
services:
  webhook:
    image: noblefactor/docker-webhook:latest
    ports:
      - "9000:9000"
    volumes:
      - ./hooks.json:/etc/webhook/hooks.json:ro
      - ./scripts:/app/scripts:ro
    environment:
      - WEBHOOK_PORT=9000
    restart: unless-stopped
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webhook
spec:
  replicas: 1
  selector:
    matchLabels:
      app: webhook
  template:
    metadata:
      labels:
        app: webhook
    spec:
      containers:
      - name: webhook
        image: noblefactor/docker-webhook:latest
        ports:
        - containerPort: 9000
        volumeMounts:
        - name: hooks-config
          mountPath: /etc/webhook
        env:
        - name: WEBHOOK_PORT
          value: "9000"
      volumes:
      - name: hooks-config
        configMap:
          name: webhook-hooks
```

### Security: JWT Authentication

For secure webhook validation, use HMAC with JWT tokens:

1. Generate a JWT token:

   ```bash
   docker run --rm noblefactor/docker-webhook /bin/New-JsonWebToken --secret "your-secret" --algorithm HS512
   ```

2. Store the token in Azure Key Vault

3. Configure hooks with HMAC validation:

   ```json
   {
     "id": "secure-hook",
     "execute-command": "/bin/echo",
     "trigger-rule-mismatch-http-response-code": 401,
     "trigger-rule": {
       "and": [
         {
           "match": {
             "type": "value",
             "value": "your-expected-value",
             "parameter": {
               "source": "header",
               "name": "X-Hub-Signature-256"
             }
           }
         }
       ]
     }
   }
   ```

## Examples

### GitHub Webhook

Trigger deployments on push events:

```json
[
  {
    "id": "github-deploy",
    "execute-command": "/app/deploy.sh",
    "trigger-rule": {
      "match": {
        "type": "value",
        "value": "refs/heads/main",
        "parameter": {
          "source": "payload",
          "name": "ref"
        }
      }
    },
    "pass-arguments-to-command": [
      {
        "source": "payload",
        "name": "after"
      }
    ]
  }
]
```

### Slack Integration

Respond to Slack slash commands:

```json
[
  {
    "id": "slack-command",
    "execute-command": "/app/slack-handler.sh",
    "trigger-rule": {
      "match": {
        "type": "value",
        "value": "/mycommand",
        "parameter": {
          "source": "payload",
          "name": "command"
        }
      }
    }
  }
]
```

## Building from Source

```bash
# Clone the repository
git clone https://github.com/NobleFactor/docker-webhook.git
cd docker-webhook

# Build the image
make New-WebhookImage

# Run tests
make test
```

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

For major changes, please open an issue first to discuss the proposed changes.

## Prior Art and Attribution

This project is a fork and extension of:

- **adnanh/webhook**: The core webhook server implementation by Adnan Hajdarević. Licensed under Apache 2.0. Source: [https://github.com/adnanh/webhook](https://github.com/adnanh/webhook)
- **almir/webhook**: The original Docker image by Almir Sarajčić. Source: [https://hub.docker.com/r/almir/webhook/](https://hub.docker.com/r/almir/webhook/)

This fork adds Noble Factor-specific improvements including s6-overlay for process management, Azure Key Vault integration for JWT-based HMAC authentication, and enhanced tooling for development and deployment.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE) for details.

The core webhook functionality is from [adnanh/webhook](https://github.com/adnanh/webhook/) (Apache 2.0 License).
