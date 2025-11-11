# Dockerfile for https://github.com/NobleFactor/docker-webhook

## Build

FROM golang:trixie AS build

LABEL org.opencontainers.image.vendor="Noble Factor" org.opencontainers.image.licenses="MIT" org.opencontainers.image.authors="David.Noble@noblefactor.com"
WORKDIR /go/src/github.com/noblefactor/webhook
ARG webhook_version=2.8.2

RUN <<EOF
apt update --yes
apt install --no-install-recommends build-deps curl libc-dev gcc libgcc
curl -L --silent -o webhook.tar.gz https://github.com/adnanh/webhook/archive/${webhook_version}.tar.gz && \
tar -xzf webhook.tar.gz --strip 1
go mod download
CGO_ENABLED=0 go build -ldflags="-s -w" -o /usr/local/bin/webhook
EOF

## Runtime

FROM debian:trixie

ARG s6_overlay_version=3.2.1.0
ARG webhook_port=9000

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
apt-get install --yes --no-install-recommends ca-certificates tzdata
rm -rf /var/lib/apt/lists/*
EOF

##### Install s6 overlay
ADD https://github.com/just-containers/s6-overlay/releases/download/v${s6_overlay_version}/s6-overlay-aarch64.tar.xz /tmp/

RUN <<EOF
tar xzf /tmp/s6-overlay-aarch64.tar.xz -C /
mkdir -p /etc/services.d/webhook/log
rm /tmp/s6-overlay-aarch64.tar.xz
EOF

##### Install webhook
COPY --from=build /usr/local/bin/webhook /usr/local/bin/webhook
RUN useradd --system --uid 1000 --user-group webhook

##### user
RUN echo webhook > /etc/services.d/webhook/user

##### run
RUN cat > /etc/services.d/webhook/run <<EOF
#!/usr/bin/execlineb -P
webhook -verbose -hooks=/usr/local/etc/webhook/hooks.json -hotreload -port=\${WEBHOOK_PORT} -secure -cert=/usr/local/etc/webhook/ssl-certificates/certificate.pem -key=/usr/local/etc/webhook/ssl-certificates/private-key.pem
EOF
RUN chmod +x /etc/services.d/webhook/run

##### log
RUN cat > /etc/services.d/webhook/log/run <<EOF
#!/usr/bin/execlineb -P
s6-log /var/log/webhook
EOF
RUN chmod +x /etc/services.d/webhook/log/run

### Setup container environment

ENV         WEBHOOK_PORT=${webhook_port}
VOLUME      [ "/usr/local/etc/webhook" ]
WORKDIR     /usr/local/etc/webhook

ENTRYPOINT ["/init"]
