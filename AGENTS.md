# kube-dash Agent Instructions

This repository contains the kube-dash Kubernetes dashboard backend. Keep changes
small, documented, and aligned with the existing Butterfly-based Go service.

## Project Shape

- `cmd/kube-dash/main.go` wires the Butterfly app entrypoint.
- `internal/dashboard/` owns the HTTP API, Kubernetes client initialization, and
  handler tests.
- `docs/` contains operator and developer documentation. Update it when API,
  auth, deployment, or runtime behavior changes.

## Development Rules

- Do not work directly on `main`; create a task branch from the latest target
  branch.
- Prefer in-cluster Kubernetes auth. Local kubeconfig support is only for
  development through `KUBE_DASH_KUBECONFIG`.
- Keep API responses stable and document any response-shape changes in
  `docs/api.md`.
- Add or update tests for handler behavior, Kubernetes client error paths, and
  request parameter handling.
- Run `go test ./...` before committing Go changes when the local toolchain is
  available.

## Documentation Rules

- Keep README concise and link to detailed docs instead of duplicating them.
- Add operational notes to `docs/deployment.md`.
- Add local setup, config, and testing notes to `docs/development.md`.
- Add endpoint contracts and examples to `docs/api.md`.
