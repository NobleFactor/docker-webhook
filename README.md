
# Dockerized Webhook (Fork)

**This repository is a fork of [adnanh/webhook](https://github.com/adnanh/webhook/) and [almir/webhook Docker image](https://hub.docker.com/r/almir/webhook/).**

## About This Fork

This fork was created to customize, extend, or maintain the original webhook Docker image for NobleFactor's needs. If you are looking for the original project, see [adnanh/webhook](https://github.com/adnanh/webhook/) and [almir/webhook](https://hub.docker.com/r/almir/webhook/).

**Changes from upstream:**
This fork contains a number of Noble Factor-specific improvements, tooling, and developer ergonomics on top of the upstream project:

- Tracked runtime PID-1 init wrapper: `scripts/noblefactor.init` — this is copied into the image at build time and performs UID/GID adjustment, chown of mounted volumes, and then execs the s6 overlay `/init`.
- Runtime service management via s6-overlay and execlineb — the `Dockerfile` extracts and installs s6/execline packages into `/package/admin` and exposes helper symlinks in `/command`.
- A repeatable runtime sanity test: `test/Test-DockerWebhookSanity` — a small script to start the image, wait for initialization, capture logs, inspect `/init` and s6 layout, and optionally keep the container for debugging.
- Man pages and shell completions under `share/man` and `share/completions` for developer tooling (e.g., `Test-DockerWebhookSanity`, `Declare-BashScript`, and other helper scripts).
- CI Quality Gate workflow: `.github/workflows/quality-gate.yml` runs Dockerfile and manpage/spell checks on pushes/PRs.
- Dockerfile and build improvements to avoid linter parser issues (heredoc -> printf-based RUN blocks) and to make the init wrapper tracked and testable.

If you'd like a short summary of what exactly changed in code (files/commits), run `git log --name-only feature/debian-s6-overlay-param-port..HEAD` or open the branch's PR.

---

## Getting Started

### Running webhook in Docker

You can run the webhook server using Docker. The simplest usage is to host your hooks JSON file on your machine and mount the directory as a volume:

```shell
docker run -d -p 9000:9000 -v /path/to/hooks:/etc/webhook --name=webhook \
  noblefactor/docker-webhook -verbose -hooks=/etc/webhook/hooks.json -hotreload
```

Replace `/path/to/hooks` with the path to your hooks file.

### Building Your Own Image

You can build your own Docker image using this fork:

```docker
FROM noblefactor/docker-webhook
COPY hooks.json.example /etc/webhook/hooks.json
```

Place your `Dockerfile` and `hooks.json.example` in the same directory, then build and run:

```shell
docker build -t my-webhook-image .
docker run -d -p 9000:9000 --name=webhook my-webhook-image -verbose -hooks=/etc/webhook/hooks.json -hotreload
```

Developer note: the repository build now copies `scripts/noblefactor.init` into the image and installs s6-overlay files. If you are building locally and want to run the runtime sanity test, see "Runtime sanity test" below.

### Customizing Entrypoint

You can specify parameters in your `Dockerfile` using the `CMD` instruction:

```docker
FROM noblefactor/docker-webhook
COPY hooks.json.example /etc/webhook/hooks.json
CMD ["-verbose", "-hooks=/etc/webhook/hooks.json", "-hotreload"]
```

After building, you can run the container without extra arguments:

```shell
docker run -d -p 9000:9000 --name=webhook my-webhook-image
```

---

## Contributing

Feel free to open issues or pull requests for improvements or fixes. For major changes, please discuss them first.

## Runtime sanity test

We include a small smoke test script to help validate that a built image starts correctly and that s6 init is present.

Usage (from the repository root):

```shell
# make the test executable (if needed)
chmod +x test/Test-DockerWebhookSanity

# run the quick sanity check (waits 2s by default)
./test/Test-DockerWebhookSanity --wait 2

# keep the container running for interactive inspection
./test/Test-DockerWebhookSanity --keep --wait 5
```

Notes:

- The test by default uses the image `noblefactor/webhook:1.0.0-preview.1`. Pass an `image:tag` positional argument to test a different image.

- The test will report success if it sees startup markers in the container logs (for example a version line and the "serving hooks" message). Missing hooks.json or TLS certs is expected for a plain smoke test and will not cause a failure if the server starts.

If you'd like, I can add a `make test-sanity` target or a GitHub Actions job that exercises the built image in CI.

## License

This project inherits its license from the upstream repositories. See [LICENSE](LICENSE) for details.
