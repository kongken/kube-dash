# Development

## Prerequisites

- Go toolchain compatible with the module in `go.mod`
- Access to a Kubernetes cluster for manual endpoint testing
- Optional local kubeconfig for development-only auth

## Run Locally

Butterfly requires a config backend. Start kube-dash with a file config:

```bash
go run ./cmd/kube-dash --config.type=file --config.file.path=config.yaml
```

For local Kubernetes API access, set:

```bash
export KUBE_DASH_KUBECONFIG="$HOME/.kube/config"
```

Unset `KUBE_DASH_KUBECONFIG` when validating in-cluster behavior.

## Test

Run the Go test suite:

```bash
go test ./...
```

Handler tests should use the Kubernetes fake client instead of requiring a live
cluster.

## API Changes

When changing handlers or response fields:

1. Update or add tests in `internal/dashboard`.
2. Update `docs/api.md`.
3. Keep README as a short entrypoint and link to detailed docs.
