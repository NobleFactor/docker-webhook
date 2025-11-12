########################################################################################################################
# SPDX-FileCopyrightText: 2016-2025 Noble Factor
# SPDX-License-Identifier: MIT
########################################################################################################################

# Dockerfile for https://github.com/NobleFactor/docker-webhook

## Build

ARG BUILDPLATFORM
FROM --platform=$BUILDPLATFORM golang:trixie AS build

LABEL org.opencontainers.image.vendor="Noble Factor" org.opencontainers.image.licenses="MIT" org.opencontainers.image.authors="David.Noble@noblefactor.com"
WORKDIR /go/src/github.com/noblefactor/webhook
ARG webhook_version=2.8.2

RUN <<EOF
apt-get update --yes
apt-get install --no-install-recommends --yes curl
# use strict curl flags so failures are visible and fail the build early
curl --fail --show-error --silent --retry 3 --retry-delay 2 --location --output webhook.tar.gz "https://github.com/adnanh/webhook/archive/${webhook_version}.tar.gz"
tar -xzf webhook.tar.gz --strip 1
go mod download
CGO_ENABLED=0 go build -ldflags="-s -w" -o /usr/local/bin/webhook
rm -rf /var/lib/apt/lists/*
EOF

## Runtime

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
apt-get install --yes --no-install-recommends ca-certificates curl openssh-client tzdata xz-utils
rm -rf /var/lib/apt/lists/*
EOF

##### Install s6 overlay

RUN <<EOF
declare -r arch_archive="s6-overlay-$(echo "$TARGETARCH" | sed 's/arm64/aarch64/').tar.xz"
declare -r noarch_archive="s6-overlay-noarch.tar.xz"

curl --fail --show-error --location --retry 3 --retry-delay 2 --output "/tmp/${arch_archive}" "https://github.com/just-containers/s6-overlay/releases/download/v${s6_overlay_version}/${arch_archive}"
tar --directory=/ --xz --extract --preserve-permissions --file="/tmp/${arch_archive}"

curl --fail --show-error --location --retry 3 --retry-delay 2 --output "/tmp/${noarch_archive}" "https://github.com/just-containers/s6-overlay/releases/download/v${s6_overlay_version}/${noarch_archive}"
tar --directory=/ --xz --extract --preserve-permissions --file="/tmp/${noarch_archive}"

rm -f "/tmp/${arch_archive}" "/tmp/${noarch_archive}"
apt-get purge --auto-remove --yes xz-utils curl
EOF

##### Install webhook

COPY --from=build /usr/local/bin/webhook /usr/local/bin/webhook
RUN <<EOF
useradd --system --uid 999 --user-group webhook
mkdir -p /etc/services.d/webhook/log
EOF

##### user
RUN echo webhook > /etc/services.d/webhook/user

##### run
RUN cat > /etc/services.d/webhook/run <<'EOF'
#!/bin/sh -e
exec webhook -verbose \
	-hooks=/usr/local/etc/webhook/hooks.json \
	-port="${WEBHOOK_PORT:-9000}" \
	-hotreload \
	-secure \
	-cert=/usr/local/etc/webhook/ssl-certificates/certificate.pem \
	-key=/usr/local/etc/webhook/ssl-certificates/private-key.pem
EOF
RUN chmod +x /etc/services.d/webhook/run

##### log
RUN cat > /etc/services.d/webhook/log/run <<EOF
#!/command/execlineb -P
s6-log /var/log/webhook
EOF
RUN chmod +x /etc/services.d/webhook/log/run

### Setup container environment

ENV         WEBHOOK_PORT=${webhook_port}
VOLUME      [ "/usr/local/etc/webhook" ]
WORKDIR     /usr/local/etc/webhook

RUN touch /noblefactor.init && chmod +x /noblefactor.init && cat > /noblefactor.init <<'EOF'
#!/usr/bin/env bash
set -o errexit -o nounset -o pipefail
declare -r user="${1}"

# Modify the give user's primary user/group ID's based on PUID/PGID environment variables at runtime.
# This allows the container to match host user/group IDs for volume permissions.

if [[ -n "${PUID:-}" ]] || [[ -n "${PGID:-}" ]]; then
    echo "Runtime user/group modification requested"

    current_uid=$(id --user "${user}")
    current_gid=$(id --group "${user}")

    if [[ -n "${PGID:-}" ]] && [[ "${PGID}" != "${current_gid}" ]]; then
        echo "Modifying webhook group: ${current_gid} -> ${PGID}"
        groupmod --gid "${PGID}" "${user}"
    fi

    if [[ -n "${PUID:-}" ]] && [[ "${PUID}" != "${current_uid}" ]]; then
        echo "Modifying webhook user: ${current_uid} -> ${PUID}"
        usermod --uid "${PUID}" "${user}"
    fi

    echo "Updating /usr/local/etc/webhook ownership..."
    chown --recursive "${user}":"$(id -gn "${user}")" /usr/local/etc/webhook 
fi

echo "Webhook user: $(id "${user}")"

# Chain to S6 overlay init
exec /init "$@"
EOF

ENTRYPOINT  ["/noblefactor.init", "webhook"]
