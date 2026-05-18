# kube-dash API

kube-dash exposes read-only Kubernetes dashboard endpoints under `/api/v1`.
All Kubernetes list calls use a 10 second request timeout.

It also exposes a Model Context Protocol endpoint at `/mcp` using the official
Go SDK streamable HTTP transport.

## Health

### `GET /healthz`

Returns process health and whether the Kubernetes client has been initialized.

```json
{
  "status": "ok",
  "kubernetes": "initialized"
}
```

## Cluster Summary

### `GET /api/v1/summary`

Returns aggregate counts for namespaces, nodes, pods, deployments, services, and
ingresses.

```json
{
  "namespaces": 4,
  "nodes": {
    "total": 3,
    "ready": 3
  },
  "pods": {
    "total": 12,
    "running": 11,
    "pending": 1,
    "succeeded": 0,
    "failed": 0,
    "unknown": 0
  },
  "deployments": 5,
  "services": 7,
  "ingresses": 2
}
```

## Resources

### `GET /api/v1/namespaces`

Returns namespaces sorted by name.

Item fields:

- `name`
- `status`
- `ageSeconds`

### `GET /api/v1/nodes`

Returns nodes sorted by name.

Item fields:

- `name`
- `ready`
- `roles`
- `version`
- `ageSeconds`

### `GET /api/v1/pods?namespace=<name>`

Returns pods sorted by namespace and name. If `namespace` is omitted, pods from
all namespaces are returned.

Item fields:

- `name`
- `namespace`
- `node`
- `phase`
- `ready`
- `restarts`
- `ageSeconds`

### `GET /api/v1/deployments?namespace=<name>`

Returns deployments sorted by namespace and name. If `namespace` is omitted,
deployments from all namespaces are returned.

Item fields:

- `name`
- `namespace`
- `replicas`
- `ready`
- `available`
- `updated`
- `ageSeconds`

### `GET /api/v1/services?namespace=<name>`

Returns services sorted by namespace and name. If `namespace` is omitted,
services from all namespaces are returned.

Item fields:

- `name`
- `namespace`
- `type`
- `clusterIP`
- `ports`
- `ageSeconds`

Each port contains:

- `name`
- `port`
- `targetPort`
- `protocol`

### `GET /api/v1/ingresses?namespace=<name>`

Returns ingresses sorted by namespace and name. If `namespace` is omitted,
ingresses from all namespaces are returned.

Item fields:

- `name`
- `namespace`
- `className`
- `rules`
- `tls`
- `loadBalancers`
- `ageSeconds`

Each rule contains:

- `host`
- `paths`

Each path contains:

- `path`
- `pathType`
- `serviceName`
- `servicePort`

Each TLS entry contains:

- `hosts`
- `secretName`

Each load balancer entry contains:

- `hostname`
- `ip`

## MCP

### `POST /mcp`

Serves MCP over streamable HTTP. The endpoint is stateless and returns JSON
responses for tool calls.

Available tools:

- `cluster_summary`
- `list_namespaces`
- `list_nodes`
- `list_pods`
- `list_deployments`
- `list_services`
- `list_ingresses`

The list tools for pods, deployments, services, and ingresses accept an optional
`namespace` argument. When `namespace` is omitted, they return resources from
all namespaces.

## Errors

If the Kubernetes client is unavailable, endpoints return `503`.

```json
{
  "error": "kubernetes client is not initialized"
}
```

If a Kubernetes request times out, endpoints return `504`. Other Kubernetes API
errors return `500` with the error message.
