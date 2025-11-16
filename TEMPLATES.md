# TEMPLATES.md

This document describes the two template files used in the `docker-webhook` project for environment and certificate configuration.

---

## 1. `services.yaml.template`

**Location:** `${project_root}/services.yaml.template`

**Purpose:**

- Serves as the base template for generating the Docker Compose YAML file for each deployment location.
- Contains placeholders for environment-specific variables (such as location, ports, secrets, and service configuration).
- Used by Makefile targets to produce a customized `webhook-<LOCATION>.yaml` file for each deployment.

**Key Placeholders:**

For `services.yaml.template`:

- `${LOCATION}`: Deployment region or environment.
- `${ROLE}`: Service role (e.g., webhook).

**Usage:**

- The Makefile and supporting scripts substitute these placeholders with actual values to generate a valid Docker Compose file for the target environment.

---

## 2. `certificate-request.conf.template`

**Location:** `${project_root}/ssl-secrets/certificates/certificate-request.conf.template`

**Purpose:**

- Provides a template for generating the OpenSSL certificate request configuration file.
- Contains placeholders for subject information, key usage, and other certificate parameters.
- Used to create `certificate-request.conf` for each deployment location, which is then used to generate SSL certificates.

**Key Placeholders:**

For `certificate-request.conf.template`:

- `${ROLE}`: Service role (used in CN and SAN).
- `${LOCATION}`: Deployment region or environment (used in CN and SAN).
- `${DOMAIN_NAME}`: Domain name for certificate subject.
- `${COUNTRY_CODE}`: Country code for certificate subject.
- `${STATE_OR_PROVINCE}`: State or province for certificate subject.
- `${CITY}`: City or locality for certificate subject.
- `${ORGANIZATION_NAME}`: Organization name for certificate subject.
- `${ORGANIZATIONAL_UNIT}`: Organizational unit for certificate subject.
- `${EMAIL_ADDRESS}`: Email address for certificate subject.

**Usage:**

- The Makefile and scripts fill in these placeholders to produce a location-specific `certificate-request.conf` file, which is then used by OpenSSL to generate self-signed certificates for the deployment.

---

## Template Workflow

1. The Makefile ensures that the required `.template` files exist in the project root.
2. When a new location is provisioned, the Makefile and scripts substitute environment variables into the templates to generate the necessary configuration files.
3. These generated files are then used to build containers and create SSL certificates for secure operation.

---

**Note:**

- Do not edit generated files directly; always update the corresponding `.template` file and regenerate as needed.
- Placeholders must match the variable names used in the Makefile and scripts for correct substitution.
