########################################################################################################################
# SPDX-FileCopyrightText: 2016-2025 Noble Factor
# SPDX-License-Identifier: MIT
########################################################################################################################

# docker-webhook Makefile

SHELL := bash
.SHELLFLAGS := -o errexit -o nounset -o pipefail -c
.ONESHELL:

## PARAMETERS

### LOCATION

LOCATION ?=

ifeq ($(strip $(LOCATION)),)
    LOCATION := $(shell curl --fail --silent "http://ip-api.com/json?fields=countryCode,region" | jq --raw-output '"\(.countryCode)-\(.region)"' | tr '[:upper:]' '[:lower:]')
else
    LOCATION := $(shell echo $(LOCATION) | tr '[:upper:]' '[:lower:]')
endif

### CONTAINER_DOMAIN_NAME

CONTAINER_DOMAIN_NAME ?= localdomain

### CONTAINER_ENVIRONMENT

CONTAINER_ENVIRONMENT ?= dev

ifeq ($(CONTAINER_ENVIRONMENT),prod)
	undefine hostname_suffix
else
	override hostname_suffix := -$(CONTAINER_ENVIRONMENT)
endif

### CONTAINER_HOSTNAME

override CONTAINER_HOSTNAME := webhook-$(LOCATION)$(hostname_suffix)

### WEBHOOK_EXECUTOR_* Variables

WEBHOOK_EXECUTOR_DESTINATION ?=
WEBHOOK_EXECUTOR_EXCLUDE     ?= false
WEBHOOK_JWT_ALGORITHM        ?= HS512
WEBHOOK_JWT_PAYLOAD          ?= {"iss":"webhook-executor","sub":"$(LOCATION)","exp":$(shell date -d "+1 hour" +%s)}
WEBHOOK_KEYVAULT_URL         ?=
WEBHOOK_SECRET_NAME          ?= webhook-executor-$(LOCATION)-secret
WEBHOOK_TOKEN_NAME           ?= webhook-executor-$(LOCATION)-token

### Azure Authentication Variables

AZURE_CLIENT_SECRET ?=
AZURE_CLIENT_ID     ?=
AZURE_TENANT_ID     ?=

### IP_ADDRESS Optional; if absent docker compose will decide based on the IP_RANGE

IP_ADDRESS ?=

### IP_RANGE (Required if the docker network driver is macvlan; unused otherwise.)

IP_RANGE ?=

### S6_OVERLAY_VERSION

S6_OVERLAY_VERSION ?= 3.2.1.0

### WEBHOOK_EXECUTOR

WEBHOOK_EXECUTOR ?=

### WEBHOOK_PGID

WEBHOOK_PGID ?= 0

### WEBHOOK_PORT

WEBHOOK_PORT ?= 9000

### WEBHOOK_PUID

WEBHOOK_PUID ?= $(shell id --user)

### WEBHOOK_URL

WEBHOOK_URL ?= https://localhost:$(WEBHOOK_PORT)/hooks/

### WEBHOOK_VERSION

WEBHOOK_VERSION ?= 2.8.2

## VARIABLES

### PROJECT

PLATFORM ?= linux/$(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')

ifeq (,$(findstring ,, $(PLATFORM)))
    # Single-platform -> use --load so the image is loaded into local daemon
    BUILDX_LOAD := --load
    BUILDX_PUSH :=
else
    # Multi-platform -> must push to remote registry (or remove --load)
    BUILDX_LOAD :=
    BUILDX_PUSH := --push
endif

TAG ?= 1.0.0-preview.1

override project_name := webhook
override project_root := $(patsubst %/,%,$(dir $(realpath $(lastword $(MAKEFILE_LIST)))))
override project_file := $(project_root)/$(project_name)-$(LOCATION).yaml
override project_networks_file := $(project_root)/$(project_name).networks.yaml

# Compose file may not exist initially; targets below will ensure generation when needed.

override IMAGE := noblefactor/$(project_name):$(TAG)
override ROLE := webhook

### CONFIGURATION

#### ssl certificates

override ssl_certificates_root := $(project_root)/$(ROLE).config/$(LOCATION)/ssl-certificates

override ssl_certificates := \
	$(ssl_certificates_root)/private-key.pem\
	$(ssl_certificates_root)/certificate.pem

#### ssh keys

override ssh_keys_root := $(project_root)/$(ROLE).config/$(LOCATION)/ssh

override ssh_keys := \
	$(ssh_keys_root)/id_rsa\
	$(ssh_keys_root)/id_rsa.pub

#### webhook hooks

override webhook_hooks := $(project_root)/$(ROLE).config/$(LOCATION)/hooks.json
override webhook_command := $(project_root)/$(ROLE).config/$(LOCATION)/command/remote-mac

### CONTAINER VOLUMES

override volume_root := $(project_root)/volumes/$(LOCATION)

override container_certificates := \
	$(volume_root)/certificates/private-key.pem\
	$(volume_root)/certificates/certificate.pem

override container_keys := \
	$(volume_root)/ssh/id_rsa\
	$(volume_root)/ssh/id_rsa.pub

override container_hooks := $(volume_root)/hooks.json

### NETWORK

OS := $(shell uname)

ifeq ($(OS),Linux)
    override network_device := $(shell ip route | awk '/^default via / { print $$5; exit }')
    override network_driver := macvlan
else ifeq ($(OS),Darwin)
    override network_device := $(shell scutil --dns | gawk '/if_index/ { print gensub(/[()]/, "", "g", $$4); exit }')
	override network_driver := bridge
else
    $(error Unsupported operating system: $OS)
endif

override network_name := $(shell \
    project="$(project_name)"; \
    device="$(network_device)"; \
    len=$$((15 - $${#device})); \
    echo "$${project:0:$${len}}_$${device}")

## TARGETS

docker_compose := sudo \
    LOCATION="$(LOCATION)" \
    CONTAINER_HOSTNAME="$(CONTAINER_HOSTNAME)" \
    CONTAINER_DOMAIN_NAME="$(CONTAINER_DOMAIN_NAME)" \
    NETWORK_NAME="$(network_name)" \
    WEBHOOK_IMAGE="$(IMAGE)" \
    WEBHOOK_PORT="$(WEBHOOK_PORT)" \
	PUID="$(WEBHOOK_PUID)" \
	PGID="$(WEBHOOK_PGID)" \
	AZURE_CLIENT_SECRET="$(AZURE_CLIENT_SECRET)" \
	AZURE_CLIENT_ID="$(AZURE_CLIENT_ID)" \
	AZURE_TENANT_ID="$(AZURE_TENANT_ID)" \
    docker compose -f "$(project_file)" -f "$(project_networks_file)"

HELP_COLWIDTH ?= 28

.PHONY: help help-short help-full clean format setup test \
	Get-WebhookStatus \
	New-Webhook \
	New-WebhookCertificates \
	New-WebhookContainer \
	New-WebhookImage \
	New-WebhookKeys \
	New-WebhookAzureAuth \
	New-WebhookExecutorToken \
	Prepare-WebhookDeployment \
	Restart-Webhook \
	Start-Webhook \
	Start-WebhookShell \
	Stop-Webhook \
	Test-ShellFormatting \
	Test-ShellScript \
	Test-WebhookDeploymentPreparation \
	Test-WebhookExecutorIntegration \
	Test-WebhookExecutorStandalone \
	Test-WebhookReadiness \
	Update-WebhookKeys \
	Update-WebhookCertificates \
	Update-WebhookHooks 
	Update-WebhookHooksEnv

##@ Help
help: help-short ## Show brief help (alias: help-short)

help-full: ## Show detailed usage (man page)
	@man "$(project_root)/docs/docker-webhook.1"

help-short: ## Show brief help for annotated targets
	@awk 'BEGIN {FS = ":.*##"; pad = $(HELP_COLWIDTH); print "Usage: make <target> [VAR=VALUE]"; print ""; print "Targets:"} /^[a-zA-Z0-9_.-]+:.*##/ {printf "  %-*s %s\n", pad, $$1, $$2} /^##@/ {printf "\n%s\n", substr($$0,5)}' $(MAKEFILE_LIST) | less -R

##@ Utilities

clean: ## Stop, remove network, prune unused images/containers/volumes (DANGEROUS)
	$(MAKE) Stop-Webhook
	sudo docker network rm --force $(network_name) || true
	sudo docker system prune --force --all
	sudo docker volume prune --force --all

format: ## Format shell scripts in-place using shfmt (4-space indent)
	@echo "==> Formatting shell scripts with shfmt"
	@shell_scripts=$$(find bin test webhook.config -type f -exec grep -lE '^#!(/usr/bin/env[[:space:]]+)?(sh|bash)' {} + 2>/dev/null || true);
	@if [[ -n "$$shell_scripts" ]]; then
		@shfmt -w -i 4 $$shell_scripts || { echo "shfmt failed"; exit 1; }
	else
		@echo "No shell scripts found to format"
	fi

# Setup: install dependencies and register pre-commit hooks
setup: ## Install dependencies and configure pre-commit for this repo
	@echo "==> Running bin/Install-Dependencies..."
	@./bin/Install-Dependencies
	@if ! command -v pre-commit >/dev/null; then
		@echo "==> Installing pre-commit..."
		@python3 -m pip install --user pre-commit
		@mkdir -p ~/.local/bin
		@ln -sf "$(python3 -c 'import site; print(site.USER_BASE)')/bin/pre-commit" ~/.local/bin/pre-commit
	fi
	@echo "==> Installing pre-commit hooks.."
	@if ~/.local/bin/pre-commit install >/dev/null; then
		@echo "Pre-commit hooks installed to .git/hooks/pre-commit."
	else
		@echo "Pre-commit install failed."
		@exit 1
	fi

test: ## Runs repository test scripts
	$(MAKE) Test-WebhookDeploymentPreparation
	$(MAKE) Test-WebhookReadiness
	@if [[ "$(WEBHOOK_EXECUTOR_EXCLUDE)" == "true" ]]; then
		echo "==> Skipping Test-WebhookExecutorIntegration because WEBHOOK_EXECUTOR_EXCLUDE=$(WEBHOOK_EXECUTOR_EXCLUDE)"
	else
		$(MAKE) Test-WebhookExecutorIntegration
	fi
	$(MAKE) Test-ShellFormatting
	$(MAKE) Test-ShellScript

Test-ShellFormatting: ## Check shell formatting (non-destructive)
	@echo "==> Running Test-ShellFormatting"
	@./test/Test-ShellFormatting || { echo "Test-ShellFormatting failed"; exit 1; }

Test-ShellScript: ## Run shellcheck wrapper script against shell scripts
	@echo "==> Running Test-ShellScript"
	@./test/Test-ShellScript --paths bin,test --recurse || { echo "Test-ShellScript failed"; exit 1; }

Test-WebhookExecutorIntegration: ## Run Test-WebhookExecutorIntegration script against the webhook with JWT from Key Vault and configurable commands (separated by ;)
	@hooks_env="webhook.config/$(LOCATION)/hooks.env"

	@if [[ ! -f "$$hooks_env" ]]; then
		@echo "Error: $$hooks_env does not exist. Run Update-WebhookHooksEnv or set up deployment config."
		@exit 1
	fi

	@source "$$hooks_env"

	@if [[ -z "$$WEBHOOK_KEYVAULT_URL" ]]; then
		@echo "Error: WEBHOOK_KEYVAULT_URL not set in $$hooks_env"
		@exit 1
	fi
	@if [[ -z "$$WEBHOOK_SECRET_NAME" ]]; then
		@echo "Error: WEBHOOK_SECRET_NAME not set in $$hooks_env"
		@exit 1
	fi
	@if [[ -z "$$AZURE_CLIENT_ID" ]]; then
		@echo "Error: AZURE_CLIENT_ID not set in $$hooks_env"
		@exit 1
	fi
	@if [[ -z "$$AZURE_TENANT_ID" ]]; then
		@echo "Error: AZURE_TENANT_ID not set in $$hooks_env"
		@exit 1
	fi

	@echo "==> Fetching JWT from Key Vault..."

	@jwt=$$(./bin/Get-JsonWebToken --keyvault-url "$$WEBHOOK_KEYVAULT_URL" --secret-name "$$WEBHOOK_SECRET_NAME") || {
		echo "Failed to fetch JWT"
		exit 1
	}

	@echo "==> Running Test-WebhookExecutorIntegration with commands: $(WEBHOOK_EXECUTOR_COMMAND)"

	@destination="$(WEBHOOK_EXECUTOR_DESTINATION)"

	@if [[ -z "$$destination" ]]; then 
		@destination="$(CONTAINER_HOSTNAME).$(CONTAINER_DOMAIN_NAME)"
	fi

	@IFS=';' read -ra cmds <<< "$(WEBHOOK_EXECUTOR_COMMAND)"
	@cmd_args=""
	@for cmd in "$${cmds[@]}"; do
		@cmd_args="$$cmd_args --command $$cmd"
	done

	@./test/Test-WebhookExecutorIntegration \
		--webhook-url "$(WEBHOOK_URL)$(WEBHOOK_EXECUTOR)" \
		--destination "$$destination" \
		--jwt "$$jwt" \
		$$cmd_args

Test-WebhookExecutorStandalone: ## Run Test-WebhookExecutorStandalone script directly against the binary with JWT from Key Vault and configurable commands (separated by ;)
	@hooks_env="webhook.config/$(LOCATION)/hooks.env"

	@if [[ ! -f "$$hooks_env" ]]; then
		@echo "Error: $$hooks_env does not exist. Run Update-WebhookHooksEnv or set up deployment config."
		@exit 1
	fi

	@source "$$hooks_env"

	@if [[ -z "$$WEBHOOK_KEYVAULT_URL" ]]; then
		@echo "Error: WEBHOOK_KEYVAULT_URL not set in $$hooks_env"
		@exit 1
	fi
	@if [[ -z "$$WEBHOOK_SECRET_NAME" ]]; then
		@echo "Error: WEBHOOK_SECRET_NAME not set in $$hooks_env"
		@exit 1
	fi
	@if [[ -z "$$AZURE_CLIENT_ID" ]]; then
		@echo "Error: AZURE_CLIENT_ID not set in $$hooks_env"
		@exit 1
	fi
	@if [[ -z "$$AZURE_TENANT_ID" ]]; then
		@echo "Error: AZURE_TENANT_ID not set in $$hooks_env"
		@exit 1
	fi

	@echo "==> Running Test-WebhookExecutorStandalone with commands: $(WEBHOOK_EXECUTOR_COMMAND)"

	@destination="$(WEBHOOK_EXECUTOR_DESTINATION)"

	@if [[ -z "$$destination" ]]; then 
		@destination="$(CONTAINER_HOSTNAME).$(CONTAINER_DOMAIN_NAME)"
	fi

	@IFS=';' read -ra cmds <<< "$(WEBHOOK_EXECUTOR_COMMAND)"
	@cmd_args=""
	@for cmd in "$${cmds[@]}"; do
		@cmd_args="$$cmd_args --command $$cmd"
	done

	@./test/Test-WebhookExecutorStandalone \
		--location "$(LOCATION)" \
		--token-name "$(WEBHOOK_TOKEN_NAME)" \
		--destination "$$destination" \
		$$cmd_args

Test-WebhookDeploymentPreparation:
	@echo "==> Running Test-DockerLocationGeneration"
	@./test/Test-WebhookDeploymentPreparation --location zz-xy || { echo "Test-WebhookDeploymentPreparation failed"; exit 1; }

Test-WebhookReadiness:
	@echo "==> Running Test-WebhookReadiness"
	@echo "==> Setting up test environment: make clean New-Webhook Start-Webhook"
	@$(MAKE) clean New-Webhook Start-Webhook
	@./test/Test-WebhookReadiness --container "$(CONTAINER_HOSTNAME)" --wait 2 || { echo "Test-WebhookReadiness failed"; exit 1; }

##@ Azure Authentication
New-WebhookAzureAuth: ## Initialize Azure service principal authentication for webhook-executor
	@if [[ -z "$(AZURE_CLIENT_ID)" ]]; then
		@read -p "Enter AZURE_CLIENT_ID: " AZURE_CLIENT_ID
	fi
	@if [[ -z "$(AZURE_TENANT_ID)" ]]; then
		@read -p "Enter AZURE_TENANT_ID: " AZURE_TENANT_ID
	fi
	@if [[ -z "$(AZURE_CLIENT_SECRET)" ]]; then
		@read -s -p "Enter AZURE_CLIENT_SECRET: " AZURE_CLIENT_SECRET
		@echo
	fi
	@if [[ -z "$(LOCATION)" ]]; then
		@echo "Error: LOCATION must be set (e.g., make New-WebhookAzureAuth LOCATION=us-wa)"
		@exit 1
	fi
	@hooks_env="webhook.config/$(LOCATION)/hooks.env"
	@mkdir --parents "webhook.config/$(LOCATION)"
	@if grep --quiet "^.*AZURE_CLIENT_ID=" "$$hooks_env" 2>/dev/null; then
		@sed --in-place "s|^.*AZURE_CLIENT_ID=.*|export AZURE_CLIENT_ID=$$AZURE_CLIENT_ID|" "$$hooks_env"
	else
		@echo "export AZURE_CLIENT_ID=$$AZURE_CLIENT_ID" >> "$$hooks_env"
	fi
	@if grep --quiet "^.*AZURE_TENANT_ID=" "$$hooks_env" 2>/dev/null; then
		@sed --in-place "s|^.*AZURE_TENANT_ID=.*|export AZURE_TENANT_ID=$$AZURE_TENANT_ID|" "$$hooks_env"
	else
		@echo "export AZURE_TENANT_ID=$$AZURE_TENANT_ID" >> "$$hooks_env"
	fi
	@echo "Azure authentication variables updated in $$hooks_env"
	@echo "Note: AZURE_CLIENT_SECRET must be set at container runtime for security"
	@echo "Validating Azure credentials..."
	@if command -v az >/dev/null 2>&1; then
		@if az login --service-principal --username "$$AZURE_CLIENT_ID" --password "$$AZURE_CLIENT_SECRET" --tenant "$$AZURE_TENANT_ID" --allow-no-subscriptions >/dev/null; then
			@echo "Azure credentials validated successfully"
		else
			@echo "Error: Azure credentials validation failed"
			@exit 1
		fi
	else
		@echo "Warning: Azure CLI not available for validation. Ensure credentials are correct."
	fi

##@ Build and Create

Prepare-WebhookDeployment: ## Ensure deployment artifacts exist; regenerate if missing or older than webhook-$(LOCATION).env
	@env_file="$(project_root)/$(ROLE)-$(LOCATION).env"
	@if [[ ! -f "$$env_file" ]]; then
		echo "Missing environment file: $$env_file"
		exit 1
	fi
	regen=0
	@if [[ ! -f "$(project_file)" || "$(project_file)" -ot "$$env_file" ]]; then
		regen=1
	fi
	@certificate_request_file="$(project_root)/$(ROLE).config/$(LOCATION)/ssl-certificates/certificate-request.conf"
	@if [[ ! -f "$${certificate_request_file}" || "$${certificate_request_file}" -ot "$${env_file}" ]]; then
		regen=1
	fi
	if (( regen )); then
		$(project_root)/bin/Prepare-WebhookDeployment --env-file="$$env_file"
	fi

New-Webhook: New-WebhookImage New-WebhookContainer ## Build image and create container
	@echo -e "\n\033[1mWhat's next:\033[0m"
	@echo "    Start Webhook in $(LOCATION): make Start-Webhook [IP_ADDRESS=<IP_ADDRESS>]"

New-WebhookCertificates: $(ssl_certificates_root)/certificate-request.conf ## Generate self-signed SSL certificates for LOCATION
	@cd "$(ssl_certificates_root)"
	@openssl req -x509 -new -config certificate-request.conf -nodes -days 365 -out certificate.pem
	@openssl req -new -config certificate-request.conf -nodes -key private-key.pem -out self-signed.csr
	@chmod -R 600 $(ssl_certificates_root)

New-WebhookContainer: $(project_file) $(ssh_keys) $(ssl_certificates) $(webhook_hooks) $(webhook_command) ## Create container from existing image and prepare volumes
	$(if $(AZURE_CLIENT_ID),, $(error AZURE_CLIENT_ID not set))
	$(if $(AZURE_CLIENT_SECRET),, $(error AZURE_CLIENT_SECRET not set))
	$(if $(AZURE_TENANT_ID),, $(error AZURE_TENANT_ID not set))

	@if [[ "$(network_driver)" == "macvlan" && -z "$(IP_RANGE)" ]]; then
		echo "An IP_RANGE is required for macvlan. Take care to ensure it does not overlap with the pool of addresses managed by your DHCP Server."
		exit 1
	fi

	@if [[ -n "$(IP_ADDRESS)" ]]; then
		if ! grepcidr "$(IP_RANGE)" <(echo "$(IP_ADDRESS)") >/dev/null 2>&1; then
			echo "Failure: $(IP_ADDRESS) is NOT in $(IP_RANGE)"
			exit 1
		fi
	fi

	$(docker_compose) stop && "$(project_root)/bin/New-DockerNetwork" --device "$(network_device)" --driver "$(network_driver)" --ip-range "$(IP_RANGE)" webhook
	$(docker_compose) create --force-recreate --pull never --remove-orphans
	@sudo docker inspect "$(CONTAINER_HOSTNAME)"

	@echo -e "\n\033[1mWhat's next:\033[0m"
	@echo "    Start Webhook in $(LOCATION): make Start-Webhook [IP_ADDRESS=<IP_ADDRESS>]"

New-WebhookExecutorToken: ## Generate a JWT token with a random secret and save it to Azure Key Vault
	@./bin/New-WebhookExecutorToken --keyvault-url "$(WEBHOOK_KEYVAULT_URL)" --secret-name "$(WEBHOOK_SECRET_NAME)" --token-name "$(WEBHOOK_TOKEN_NAME)" --algorithm "$(WEBHOOK_JWT_ALGORITHM)" --payload "$(WEBHOOK_JWT_PAYLOAD)" --location "$(LOCATION)"

New-WebhookImage: ## Build the Webhook image only
	@echo "PLATFORM=$(PLATFORM)"
	@sudo docker buildx build \
		--platform $(PLATFORM) \
		--build-arg webhook_executor_exclude=$(WEBHOOK_EXECUTOR_EXCLUDE) \
		--build-arg s6_overlay_version=$(S6_OVERLAY_VERSION) \
		--build-arg webhook_version=$(WEBHOOK_VERSION) \
		--build-arg webhook_port=$(WEBHOOK_PORT) \
		$(BUILDX_LOAD) $(BUILDX_PUSH) \
		--progress=plain \
		--tag "$(IMAGE)" .
	@echo -e "\n\033[1mWhat's next:\033[0m"
	@echo "    Create Webhook container in $(LOCATION): make New-WebhookContainer [IP_ADDRESS=<IP_ADDRESS>]"

New-WebhookKeys: ## Generate SSH keys for LOCATION
	@mkdir --parent $(ssh_keys_root)
	@ssh-keygen -t rsa -b 4096 -f "$(ssh_keys_root)/id_rsa" -N "" <<< $$'y\n'
	@chmod -R 600 $(ssh_keys_root)

##@ Lifecycle
Restart-Webhook: $(container_certificates) $(container_hooks) $(container_keys) ## Restart container
	$(docker_compose) restart
	$(MAKE) Get-WebhookStatus
 
Start-Webhook: $(container_certificates) $(container_hooks) $(container_keys) ## Start container
	$(docker_compose) start
	$(MAKE) Get-WebhookStatus

Start-WebhookShell: ## Open interactive shell in the container
	@sudo docker exec --interactive --tty ${CONTAINER_HOSTNAME} /bin/bash

Stop-Webhook: ## Stop container
	$(docker_compose) stop
	$(MAKE) Get-WebhookStatus

##@ Runtime status
Get-WebhookStatus: $(project_file) ## Show compose status (JSON)
	$(docker_compose) ps --all --format json --no-trunc | jq .

##@ Runtime resource updates

Update-WebhookKeys: $(ssh_keys) ## Copy SSH keys into container volume for LOCATION
	@mkdir --parent "$(volume_root)/ssh"
	@cp --verbose $(ssh_keys) "$(volume_root)/ssh"

Update-WebhookCertificates: $(ssl_certificates) ## Copy SSL certificates into container volume for LOCATION
	@mkdir --parent "$(volume_root)/ssl-certificates" "$(volume_root)/ssh"
	@cp --verbose $(ssl_certificates) "$(volume_root)/ssl-certificates"
	@cp --verbose $(ssh_keys) "$(volume_root)/ssh"
	@echo -e "\n\033[1mWhat's next:\033[0m"
	@echo "    Ensure that Webhook in us-wa loads new certificates: make Restart-Webhook"

Update-WebhookHooks: $(webhook_hooks) $(webhook_command) ## Copy hooks.json and command script into container volume for LOCATION
	@mkdir --parent "$(volume_root)"
	@cp --verbose "$(webhook_hooks)" "$(volume_root)"

Update-WebhookHooksEnv: ## Copy hooks.env to webhook volumes
	@mkdir --parent "$(volume_root)"
	@cp --verbose "webhook.config/$(LOCATION)/hooks.env" "$(volume_root)/hooks.env"
	@echo -e "\n\033[1mWhat's next:\033[0m"
	@echo "    Restart Webhook to load new hooks.env: make Restart-Webhook"

## BUILD RULES

# ssl certificates

$(ssl_certificates_root)/certificate-request.conf: $(project_root)/webhook-$(LOCATION).env
	$(project_root)/bin/Prepare-WebhookDeployment --env-file $(project_root)/webhook-$(LOCATION).env

$(ssl_certificates): $(ssl_certificates_root)/certificate-request.conf
	@echo $(ssl_certificates)
	$(MAKE) New-WebhookCertificates

$(container_certificates): $(ssl_certificates)
	$(MAKE) Update-WebhookCertificates

### ssh keys

$(ssh_keys):
	$(MAKE) New-WebhookKeys

$(container_keys): $(ssh_keys)
	$(MAKE) Update-WebhookKeys

### webhook hooks

$(webhook_hooks):
	@echo '[]' > $(webhook_hooks)

$(container_hooks):
	$(MAKE) Update-WebhookHooks

## Location artifact rules: if missing or stale vs env/templates, (re)generate via Prepare-WebhookDeployment

override env_file := $(project_root)/webhook-$(LOCATION).env
override env_stamp := $(project_root)/.env-$(LOCATION).stamp
override compose_template := $(project_root)/services.yaml.template
override certreq_template := $(project_root)/certificate-request.conf.template

$(env_file):
	@echo "Missing environment file: $@"
	@echo "Create it or symlink it into the project root (e.g., from test/baseline)."
	@echo "Expected path: $(project_root)/$(ROLE)-$(LOCATION).env"

$(env_stamp): $(env_file)
	@touch "$@"

$(project_file): $(compose_template) $(certreq_template) $(env_stamp)
	$(MAKE) Prepare-WebhookDeployment
