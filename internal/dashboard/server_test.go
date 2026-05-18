package dashboard

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestSummary(t *testing.T) {
	router := testRouter(
		&corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "default"}},
		&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{Name: "node-1"},
			Status: corev1.NodeStatus{Conditions: []corev1.NodeCondition{
				{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
			}},
		},
		&corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "pod-1", Namespace: "default"},
			Status:     corev1.PodStatus{Phase: corev1.PodRunning},
		},
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"}},
		&corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"}},
		&networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"}},
	)

	response := performRequest(router, http.MethodGet, "/api/v1/summary")
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["namespaces"].(float64) != 1 {
		t.Fatalf("expected one namespace, got %#v", body["namespaces"])
	}
	if body["ingresses"].(float64) != 1 {
		t.Fatalf("expected one ingress, got %#v", body["ingresses"])
	}
}

func TestNamespaceFilterForPods(t *testing.T) {
	router := testRouter(
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "frontend", Namespace: "default"}},
		&corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "scheduler", Namespace: "kube-system"}},
	)

	response := performRequest(router, http.MethodGet, "/api/v1/pods?namespace=default")
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var body struct {
		Items []struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		} `json:"items"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Items) != 1 || body.Items[0].Name != "frontend" || body.Items[0].Namespace != "default" {
		t.Fatalf("unexpected pod list: %#v", body.Items)
	}
}

func TestNamespaceFilterForIngresses(t *testing.T) {
	pathType := networkingv1.PathTypePrefix
	router := testRouter(
		&networkingv1.Ingress{
			ObjectMeta: metav1.ObjectMeta{Name: "frontend", Namespace: "default"},
			Spec: networkingv1.IngressSpec{Rules: []networkingv1.IngressRule{
				{
					Host: "dash.example.com",
					IngressRuleValue: networkingv1.IngressRuleValue{HTTP: &networkingv1.HTTPIngressRuleValue{
						Paths: []networkingv1.HTTPIngressPath{
							{
								Path:     "/",
								PathType: &pathType,
								Backend: networkingv1.IngressBackend{Service: &networkingv1.IngressServiceBackend{
									Name: "frontend",
									Port: networkingv1.ServiceBackendPort{Number: 80},
								}},
							},
						},
					}},
				},
			}},
		},
		&networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: "scheduler", Namespace: "kube-system"}},
	)

	response := performRequest(router, http.MethodGet, "/api/v1/ingresses?namespace=default")
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var body struct {
		Items []struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
			Rules     []struct {
				Host  string `json:"host"`
				Paths []struct {
					Path        string `json:"path"`
					PathType    string `json:"pathType"`
					ServiceName string `json:"serviceName"`
					ServicePort string `json:"servicePort"`
				} `json:"paths"`
			} `json:"rules"`
		} `json:"items"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Items) != 1 || body.Items[0].Name != "frontend" || body.Items[0].Namespace != "default" {
		t.Fatalf("unexpected ingress list: %#v", body.Items)
	}
	if body.Items[0].Rules[0].Host != "dash.example.com" {
		t.Fatalf("unexpected ingress host: %#v", body.Items[0].Rules)
	}
	if body.Items[0].Rules[0].Paths[0].ServiceName != "frontend" || body.Items[0].Rules[0].Paths[0].ServicePort != "80" {
		t.Fatalf("unexpected ingress backend: %#v", body.Items[0].Rules[0].Paths)
	}
}

func TestMCPListTools(t *testing.T) {
	router := testRouter()

	request := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader(`{"jsonrpc":"2.0","id":1,"method":"tools/list","params":{}}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, response.Code, response.Body.String())
	}

	var body struct {
		Result struct {
			Tools []struct {
				Name string `json:"name"`
			} `json:"tools"`
		} `json:"result"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Result.Tools) == 0 {
		t.Fatalf("expected MCP tools, got %#v", body)
	}
	found := false
	for _, tool := range body.Result.Tools {
		if tool.Name == "cluster_summary" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected cluster_summary tool, got %#v", body.Result.Tools)
	}
}

func testRouter(objects ...runtime.Object) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	server := NewServerWithClient(fake.NewSimpleClientset(objects...))
	server.RegisterRoutes(router)
	return router
}

func performRequest(router http.Handler, method string, path string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}
