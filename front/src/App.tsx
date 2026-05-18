import { useEffect, useMemo, useState } from "react";
import type { ReactNode } from "react";
import {
  Activity,
  Box,
  Boxes,
  Cloud,
  GitBranch,
  Layers,
  RefreshCcw,
  Server,
  ShieldCheck
} from "lucide-react";
import {
  DashboardData,
  Deployment,
  Ingress,
  Node,
  Pod,
  Service,
  loadDashboardData
} from "@/api";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Select } from "@/components/ui/select";
import { formatAge } from "@/lib/utils";

const ALL_NAMESPACES = "_all";

function App() {
  const [namespace, setNamespace] = useState(ALL_NAMESPACES);
  const [data, setData] = useState<DashboardData | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  const [refreshKey, setRefreshKey] = useState(0);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    loadDashboardData(namespace)
      .then((result) => {
        if (!cancelled) {
          setData(result);
        }
      })
      .catch((err: unknown) => {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : "Failed to load dashboard");
        }
      })
      .finally(() => {
        if (!cancelled) {
          setLoading(false);
        }
      });

    return () => {
      cancelled = true;
    };
  }, [namespace, refreshKey]);

  const selectedNamespace = namespace === ALL_NAMESPACES ? "All namespaces" : namespace;

  return (
    <main className="min-h-screen">
      <header className="border-b bg-card">
        <div className="mx-auto flex max-w-7xl flex-col gap-4 px-4 py-4 sm:flex-row sm:items-center sm:justify-between lg:px-6">
          <div>
            <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
              <ShieldCheck className="size-4 text-emerald-700" />
              In-cluster Kubernetes dashboard
            </div>
            <h1 className="mt-1 text-2xl font-semibold tracking-normal">kube-dash</h1>
          </div>
          <div className="flex flex-wrap items-center gap-2">
            <Select
              aria-label="Namespace"
              value={namespace}
              onChange={(event) => setNamespace(event.target.value)}
              disabled={!data}
            >
              <option value={ALL_NAMESPACES}>All namespaces</option>
              {data?.namespaces.map((item) => (
                <option key={item.name} value={item.name}>
                  {item.name}
                </option>
              ))}
            </Select>
            <Button
              variant="outline"
              size="icon"
              aria-label="Refresh"
              title="Refresh"
              onClick={() => setRefreshKey((current) => current + 1)}
            >
              <RefreshCcw className={loading ? "animate-spin" : ""} />
            </Button>
          </div>
        </div>
      </header>

      <div className="mx-auto max-w-7xl space-y-6 px-4 py-6 lg:px-6">
        {error ? <ErrorBanner message={error} /> : null}
        {data ? (
          <>
            <Overview data={data} />
            <section className="grid gap-6 xl:grid-cols-[1fr_1fr]">
              <NodesTable nodes={data.nodes} />
              <PodsTable pods={data.pods} namespace={selectedNamespace} />
              <DeploymentsTable deployments={data.deployments} />
              <ServicesTable services={data.services} />
              <IngressesTable ingresses={data.ingresses} />
            </section>
          </>
        ) : (
          <LoadingState />
        )}
      </div>
    </main>
  );
}

function Overview({ data }: { data: DashboardData }) {
  const cards = [
    {
      label: "Namespaces",
      value: data.summary.namespaces,
      icon: Layers,
      detail: `${data.namespaces.length} visible`
    },
    {
      label: "Nodes",
      value: data.summary.nodes.total,
      icon: Server,
      detail: `${data.summary.nodes.ready} ready`
    },
    {
      label: "Pods",
      value: data.summary.pods.total,
      icon: Boxes,
      detail: `${data.summary.pods.running} running`
    },
    {
      label: "Deployments",
      value: data.summary.deployments,
      icon: GitBranch,
      detail: `${data.deployments.length} in view`
    },
    {
      label: "Services",
      value: data.summary.services,
      icon: Cloud,
      detail: `${data.services.length} in view`
    },
    {
      label: "Ingresses",
      value: data.summary.ingresses,
      icon: Activity,
      detail: `${data.ingresses.length} in view`
    }
  ];

  return (
    <section className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-6">
      {cards.map((item) => {
        const Icon = item.icon;
        return (
          <Card key={item.label}>
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
              <CardTitle className="text-muted-foreground">{item.label}</CardTitle>
              <Icon className="size-4 text-primary" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-semibold">{item.value}</div>
              <p className="mt-1 text-xs text-muted-foreground">{item.detail}</p>
            </CardContent>
          </Card>
        );
      })}
    </section>
  );
}

function NodesTable({ nodes }: { nodes: Node[] }) {
  return (
    <ResourceCard title="Nodes" count={nodes.length}>
      <table className="w-full text-left text-sm">
        <thead>
          <tr className="border-b text-xs uppercase text-muted-foreground">
            <th className="py-2 font-medium">Name</th>
            <th className="py-2 font-medium">Roles</th>
            <th className="py-2 font-medium">Version</th>
            <th className="py-2 text-right font-medium">Age</th>
          </tr>
        </thead>
        <tbody>
          {nodes.map((node) => (
            <tr key={node.name} className="border-b last:border-0">
              <td className="py-3">
                <div className="font-medium">{node.name}</div>
                <Badge tone={node.ready ? "success" : "destructive"}>
                  {node.ready ? "Ready" : "NotReady"}
                </Badge>
              </td>
              <td className="py-3 text-muted-foreground">{node.roles.join(", ") || "-"}</td>
              <td className="py-3 text-muted-foreground">{node.version || "-"}</td>
              <td className="py-3 text-right text-muted-foreground">
                {formatAge(node.ageSeconds)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </ResourceCard>
  );
}

function PodsTable({ pods, namespace }: { pods: Pod[]; namespace: string }) {
  const unhealthy = useMemo(
    () => pods.filter((pod) => !["Running", "Succeeded"].includes(pod.phase)).length,
    [pods]
  );

  return (
    <ResourceCard title="Pods" count={pods.length} detail={`${namespace} · ${unhealthy} attention`}>
      <table className="w-full text-left text-sm">
        <thead>
          <tr className="border-b text-xs uppercase text-muted-foreground">
            <th className="py-2 font-medium">Pod</th>
            <th className="py-2 font-medium">Status</th>
            <th className="py-2 font-medium">Ready</th>
            <th className="py-2 text-right font-medium">Restarts</th>
          </tr>
        </thead>
        <tbody>
          {pods.slice(0, 12).map((pod) => (
            <tr key={`${pod.namespace}/${pod.name}`} className="border-b last:border-0">
              <td className="max-w-[14rem] py-3">
                <div className="truncate font-medium">{pod.name}</div>
                <div className="truncate text-xs text-muted-foreground">{pod.namespace}</div>
              </td>
              <td className="py-3">
                <Badge tone={pod.phase === "Running" ? "success" : "warning"}>{pod.phase}</Badge>
              </td>
              <td className="py-3 text-muted-foreground">
                {pod.ready.ready}/{pod.ready.total}
              </td>
              <td className="py-3 text-right text-muted-foreground">{pod.restarts}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </ResourceCard>
  );
}

function DeploymentsTable({ deployments }: { deployments: Deployment[] }) {
  return (
    <ResourceCard title="Deployments" count={deployments.length}>
      <table className="w-full text-left text-sm">
        <thead>
          <tr className="border-b text-xs uppercase text-muted-foreground">
            <th className="py-2 font-medium">Name</th>
            <th className="py-2 font-medium">Ready</th>
            <th className="py-2 font-medium">Available</th>
            <th className="py-2 text-right font-medium">Age</th>
          </tr>
        </thead>
        <tbody>
          {deployments.map((deployment) => (
            <tr key={`${deployment.namespace}/${deployment.name}`} className="border-b last:border-0">
              <td className="max-w-[14rem] py-3">
                <div className="truncate font-medium">{deployment.name}</div>
                <div className="truncate text-xs text-muted-foreground">{deployment.namespace}</div>
              </td>
              <td className="py-3 text-muted-foreground">
                {deployment.ready}/{deployment.replicas}
              </td>
              <td className="py-3 text-muted-foreground">{deployment.available}</td>
              <td className="py-3 text-right text-muted-foreground">
                {formatAge(deployment.ageSeconds)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </ResourceCard>
  );
}

function ServicesTable({ services }: { services: Service[] }) {
  return (
    <ResourceCard title="Services" count={services.length}>
      <table className="w-full text-left text-sm">
        <thead>
          <tr className="border-b text-xs uppercase text-muted-foreground">
            <th className="py-2 font-medium">Name</th>
            <th className="py-2 font-medium">Type</th>
            <th className="py-2 font-medium">Cluster IP</th>
            <th className="py-2 text-right font-medium">Ports</th>
          </tr>
        </thead>
        <tbody>
          {services.map((service) => (
            <tr key={`${service.namespace}/${service.name}`} className="border-b last:border-0">
              <td className="max-w-[14rem] py-3">
                <div className="truncate font-medium">{service.name}</div>
                <div className="truncate text-xs text-muted-foreground">{service.namespace}</div>
              </td>
              <td className="py-3">
                <Badge tone="muted">{service.type}</Badge>
              </td>
              <td className="py-3 text-muted-foreground">{service.clusterIP || "-"}</td>
              <td className="py-3 text-right text-muted-foreground">
                {service.ports.map((port) => port.port).join(", ") || "-"}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </ResourceCard>
  );
}

function IngressesTable({ ingresses }: { ingresses: Ingress[] }) {
  return (
    <ResourceCard title="Ingresses" count={ingresses.length}>
      <table className="w-full text-left text-sm">
        <thead>
          <tr className="border-b text-xs uppercase text-muted-foreground">
            <th className="py-2 font-medium">Name</th>
            <th className="py-2 font-medium">Class</th>
            <th className="py-2 font-medium">Hosts</th>
            <th className="py-2 text-right font-medium">TLS</th>
          </tr>
        </thead>
        <tbody>
          {ingresses.map((ingress) => (
            <tr key={`${ingress.namespace}/${ingress.name}`} className="border-b last:border-0">
              <td className="max-w-[14rem] py-3">
                <div className="truncate font-medium">{ingress.name}</div>
                <div className="truncate text-xs text-muted-foreground">{ingress.namespace}</div>
              </td>
              <td className="py-3 text-muted-foreground">{ingress.className || "-"}</td>
              <td className="max-w-[14rem] py-3 text-muted-foreground">
                <div className="truncate">
                  {ingress.rules.map((rule) => rule.host || "*").join(", ") || "-"}
                </div>
              </td>
              <td className="py-3 text-right text-muted-foreground">{ingress.tls.length}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </ResourceCard>
  );
}

function ResourceCard({
  title,
  count,
  detail,
  children
}: {
  title: string;
  count: number;
  detail?: string;
  children: ReactNode;
}) {
  return (
    <Card className="min-w-0 overflow-hidden">
      <CardHeader className="flex flex-row items-center justify-between border-b pb-3">
        <div>
          <CardTitle>{title}</CardTitle>
          {detail ? <p className="mt-1 text-xs text-muted-foreground">{detail}</p> : null}
        </div>
        <Badge tone="muted">{count}</Badge>
      </CardHeader>
      <CardContent className="overflow-x-auto p-4">{count > 0 ? children : <EmptyState />}</CardContent>
    </Card>
  );
}

function ErrorBanner({ message }: { message: string }) {
  return (
    <div className="rounded-lg border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
      {message}
    </div>
  );
}

function LoadingState() {
  return (
    <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: 6 }).map((_, index) => (
        <Card key={index}>
          <CardContent className="p-4">
            <div className="h-4 w-24 rounded bg-muted" />
            <div className="mt-4 h-8 w-16 rounded bg-muted" />
          </CardContent>
        </Card>
      ))}
    </div>
  );
}

function EmptyState() {
  return (
    <div className="flex h-28 items-center justify-center text-sm text-muted-foreground">
      <Box className="mr-2 size-4" />
      No resources
    </div>
  );
}

export default App;
