
# k8s-bot

`k8s-bot` is a Go Telegram bot for monitoring a Kubernetes cluster and Argo CD.

## What it does

- monitors Kubernetes nodes
- monitors Argo CD applications
- sends a startup report to Telegram
- periodically checks Kubernetes API availability
- supports repeated alerts with a configurable interval
- supports a Telegram proxy
- uses configurable log levels

## How it works

On startup, the application:

1. reads configuration from `config.toml` in the project root
2. overrides values from environment variables with the `K8SBOT_` prefix
3. creates Kubernetes client, dynamic client, and Telegram bot instances
4. sends a startup report
5. starts node and Argo CD watchers, if enabled
6. starts periodic Kubernetes API probing

## Configuration

### Configuration sources

The application looks for:

- `./config.toml`
- `./internal/app/config.toml`

It also supports environment variables with the `K8SBOT_` prefix.

### Environment variables

Supported fields:

- `K8SBOT_APP_LOG_LEVEL` — logging level (`debug`, `info`, `warn`, `error`, ...)
- `K8SBOT_MONITORING_API_PROBE_INTERVAL` — Kubernetes API probe interval, for example `60s`
- `K8SBOT_MONITORING_ALERT_REPEAT_INTERVAL` — repeated alert interval, for example `5m`
- `K8SBOT_MONITORING_NODES_ENABLED` — enable node monitoring (`true` / `false`)
- `K8SBOT_MONITORING_ARGOCD_ENABLED` — enable Argo CD monitoring (`true` / `false`)
- `K8SBOT_ARGOCD_NAMESPACE` — namespace where Argo CD applications are listed
- `K8SBOT_CLUSTER_API_URL` — Kubernetes API server URL
- `K8SBOT_CLUSTER_TOKEN` — Kubernetes API access token
- `K8SBOT_TELEGRAM_TOKEN` — Telegram bot token
- `K8SBOT_TELEGRAM_CHAT_ID` — Telegram chat/channel ID
- `K8SBOT_TELEGRAM_PROXY` — optional HTTP/HTTPS proxy for Telegram

> Note: values are loaded through `viper`, so environment variable names map to nested config fields.

### Example `config.toml`

Example configuration:

```toml
[app]
log_level = "info"

[argocd]
namespace = "argocd"

[monitoring]
api_probe_interval = "60s"
alert_repeat_interval = "5m"
nodes_enabled = true
argocd_enabled = true

[cluster]
api_url = "https://k8s.acme.com:6443"
token = "xxx"

[telegram]
token = "xxx"
chat_id = -11111
proxy = ""

```

## Monitoring behavior

### Nodes

If `monitoring.nodes_enabled = true`, the bot:

- watches `core/v1 nodes`
- tracks `Ready` status changes
- sends an alert when a node becomes unhealthy
- sends a recovery notification when a node becomes healthy again

### Argo CD

If `monitoring.argocd_enabled = true`, the bot:

- watches `applications.argoproj.io`
- uses the namespace from `argocd.namespace`
- sends alerts when application health becomes unhealthy
- logs discovered applications and their sync status on startup

### Cluster availability check

Regardless of the watchers, the bot periodically calls `ServerVersion()` on the Kubernetes API.  
If the cluster becomes unavailable, it logs a warning.

## RBAC

The repository includes `k8s-bot-rbac.yaml`, which defines:

- `ServiceAccount`
- `ClusterRole`
- `ClusterRoleBinding`

It grants access to:

- `nodes` — `get`, `list`, `watch`
- `applications.argoproj.io` — `get`, `list`, `watch`

## Build and run

The project uses:

- Go `1.26.2`
- `Wire` for dependency injection
- `zerolog` for logging
- `viper` for configuration

```bash
CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o k8s-bot .
GOMAXPROCS=1 GOMEMLIMIT=128MiB ./k8s-bot
```

## Important notes

- The Kubernetes client is created with `TLSClientConfig.Insecure = true`, so TLS certificate verification is disabled.
- Telegram notifications are sent in Markdown mode.
- If `api_probe_interval` is not set, the default value is `1m`.
- If `alert_repeat_interval` is not set, the default value is `5m`.
- Secrets should not be stored in the repository.

