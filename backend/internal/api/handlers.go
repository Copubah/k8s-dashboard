package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"k8s-dashboard/backend/internal/kube"
)

type Handlers struct {
	client *kube.Client
}

func NewHandlers(client *kube.Client) *Handlers {
	return &Handlers{client: client}
}

func (h *Handlers) ListPods(c *gin.Context) {
	namespace := h.namespace(c)
	pods, err := h.client.Set.CoreV1().Pods(namespace).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		respondError(c, http.StatusBadGateway, fmt.Errorf("list pods: %w", err))
		return
	}

	response := make([]PodResponse, 0, len(pods.Items))
	for _, pod := range pods.Items {
		response = append(response, podResponse(pod))
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handlers) ListDeployments(c *gin.Context) {
	namespace := h.namespace(c)
	deployments, err := h.client.Set.AppsV1().Deployments(namespace).List(c.Request.Context(), metav1.ListOptions{})
	if err != nil {
		respondError(c, http.StatusBadGateway, fmt.Errorf("list deployments: %w", err))
		return
	}

	response := make([]DeploymentResponse, 0, len(deployments.Items))
	for _, deployment := range deployments.Items {
		response = append(response, deploymentResponse(deployment))
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handlers) PodLogs(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	tailLines := int64(200)
	if raw := c.Query("tailLines"); raw != "" {
		parsed, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || parsed < 1 {
			respondError(c, http.StatusBadRequest, fmt.Errorf("tailLines must be a positive number"))
			return
		}
		tailLines = parsed
	}

	logs, err := h.client.Set.CoreV1().Pods(namespace).GetLogs(name, &corev1.PodLogOptions{
		TailLines: &tailLines,
	}).DoRaw(c.Request.Context())
	if err != nil {
		respondError(c, http.StatusBadGateway, fmt.Errorf("get pod logs: %w", err))
		return
	}

	c.JSON(http.StatusOK, LogsResponse{Pod: name, Namespace: namespace, Logs: string(logs)})
}

func (h *Handlers) ScaleDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")

	var request ScaleRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		respondError(c, http.StatusBadRequest, fmt.Errorf("invalid scale request: %w", err))
		return
	}

	scale, err := h.client.Set.AppsV1().Deployments(namespace).GetScale(c.Request.Context(), name, metav1.GetOptions{})
	if err != nil {
		respondError(c, http.StatusBadGateway, fmt.Errorf("get deployment scale: %w", err))
		return
	}

	scale.Spec.Replicas = *request.Replicas
	updated, err := h.client.Set.AppsV1().Deployments(namespace).UpdateScale(c.Request.Context(), name, scale, metav1.UpdateOptions{})
	if err != nil {
		respondError(c, http.StatusBadGateway, fmt.Errorf("update deployment scale: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":      name,
		"namespace": namespace,
		"replicas":  updated.Spec.Replicas,
	})
}

func (h *Handlers) RestartDeployment(c *gin.Context) {
	namespace := c.Param("namespace")
	name := c.Param("name")
	now := time.Now().UTC().Format(time.RFC3339)

	// Updating this annotation mirrors `kubectl rollout restart deployment/name`.
	patch := []byte(fmt.Sprintf(`{"spec":{"template":{"metadata":{"annotations":{"kubectl.kubernetes.io/restartedAt":"%s"}}}}}`, now))
	deployment, err := h.client.Set.AppsV1().Deployments(namespace).Patch(
		c.Request.Context(),
		name,
		types.StrategicMergePatchType,
		patch,
		metav1.PatchOptions{},
	)
	if err != nil {
		respondError(c, http.StatusBadGateway, fmt.Errorf("restart deployment: %w", err))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"name":        deployment.Name,
		"namespace":   deployment.Namespace,
		"restartedAt": now,
	})
}

func (h *Handlers) namespace(c *gin.Context) string {
	if namespace := c.Query("namespace"); namespace != "" {
		return namespace
	}
	return h.client.Namespace
}

func podResponse(pod corev1.Pod) PodResponse {
	var restarts int32
	for _, status := range pod.Status.ContainerStatuses {
		restarts += status.RestartCount
	}

	return PodResponse{
		Name:      pod.Name,
		Namespace: pod.Namespace,
		Status:    string(pod.Status.Phase),
		Node:      pod.Spec.NodeName,
		Restarts:  restarts,
		Age:       age(pod.CreationTimestamp.Time),
	}
}

func deploymentResponse(deployment appsv1.Deployment) DeploymentResponse {
	replicas := int32(0)
	if deployment.Spec.Replicas != nil {
		replicas = *deployment.Spec.Replicas
	}

	return DeploymentResponse{
		Name:              deployment.Name,
		Namespace:         deployment.Namespace,
		Replicas:          replicas,
		ReadyReplicas:     deployment.Status.ReadyReplicas,
		AvailableReplicas: deployment.Status.AvailableReplicas,
		UpdatedReplicas:   deployment.Status.UpdatedReplicas,
		Age:               age(deployment.CreationTimestamp.Time),
	}
}

func age(created time.Time) string {
	if created.IsZero() {
		return "unknown"
	}
	return time.Since(created).Round(time.Second).String()
}

func respondError(c *gin.Context, status int, err error) {
	c.JSON(status, ErrorResponse{Error: err.Error()})
}
