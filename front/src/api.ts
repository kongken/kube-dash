export interface Summary {
  namespaces: number;
  nodes: {
    total: number;
    ready: number;
  };
  pods: {
    total: number;
    running: number;
    pending: number;
    succeeded: number;
    failed: number;
    unknown: number;
  };
  deployments: number;
  services: number;
  ingresses: number;
}

export interface Namespace {
  name: string;
  status: string;
  ageSeconds: number;
}

export interface Node {
  name: string;
  ready: boolean;
  roles: string[];
  version: string;
  ageSeconds: number;
}

export interface Pod {
  name: string;
  namespace: string;
  node: string;
  phase: string;
  ready: {
    ready: number;
    total: number;
  };
  restarts: number;
  ageSeconds: number;
}

export interface Deployment {
  name: string;
  namespace: string;
  replicas: number;
  ready: number;
  available: number;
  updated: number;
  ageSeconds: number;
}

export interface ServicePort {
  name: string;
  port: number;
  targetPort: string;
  protocol: string;
}

export interface Service {
  name: string;
  namespace: string;
  type: string;
  clusterIP: string;
  ports: ServicePort[];
  ageSeconds: number;
}

export interface IngressPath {
  path: string;
  pathType: string;
  serviceName: string;
  servicePort: string;
}

export interface IngressRule {
  host: string;
  paths: IngressPath[];
}

export interface Ingress {
  name: string;
  namespace: string;
  className: string;
  rules: IngressRule[];
  tls: Array<{ hosts: string[]; secretName: string }>;
  loadBalancers: Array<{ hostname: string; ip: string }>;
  ageSeconds: number;
}

export interface ListResponse<T> {
  items: T[];
}

export interface DashboardData {
  summary: Summary;
  namespaces: Namespace[];
  nodes: Node[];
  pods: Pod[];
  deployments: Deployment[];
  services: Service[];
  ingresses: Ingress[];
}

async function request<T>(path: string): Promise<T> {
  const response = await fetch(path);
  if (!response.ok) {
    let message = `${response.status} ${response.statusText}`;
    try {
      const body = (await response.json()) as { error?: string };
      if (body.error) {
        message = body.error;
      }
    } catch {
      // Keep the HTTP status message when the response is not JSON.
    }
    throw new Error(message);
  }
  return response.json() as Promise<T>;
}

function namespaceQuery(namespace: string) {
  return namespace === "_all" ? "" : `?namespace=${encodeURIComponent(namespace)}`;
}

export async function loadDashboardData(namespace: string): Promise<DashboardData> {
  const query = namespaceQuery(namespace);
  const [summary, namespaces, nodes, pods, deployments, services, ingresses] =
    await Promise.all([
      request<Summary>("/api/v1/summary"),
      request<ListResponse<Namespace>>("/api/v1/namespaces"),
      request<ListResponse<Node>>("/api/v1/nodes"),
      request<ListResponse<Pod>>(`/api/v1/pods${query}`),
      request<ListResponse<Deployment>>(`/api/v1/deployments${query}`),
      request<ListResponse<Service>>(`/api/v1/services${query}`),
      request<ListResponse<Ingress>>(`/api/v1/ingresses${query}`)
    ]);

  return {
    summary,
    namespaces: namespaces.items,
    nodes: nodes.items,
    pods: pods.items,
    deployments: deployments.items,
    services: services.items,
    ingresses: ingresses.items
  };
}
