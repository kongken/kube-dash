package dashboard

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type noParams struct{}

type namespaceParams struct {
	Namespace string `json:"namespace,omitempty" jsonschema:"Kubernetes namespace to filter by. Leave empty to query all namespaces."`
}

func (s *Server) mcpHTTPHandler() http.Handler {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "kube-dash",
		Version: "0.1.0",
	}, &mcp.ServerOptions{
		Instructions: "Read-only Kubernetes dashboard tools for cluster summary and workload discovery.",
	})

	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "cluster_summary",
		Description: "Return Kubernetes cluster counts and readiness summary.",
	}, s.mcpClusterSummary)
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list_namespaces",
		Description: "List Kubernetes namespaces.",
	}, s.mcpListNamespaces)
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list_nodes",
		Description: "List Kubernetes nodes and readiness information.",
	}, s.mcpListNodes)
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list_pods",
		Description: "List Kubernetes pods. Optionally filter by namespace.",
	}, s.mcpListPods)
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list_deployments",
		Description: "List Kubernetes deployments. Optionally filter by namespace.",
	}, s.mcpListDeployments)
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list_services",
		Description: "List Kubernetes services. Optionally filter by namespace.",
	}, s.mcpListServices)
	mcp.AddTool(mcpServer, &mcp.Tool{
		Name:        "list_ingresses",
		Description: "List Kubernetes ingresses. Optionally filter by namespace.",
	}, s.mcpListIngresses)

	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return mcpServer
	}, &mcp.StreamableHTTPOptions{
		Stateless:    true,
		JSONResponse: true,
	})
}

func (s *Server) mcpClient() (KubernetesClient, error) {
	if s.client == nil {
		return nil, fmt.Errorf("kubernetes client is not initialized")
	}
	return s.client, nil
}

func (s *Server) mcpClusterSummary(ctx context.Context, _ *mcp.CallToolRequest, _ noParams) (*mcp.CallToolResult, gin.H, error) {
	client, err := s.mcpClient()
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	pods, err := client.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	deployments, err := client.AppsV1().Deployments(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	services, err := client.CoreV1().Services(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}
	ingresses, err := client.NetworkingV1().Ingresses(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	return nil, gin.H{
		"namespaces":  len(namespaces.Items),
		"nodes":       nodeSummary(nodes.Items),
		"pods":        podSummary(pods.Items),
		"deployments": len(deployments.Items),
		"services":    len(services.Items),
		"ingresses":   len(ingresses.Items),
	}, nil
}

func (s *Server) mcpListNamespaces(ctx context.Context, _ *mcp.CallToolRequest, _ noParams) (*mcp.CallToolResult, gin.H, error) {
	client, err := s.mcpClient()
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	list, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	items := make([]gin.H, 0, len(list.Items))
	for _, ns := range list.Items {
		items = append(items, gin.H{
			"name":       ns.Name,
			"status":     ns.Status.Phase,
			"ageSeconds": ageSeconds(ns.CreationTimestamp.Time),
		})
	}
	sortByName(items)
	return nil, gin.H{"items": items}, nil
}

func (s *Server) mcpListNodes(ctx context.Context, _ *mcp.CallToolRequest, _ noParams) (*mcp.CallToolResult, gin.H, error) {
	client, err := s.mcpClient()
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	list, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	items := make([]gin.H, 0, len(list.Items))
	for _, node := range list.Items {
		items = append(items, gin.H{
			"name":       node.Name,
			"ready":      nodeReady(node),
			"roles":      nodeRoles(node),
			"version":    node.Status.NodeInfo.KubeletVersion,
			"ageSeconds": ageSeconds(node.CreationTimestamp.Time),
		})
	}
	sortByName(items)
	return nil, gin.H{"items": items}, nil
}

func (s *Server) mcpListPods(ctx context.Context, _ *mcp.CallToolRequest, params namespaceParams) (*mcp.CallToolResult, gin.H, error) {
	client, err := s.mcpClient()
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	list, err := client.CoreV1().Pods(namespaceOrAll(params.Namespace)).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	items := make([]gin.H, 0, len(list.Items))
	for _, pod := range list.Items {
		items = append(items, gin.H{
			"name":       pod.Name,
			"namespace":  pod.Namespace,
			"node":       pod.Spec.NodeName,
			"phase":      pod.Status.Phase,
			"ready":      readyContainers(pod),
			"restarts":   restartCount(pod),
			"ageSeconds": ageSeconds(pod.CreationTimestamp.Time),
		})
	}
	sortByNamespaceAndName(items)
	return nil, gin.H{"items": items}, nil
}

func (s *Server) mcpListDeployments(ctx context.Context, _ *mcp.CallToolRequest, params namespaceParams) (*mcp.CallToolResult, gin.H, error) {
	client, err := s.mcpClient()
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	list, err := client.AppsV1().Deployments(namespaceOrAll(params.Namespace)).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	items := make([]gin.H, 0, len(list.Items))
	for _, deployment := range list.Items {
		items = append(items, deploymentResponse(deployment))
	}
	sortByNamespaceAndName(items)
	return nil, gin.H{"items": items}, nil
}

func (s *Server) mcpListServices(ctx context.Context, _ *mcp.CallToolRequest, params namespaceParams) (*mcp.CallToolResult, gin.H, error) {
	client, err := s.mcpClient()
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	list, err := client.CoreV1().Services(namespaceOrAll(params.Namespace)).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	items := make([]gin.H, 0, len(list.Items))
	for _, service := range list.Items {
		items = append(items, serviceResponse(service))
	}
	sortByNamespaceAndName(items)
	return nil, gin.H{"items": items}, nil
}

func (s *Server) mcpListIngresses(ctx context.Context, _ *mcp.CallToolRequest, params namespaceParams) (*mcp.CallToolResult, gin.H, error) {
	client, err := s.mcpClient()
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithTimeout(ctx, requestTimeout)
	defer cancel()

	list, err := client.NetworkingV1().Ingresses(namespaceOrAll(params.Namespace)).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	items := make([]gin.H, 0, len(list.Items))
	for _, ingress := range list.Items {
		items = append(items, ingressResponse(ingress))
	}
	sortByNamespaceAndName(items)
	return nil, gin.H{"items": items}, nil
}

func namespaceOrAll(namespace string) string {
	if namespace == "" {
		return metav1.NamespaceAll
	}
	return namespace
}
