# Frontend

The `front` directory contains a React, TypeScript, Vite, Tailwind, and
shadcn-style UI frontend for kube-dash.

## Local development

Run the backend on `localhost:8080`, then start the frontend:

```bash
cd front
npm install
npm run dev
```

Vite proxies `/api/*` and `/healthz` to `http://localhost:8080`.

## Build

```bash
cd front
npm run build
```

The production bundle is written to `front/dist`.

## Screens

The dashboard shows cluster summary cards, a namespace selector, and resource
tables for nodes, pods, deployments, services, and ingresses. Data is loaded
from the existing read-only `/api/v1` backend endpoints.
