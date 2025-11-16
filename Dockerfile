########################################################################################################################
# SPDX-FileCopyrightText: 2016-2025 Noble Factor
# SPDX-License-Identifier: MIT
########################################################################################################################

# Dockerfile for https://github.com/NobleFactor/docker-webhook

ARG BUILDPLATFORM
FROM --platform=$BUILDPLATFORM golang:trixie AS webhook_build

LABEL org.opencontainers.image.vendor="Noble Factor" org.opencontainers.image.licenses="MIT" org.opencontainers.image.authors="David.Noble@noblefactor.com"

# Build webhook from https://github.com/adnanh/webhook by Adnan Hajdarević
WORKDIR /go/src/github.com/adnanh/webhook
ARG webhook_version=2.8.2

# Attribution: This code is from https://github.com/adnanh/webhook by Adnan Hajdarević
# Directory structure after RUN:
# /go/src/github.com/adnanh/webhook/
# ├── go.mod                    # Module definition from adnanh/webhook repo
# ├── go.sum                    # Dependencies from adnanh/webhook repo
# ├── main.go                   # Main webhook source file
# ├── ... (other webhook source files: hooks.go, rules.go, etc.)
# ├── webhook.tar.gz            # Downloaded tarball (not removed)
# └── /usr/local/bin/
#     └── webhook               # Built webhook binary

RUN <<EOF
apt-get update --yes
apt-get install --no-install-recommends --yes curl
curl --fail --show-error --silent --retry 3 --retry-delay 2 --location --output webhook.tar.gz "https://github.com/adnanh/webhook/archive/${webhook_version}.tar.gz"
# Verify checksum (update this value for new versions)
echo "84f2d581d549236512d3c214e7d97bf7 webhook.tar.gz" | md5sum -c -
tar -xzf webhook.tar.gz --strip 1
go mod download
CGO_ENABLED=0 go build -ldflags="-s -w" -o /usr/local/bin/webhook
apt remove --yes curl
rm -rf /var/lib/apt/lists/*
EOF

ARG BUILDPLATFORM
FROM --platform=$BUILDPLATFORM golang:trixie AS webhook_executor_build

LABEL org.opencontainers.image.vendor="Noble Factor" org.opencontainers.image.licenses="MIT" org.opencontainers.image.authors="David.Noble@noblefactor.com"

WORKDIR /go/src/github.com/noblefactor/docker-webhook
COPY src/ ./
ARG webhook_executor_exclude=false

# Attribution: This code is from https://github.com/NobleFactor/docker-webhook by David Noble
# Directory structure after COPY:
# /go/src/github.com/noblefactor/docker-webhook/
# ├── go.mod                    # Our module definition
# ├── go.sum                    # Our module dependencies (generated later)
# ├── internal/                 # Our internal packages (from src/internal/)
# │   ├── azure/
# │   │   └── keyvault.go       # Azure Key Vault functions
# │   └── jwt/
# │       └── jwt.go            # JWT validation functions
# └── cmd/                      # Our source directory (from src/cmd/)
#     └── webhook-executor/
#         └── main.go           # Main Go file for webhook-executor

SHELL [ "/usr/bin/env", "bash", "-o", "errexit", "-o", "nounset", "-o", "pipefail", "-c" ]

# Initialize module and download dependencies (safer: require committed go.mod/go.sum)
RUN <<'EOF'
if [[ ${webhook_executor_exclude:-false} != true ]]; then
    echo "Building webhook-executor..."

    # Build the executor from the module root.
    cd cmd/webhook-executor
    go mod tidy
    CGO_ENABLED=0 go build -ldflags="-s -w" -o /usr/local/bin/webhook-executor .
else
    echo "Skipping webhook-executor build"
fi
EOF

FROM debian:trixie-slim AS runtime

ARG TARGETARCH
ARG webhook_port=9000
ARG s6_overlay_version=3.2.1.0

SHELL [ "/usr/bin/env", "bash", "-o", "errexit", "-o", "nounset", "-o", "pipefail", "-c" ]

### Setup S6 Overlay and webook

# S6 service tree structure:
#
# /etc/services.d/
# └── webhook/
#     ├── run         # Starts the webhook process
#     ├── user        # Specifies the 'webhook' user for the service
#     └── log/
#         └── run     # Handles logging for the webhook service

##### Install prerequisites
RUN <<EOF
apt-get update --yes
apt-get install --yes --no-install-recommends ca-certificates curl jq openssh-client tzdata xz-utils
rm -rf /var/lib/apt/lists/*
EOF

##### Install s6 overlay
RUN <<EOF
case $TARGETARCH in
  amd64)
    arch=x86_64
    ;;
  arm64)
    arch=aarch64
    ;;
  *)
    echo "Unsupported architecture: $TARGETARCH" >&2
    exit 1
    ;;
esac
declare -r arch_archive="s6-overlay-${arch}.tar.xz"
declare -r noarch_archive="s6-overlay-noarch.tar.xz"

curl --fail --show-error --location --retry 3 --retry-delay 2 --output "/tmp/${arch_archive}" "https://github.com/just-containers/s6-overlay/releases/download/v${s6_overlay_version}/${arch_archive}"
tar --directory=/ --xz --extract --preserve-permissions --file="/tmp/${arch_archive}"

curl --fail --show-error --location --retry 3 --retry-delay 2 --output "/tmp/${noarch_archive}" "https://github.com/just-containers/s6-overlay/releases/download/v${s6_overlay_version}/${noarch_archive}"
tar --directory=/ --xz --extract --preserve-permissions --file="/tmp/${noarch_archive}"

rm -f "/tmp/${arch_archive}" "/tmp/${noarch_archive}"
apt-get purge --auto-remove --yes xz-utils curl
EOF

##### Install webhook and (maybe) webhook-executor

COPY --from=webhook_build /usr/local/bin/webhook /usr/local/bin/webhook
COPY --from=webhook_executor_build /usr/local/bin/webhook-executor /usr/local/bin/webhook-executor

RUN <<EOF
useradd --system --uid 999 --user-group webhook
mkdir -p /etc/services.d/webhook/log
EOF

##### user
RUN echo webhook > /etc/services.d/webhook/user

##### run
RUN install --owner=root --group=root --mode=0755 /dev/stdin /etc/services.d/webhook/run <<'EOF'
#!/bin/sh -e
# shellcheck shell=sh
if [ -f /usr/local/etc/webhook/hooks.env ]; then
    . /usr/local/etc/webhook/hooks.env
fi
exec webhook -verbose \
	-hooks=/usr/local/etc/webhook/hooks.json \
	-port="${WEBHOOK_PORT:-9000}" \
	-hotreload \
	-secure \
	-cert=/usr/local/etc/webhook/ssl-certificates/certificate.pem \
	-key=/usr/local/etc/webhook/ssl-certificates/private-key.pem
EOF

##### log
RUN install --owner=root --group=root --mode=0755 /dev/stdin /etc/services.d/webhook/log/run <<'EOF'
#!/command/execlineb -P
s6-log /var/log/webhook
EOF

### Setup container environment

ENV         WEBHOOK_PORT=${webhook_port}
VOLUME      [ "/usr/local/etc/webhook" ]
WORKDIR     /usr/local/etc/webhook

COPY data/noblefactor.init /noblefactor.init
RUN chmod +x /noblefactor.init

ENTRYPOINT  ["/noblefactor.init", "webhook", "/usr/local/etc/webhook"]
