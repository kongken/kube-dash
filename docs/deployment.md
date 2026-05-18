# Deployment

kube-dash is designed to run inside Kubernetes and uses in-cluster
authentication by default.

## Container Image

Images are built and published to GitHub Container Registry by the
`Docker` workflow (`.github/workflows/docker-publish.yml`) on every push to
`main`, every semver tag (`v*.*.*`), and on a daily schedule.

```
ghcr.io/kongken/kube-dash:main      # latest main
ghcr.io/kongken/kube-dash:vX.Y.Z    # tagged release
```

Pull requests trigger a build (without push) to validate the Dockerfile.
Pushed images are signed with cosign using the workflow's OIDC identity.

To build locally:

```bash
docker build -t kube-dash:dev .
```

## Runtime Auth

The server calls `rest.InClusterConfig()` unless `KUBE_DASH_KUBECONFIG` is set.
Do not set `KUBE_DASH_KUBECONFIG` in production pods.

The service account needs read access to the resources exposed by the API:

- namespaces
- nodes
- pods
- services
- deployments
- ingresses

## Minimal RBAC

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: kube-dash
  namespace: kube-dash
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: kube-dash-readonly
rules:
  - apiGroups: [""]
    resources: ["namespaces", "nodes", "pods", "services"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["apps"]
    resources: ["deployments"]
    verbs: ["get", "list", "watch"]
  - apiGroups: ["networking.k8s.io"]
    resources: ["ingresses"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: kube-dash-readonly
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: kube-dash-readonly
subjects:
  - kind: ServiceAccount
    name: kube-dash
    namespace: kube-dash
```

## Runtime Configuration

Butterfly requires a config backend. For file-backed configuration, mount a
config file and start the service with:

```bash
kube-dash --config.type=file --config.file.path=/etc/kube-dash/config.yaml
```

The current application config object is intentionally empty, so the config file
can remain minimal until new runtime settings are added.

## Health Check

Use `/healthz` for readiness or liveness probes. The endpoint reports whether
the Kubernetes client has been initialized.
