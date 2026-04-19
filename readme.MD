
# k8s-bot

`k8s-bot` is a Go-based Telegram bot for monitoring a Kubernetes cluster and Argo CD applications.

## Features

- Kubernetes node monitoring
- Argo CD application health and sync status tracking
- Telegram notifications for startup reports and alerts
- Configurable logging level
- Optional Telegram proxy support

## Configuration

The application reads configuration from `config.toml` in the project root and from environment variables with the `K8SBOT_` prefix.

### Environment variables

The following environment variables are supported:

- `K8SBOT_CLUSTER_API_URL` — Kubernetes API server URL
- `K8SBOT_CLUSTER_TOKEN` — Kubernetes service account token
- `K8SBOT_TELEGRAM_TOKEN` — Telegram bot token
- `K8SBOT_TELEGRAM_CHAT_ID` — Telegram chat ID
- `K8SBOT_TELEGRAM_PROXY` — Optional HTTP/HTTPS proxy for Telegram requests

### `config.toml`

Example configuration:

```toml
[app]
log_level = "info"
[monitoring]
check_interval = "60s"
nodes_enabled = true
argocd_enabled = true

[namespaces]
include = ["default", "production"]

[cluster]
api_url = "https://k8s.acme.com:6443"
token = "xxx"

[telegram]
token = "xxx"
chat_id = "xxx"
proxy = ""

```

## Notes

- `check_interval`, `nodes_enabled`, `argocd_enabled`, and `namespaces.include` are present in configuration, but not all of them are currently used in the code.
- The Kubernetes client is configured with `insecure: true`, so TLS verification is disabled.
- Keep secrets out of the repository.



