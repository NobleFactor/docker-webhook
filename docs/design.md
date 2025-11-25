
# docker-webhook — Requirements & Design Document

## Document Status

| Attribute | Value |
|-----------|-------|
| **Version** | 1.0-draft |
| **Status** | Draft - Pending Review |
| **Last Updated** | November 16, 2025 |
| **Authors** | David Noble |
| **Reviewers** | TBD |
| **Approval Date** | TBD |

## 1. Purpose

Provide a reliable, secure, and scalable architecture for running remote-executor jobs triggered by webhooks, using:

- [Hookdec](https://hookdeck.com/docs) as the resilient ingress layer, and
- [Webhook] as the runner for local jobs.

## 2. System Components

| Component | Role |
|-----------|------|
| Webhook Source (GitHub, DockerHub, etc.) | Sends webhook events to Hookdeck |
| Hookdeck | Webhook proxy, queueing, retries, HMAC signature forwarding |
| Webhook | Receives verified webhooks, validates JWT, executes remote jobs synchronously |
| Remote-executor | Executes SSH commands on destination hosts |
| Monitoring / Logging | Health, metrics, error logging, event replay support |

## 3. High-Level Architecture

\`\`\`
Webhook Source
      │
      ▼
  Hookdeck (proxy)
      │ HMAC signed forwarding
      ├───► Other Endpoints
      │     (monitoring, logging, databases, external service providers, etc.)
      ▼
  Webhook HTTP Endpoint
      │
      ▼
  webhook-executor (subprocess)
      │
      ▼
  SSH command → target host
\`\`\`

Notes:

- Hookdeck handles retries, rate-limiting, and backpressure.  
- Webhook verifies Hookdeck signature **and JWT** for each job request.  
- Jobs are executed synchronously by the webhook instance.

## 4. Functional Requirements

1. **Webhook reception**
   - Hookdeck receives webhooks from multiple providers.
   - Retains events in case of delivery failures.
   - Rate-limits per destination.

2. **Security**
   - Verify Hookdeck HMAC signature.
   - JWT verification for authorized jobs:
     - Validate `iss`, `exp`, `sub`, and claims.
     - Secret/key management via environment variable or secret store.
   - Optional Idempotency check using `Idempotency-Key` or event ID.

3. **Job Execution**
   - Webhook executes remote-executor commands synchronously via SSH.
   - Apply timeouts to prevent hanging executions.

4. **Operational**
   - Healthcheck endpoint (`/healthz`) for runner.
   - Logging of received webhooks, verification status, execution results.
   - Metrics export (Prometheus or similar).

## 5. Non-Functional Requirements
- **Reliability:** retries, dead-letter queue, durable persistence.
- **Security:** JWT + HMAC, secure secrets handling.
- **Scalability:** Hookdeck queues allow smooth scaling by deploying multiple webhook instances.
- **Observability:** logging, monitoring, replay capability.
- **Maintainability:** containerized job runner for CI/CD deployment.

## 6. Implementation Notes

1. **Hookdeck**
   - Use CLI tunnel in development: `hookdeck listen 8080 my-source`
   - In production, configure public endpoint with signature verification.

2. **Job Runner Container**
   - HTTP endpoint `/hooks/github`
   - JWT verification for incoming jobs.
   - Executes jobs synchronously using remote-executor.

3. **Remote-executor (webhook-executor)**
   - Go-based binary implementing native SSH execution with golang.org/x/crypto/ssh
   - Structured JSON responses with status codes (0-255), HTTP-style reason phrases, correlation IDs, and error details
   - Correlation ID generation (UUID v4) for end-to-end request traceability
   - Differentiated error handling: executor failures, SSH connection failures, remote command failures
   - JWT validation against Azure Key Vault secrets
   - Comprehensive argument parsing with auto-generated correlation IDs
   - UUID-prefixed logging for distributed tracing

4. **Secrets**
   - Hookdeck signing secret (HMAC)
   - JWT secret for signing/verifying jobs
   - SSH private keys (if needed) secured inside container or secret store.

## 7. Sample Flow

1. GitHub pushes code → webhook sent → Hookdeck receives.
2. Hookdeck queues and forwards event (with HMAC signature) → Job Runner HTTP endpoint.
3. Job Runner verifies Hookdeck signature + JWT.
4. Job Runner executes remote-executor synchronously.
5. Job completes → logs result, reports metrics.

## 8. Checklist Before Deployment

- [ ] Hookdeck signature verification implemented
- [ ] JWT validation implemented
- [ ] Idempotency / dedupe logic
- [ ] Remote-executor timeout handling
- [ ] Logging & monitoring integrated
- [ ] Health checks configured

## 9. Webhook-Executor Design

### Overview

The webhook-executor is a Go-based command-line tool that executes remote commands via SSH and returns structured JSON responses. It provides reliable, traceable execution with comprehensive error handling and security validation.

### Architecture

```text
CLI Arguments → Argument Parsing → JWT Validation → SSH Execution → Response Generation
     │                │                 │             │             │
     └─ correlation-id └─ Azure Key Vault └─ golang.org/x/crypto/ssh └─ JSON Output
```

### Key Components

#### Argument Parsing

- Supports command-line flags: `--destination`, `--command`, `--jwt`, `--correlation-id`, `--X-Forwarded-For`
- Auto-generates UUID v4 correlation IDs if not provided
- Validates required parameters
- `--X-Forwarded-For`: Client IP chain from X-Forwarded-For header for security logging, parses comma-separated IPs and validates each

#### JWT Validation

- Validates JWT tokens against Azure Key Vault secrets
- Supports HS256, HS384, HS512 algorithms
- Extracts claims for authorization

#### SSH Execution

- Native SSH implementation using golang.org/x/crypto/ssh
- Supports password and key-based authentication
- Executes commands synchronously with timeout handling
- Captures stdout, stderr, and exit codes

#### Response Structure

Returns JSON with consistent schema:

```json
{
   "status": 0,
   "reason": "OK",
   "stdout": "",
   "stderr": "",
   "error": null,
   "authToken": null,
   "correlationId": "uuid-v4-string"
}
```

#### Error Handling

- **Status 0-125**: Command execution results
- **Status 126**: Command found but not executable
- **Status 127**: Command not found
- **Status 200+**: Executor/SSH errors (e.g., 200: SSH connection failed)
- Reason phrases provide HTTP-style descriptions
- Correlation IDs enable request tracing across logs

#### Logging

- UUID-prefixed log messages for traceability
- Structured logging with correlation ID context
- Error details logged with full context
  
Note: webhook-executor writes diagnostic logs to the container logging pipeline when possible (it attempts to open /proc/1/fd/2 and send logs there; if unavailable it falls back to the process' stderr). This keeps executor diagnostic output out of any command-capture/response bodies returned by a caller (e.g. the webhook server) while still preserving logs in the container's log collection (s6, journald, or docker logs).

### Security

- JWT-based authentication with Azure Key Vault integration
- Secure SSH connections with proper key management
- No sensitive data in logs or responses

### Usage Example

```bash
./webhook-executor --destination user@host --command "echo hello" --jwt eyJ0eXAi... --X-Forwarded-For "203.0.113.1,198.51.100.1"
```

**Webhook Configuration**: The webhook passes client IP chain via `X-Forwarded-For` header:

```json
{
  "source": "header",
  "name": "X-Forwarded-For"
}
```

This passes `--X-Forwarded-For "<ip-chain>"` which webhook-executor parses into validated net.IP array for audit trails, extracting the first valid IP as the client IP.

Output:

```json
{
   "status": 0,
   "reason": "OK",
   "stdout": "",
   "stderr": "",
   "error": null,
   "authToken": null,
   "correlationId": "550e8400-e29b-41d4-a716-446655440000"
}
```
