
# Dockerized Webhook (Fork)

**This repository is a fork of [adnanh/webhook](https://github.com/adnanh/webhook/) and [almir/webhook Docker image](https://hub.docker.com/r/almir/webhook/).**

## About This Fork

This fork was created to customize, extend, or maintain the original webhook Docker image for NobleFactor's needs. If you are looking for the original project, see [adnanh/webhook](https://github.com/adnanh/webhook/) and [almir/webhook](https://hub.docker.com/r/almir/webhook/).

**Changes from upstream:**
- [List any changes, fixes, or customizations here. If none, you can state: "No changes yet, this is a direct fork for future development."]

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

## License

This project inherits its license from the upstream repositories. See [LICENSE](LICENSE) for details.
