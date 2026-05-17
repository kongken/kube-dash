package dashboard

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
