package dashboard

import (
	"context"
	"errors"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const requestTimeout = 10 * time.Second

type KubernetesClient interface {
	kubernetes.Interface
}

type Server struct {
	client KubernetesClient
}

func NewServer() *Server {
	return &Server{}
}

func NewServerWithClient(client KubernetesClient) *Server {
	return &Server{client: client}
}

func (s *Server) Init() error {
	if s.client != nil {
		return nil
	}

	config, err := kubernetesConfig()
	if err != nil {
		return err
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	s.client = client
	return nil
}

func (s *Server) Close() error {
	return nil
}

func (s *Server) RegisterRoutes(r *gin.Engine) {
	r.GET("/healthz", s.healthz)

	api := r.Group("/api/v1")
	api.GET("/summary", s.summary)
	api.GET("/namespaces", s.namespaces)
	api.GET("/nodes", s.nodes)
	api.GET("/pods", s.pods)
	api.GET("/deployments", s.deployments)
	api.GET("/services", s.services)
}

func kubernetesConfig() (*rest.Config, error) {
	if kubeconfig := os.Getenv("KUBE_DASH_KUBECONFIG"); kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func (s *Server) requireClient(c *gin.Context) (KubernetesClient, bool) {
	if s.client == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "kubernetes client is not initialized"})
		return nil, false
	}
	return s.client, true
}

func requestContext(c *gin.Context) (context.Context, context.CancelFunc) {
	return context.WithTimeout(c.Request.Context(), requestTimeout)
}

func namespaceParam(c *gin.Context) string {
	if namespace := c.Query("namespace"); namespace != "" {
		return namespace
	}
	return metav1.NamespaceAll
}

func (s *Server) healthz(c *gin.Context) {
	status := gin.H{"status": "ok"}
	if s.client == nil {
		status["kubernetes"] = "not_initialized"
	} else {
		status["kubernetes"] = "initialized"
	}
	c.JSON(http.StatusOK, status)
}

func (s *Server) summary(c *gin.Context) {
	client, ok := s.requireClient(c)
	if !ok {
		return
	}

	ctx, cancel := requestContext(c)
	defer cancel()

	namespaces, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		writeError(c, err)
		return
	}
	nodes, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		writeError(c, err)
		return
	}
	pods, err := client.CoreV1().Pods(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		writeError(c, err)
		return
	}
	deployments, err := client.AppsV1().Deployments(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		writeError(c, err)
		return
	}
	services, err := client.CoreV1().Services(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		writeError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"namespaces":  len(namespaces.Items),
		"nodes":       nodeSummary(nodes.Items),
		"pods":        podSummary(pods.Items),
		"deployments": len(deployments.Items),
		"services":    len(services.Items),
	})
}

func (s *Server) namespaces(c *gin.Context) {
	client, ok := s.requireClient(c)
	if !ok {
		return
	}

	ctx, cancel := requestContext(c)
	defer cancel()

	list, err := client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		writeError(c, err)
		return
	}

	items := make([]gin.H, 0, len(list.Items))
	for _, ns := range list.Items {
		items = append(items, gin.H{
			"name":       ns.Name,
			"status":     ns.Status.Phase,
			"ageSeconds": ageSeconds(ns.CreationTimestamp.Time),
		})
	}
	sort.Slice(items, func(i, j int) bool { return items[i]["name"].(string) < items[j]["name"].(string) })
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) nodes(c *gin.Context) {
	client, ok := s.requireClient(c)
	if !ok {
		return
	}

	ctx, cancel := requestContext(c)
	defer cancel()

	list, err := client.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		writeError(c, err)
		return
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
	sort.Slice(items, func(i, j int) bool { return items[i]["name"].(string) < items[j]["name"].(string) })
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) pods(c *gin.Context) {
	client, ok := s.requireClient(c)
	if !ok {
		return
	}

	ctx, cancel := requestContext(c)
	defer cancel()

	list, err := client.CoreV1().Pods(namespaceParam(c)).List(ctx, metav1.ListOptions{})
	if err != nil {
		writeError(c, err)
		return
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
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) deployments(c *gin.Context) {
	client, ok := s.requireClient(c)
	if !ok {
		return
	}

	ctx, cancel := requestContext(c)
	defer cancel()

	list, err := client.AppsV1().Deployments(namespaceParam(c)).List(ctx, metav1.ListOptions{})
	if err != nil {
		writeError(c, err)
		return
	}

	items := make([]gin.H, 0, len(list.Items))
	for _, deployment := range list.Items {
		items = append(items, deploymentResponse(deployment))
	}
	sortByNamespaceAndName(items)
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (s *Server) services(c *gin.Context) {
	client, ok := s.requireClient(c)
	if !ok {
		return
	}

	ctx, cancel := requestContext(c)
	defer cancel()

	list, err := client.CoreV1().Services(namespaceParam(c)).List(ctx, metav1.ListOptions{})
	if err != nil {
		writeError(c, err)
		return
	}

	items := make([]gin.H, 0, len(list.Items))
	for _, service := range list.Items {
		items = append(items, serviceResponse(service))
	}
	sortByNamespaceAndName(items)
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func writeError(c *gin.Context, err error) {
	if errors.Is(err, context.DeadlineExceeded) {
		c.JSON(http.StatusGatewayTimeout, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

func ageSeconds(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return int64(time.Since(t).Seconds())
}

func sortByNamespaceAndName(items []gin.H) {
	sort.Slice(items, func(i, j int) bool {
		left := items[i]["namespace"].(string) + "/" + items[i]["name"].(string)
		right := items[j]["namespace"].(string) + "/" + items[j]["name"].(string)
		return left < right
	})
}

func nodeSummary(nodes []corev1.Node) gin.H {
	ready := 0
	for _, node := range nodes {
		if nodeReady(node) {
			ready++
		}
	}
	return gin.H{"total": len(nodes), "ready": ready}
}

func podSummary(pods []corev1.Pod) gin.H {
	phases := map[corev1.PodPhase]int{}
	for _, pod := range pods {
		phases[pod.Status.Phase]++
	}
	return gin.H{
		"total":     len(pods),
		"running":   phases[corev1.PodRunning],
		"pending":   phases[corev1.PodPending],
		"succeeded": phases[corev1.PodSucceeded],
		"failed":    phases[corev1.PodFailed],
		"unknown":   phases[corev1.PodUnknown],
	}
}

func nodeReady(node corev1.Node) bool {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

func nodeRoles(node corev1.Node) []string {
	roles := []string{}
	for key := range node.Labels {
		const prefix = "node-role.kubernetes.io/"
		if role, ok := strings.CutPrefix(key, prefix); ok {
			if role == "" {
				role = "node"
			}
			roles = append(roles, role)
		}
	}
	sort.Strings(roles)
	return roles
}

func readyContainers(pod corev1.Pod) gin.H {
	ready := 0
	total := len(pod.Status.ContainerStatuses)
	for _, status := range pod.Status.ContainerStatuses {
		if status.Ready {
			ready++
		}
	}
	return gin.H{"ready": ready, "total": total}
}

func restartCount(pod corev1.Pod) int32 {
	var restarts int32
	for _, status := range pod.Status.ContainerStatuses {
		restarts += status.RestartCount
	}
	return restarts
}

func deploymentResponse(deployment appsv1.Deployment) gin.H {
	return gin.H{
		"name":       deployment.Name,
		"namespace":  deployment.Namespace,
		"replicas":   deployment.Status.Replicas,
		"ready":      deployment.Status.ReadyReplicas,
		"available":  deployment.Status.AvailableReplicas,
		"updated":    deployment.Status.UpdatedReplicas,
		"ageSeconds": ageSeconds(deployment.CreationTimestamp.Time),
	}
}

func serviceResponse(service corev1.Service) gin.H {
	ports := make([]gin.H, 0, len(service.Spec.Ports))
	for _, port := range service.Spec.Ports {
		ports = append(ports, gin.H{
			"name":       port.Name,
			"port":       port.Port,
			"targetPort": port.TargetPort.String(),
			"protocol":   port.Protocol,
		})
	}

	return gin.H{
		"name":       service.Name,
		"namespace":  service.Namespace,
		"type":       service.Spec.Type,
		"clusterIP":  service.Spec.ClusterIP,
		"ports":      ports,
		"ageSeconds": ageSeconds(service.CreationTimestamp.Time),
	}
}
