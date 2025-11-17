
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
  Webhook endpoint
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

3. **Remote-executor**
   - Receives command, destination, optional identity.
   - Executes over SSH with timeout.
   - Returns success/failure to runner logs.

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
