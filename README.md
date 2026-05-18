# kube-dash

A small Kubernetes dashboard backend built with the Butterfly Go framework.

## API

- `GET /healthz`
- `GET /api/v1/summary`
- `GET /api/v1/namespaces`
- `GET /api/v1/nodes`
- `GET /api/v1/pods?namespace=<name>`
- `GET /api/v1/deployments?namespace=<name>`
- `GET /api/v1/services?namespace=<name>`
- `GET /api/v1/ingresses?namespace=<name>`

The service uses Kubernetes in-cluster authentication by default. For local
development only, set `KUBE_DASH_KUBECONFIG=/path/to/kubeconfig`.

Butterfly requires a config backend. Run locally with a file config:

```bash
go run ./cmd/kube-dash --config.type=file --config.file.path=config.yaml
```

## Documentation

- [API reference](docs/api.md)
- [Development guide](docs/development.md)
- [Deployment notes](docs/deployment.md)
