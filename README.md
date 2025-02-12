# alertmanager-discord

This is a webserver that accepts webhooks from AlertManager. It will post your Prometheus alert notifications into a Discord channel as they trigger:

![](/.github/discord-screenshot.png)

## Warning

This program is not a replacement to alertmanager, it accepts webhooks from alertmanager, not Prometheus.

The standard "dataflow" should be:

```text
Prometheus -------------> alertmanager -------------------> alertmanager-discord

alerting:                 receivers:
  alertmanagers:          - name: 'discord_webhook'         environment:
  - static_configs:         webhook_configs:                   - DISCORD_WEBHOOK=https://discordapp.com/api/we...
    - targets:              - url: 'http://localhost:9094'
       - 127.0.0.1:9093
```

## Features

- REST API
- Small, standalone binary ( less than 12 Mb)
- Small Docker (OCI) Image (also less than 12 Mb) with minimal dependencies
- Helm Chart for deployment to Kubernetes.
  - includes Cilium Network Policies which can be optionally enabled.
- Liveness and Readiness probes, at `/liveness` and `/readiness`.
- Unit and Integration tests, approx 90% coverage.
- Structured Logging.
- Prometheus metrics at `/metrics`.

### Roadmap

- Template Discord messages
- REST API documented with OpenAPI (Swagger) specification.

## Example alertmanager config

```yaml
global:
  # The smarthost and SMTP sender used for mail notifications.
  smtp_smarthost: "localhost:25"
  smtp_from: "alertmanager@example.org"
  smtp_auth_username: "alertmanager"
  smtp_auth_password: "password"

# The directory from which notification templates are read.
templates:
  - "/etc/alertmanager/template/*.tmpl"

# The root route on which each incoming alert enters.
route:
  group_by: ["alertname"]
  group_wait: 20s
  group_interval: 5m
  repeat_interval: 3h
  receiver: discord_webhook

receivers:
  - name: "discord_webhook"
    webhook_configs:
      - url: "http://localhost:9094"
```

## Deployment

### Running binary

```shell
go run . --discord_webhook_url=https://discord.com/api/webhooks/123456789123456789/abc
```

You may instead provide the Discord webhook url by environment variable, `DISCORD_WEBHOOK_URL`, or via a configuration file:

```yaml
discord_webhook_url: https://discord.com/api/webhooks/123456789123456789/abc
```

```shell
go run . --configuration_file_path=/path/to/your/config.yaml
```

### Docker or OCI-compatible container runtime

If you wish to deploy this to Docker, or similar OCI-compatible container runtime, you can pull the OCI image from the [Docker Hub repository](https://hub.docker.com/r/speckle/alertmanager-discord/).

### Kubernetes Helm Chart

If you wish to deploy this to Kubernetes, this repository contains a Helm Chart.

Firstly, please deploy a Secret with your configuration information, the discord webhook url is required.:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: discord-config
data:
  "config.yaml": |
    discord_webhook_url: https://discord.com/api/webhooks/123456789123456789/abc
```

```shell
helm upgrade --install \
--create-namespace \
--namespace alertmanager-discord
alertmanager-discord \
./deploy/helm
```

You can optionally also provide a values yaml file, `--values ./your-values.yaml`, to override the default values.

## Development

To build the binary locally:

```shell
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o /tmp/alertmanager-discord ./cmd/alertforward
```

To build the Dockerfile locally:

```shell
docker build . -t speckle/alertmanager-discord:local
```

Or to build the Dockerfile on Apple Silicon (M1, M2 etc.):

```shell
docker buildx build --platform=linux/amd64 . -t speckle/alertmanager-discord:local
```

### Pre-commit

A pre-commit configuration is provided. With [pre-commit](https://pre-commit.com/) installed, run:

```shell
pre-commit install
```

This should install hooks on git, which will cause pre-commit to run every time a git commit is created.

Alternatively, to run pre-commit on the entire repository:

```shell
pre-commit run --all-files
```

### Upgrading

```shell
go get -u ./...
```

### Testing

```shell
go test ./... -v -cover -test.shuffle on
```

## Design philosophy

- small footprint
- Minimal external dependencies
- binary should be agnostic to deployment location or method.
- synchronous; the connection to the server is kept open until the connection to Discord has responded (or errored). This allows the response code or error to be returned to the request - we can have more confidence that the message was sent, and have a better ability to quickly correlate which requests caused an error.

## Acknowledgements

This repository is forked from [benjojo/alertmanager-discord](https://github.com/benjojo/alertmanager-discord) under the Apache 2.0 license
