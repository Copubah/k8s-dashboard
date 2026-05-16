package api

type ErrorResponse struct {
	Error string `json:"error"`
}

type PodResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Node      string `json:"node"`
	Restarts  int32  `json:"restarts"`
	Age       string `json:"age"`
}

type DeploymentResponse struct {
	Name              string `json:"name"`
	Namespace         string `json:"namespace"`
	Replicas          int32  `json:"replicas"`
	ReadyReplicas     int32  `json:"readyReplicas"`
	AvailableReplicas int32  `json:"availableReplicas"`
	UpdatedReplicas   int32  `json:"updatedReplicas"`
	Age               string `json:"age"`
}

type ScaleRequest struct {
	Replicas *int32 `json:"replicas" binding:"required,min=0"`
}

type LogsResponse struct {
	Pod       string `json:"pod"`
	Namespace string `json:"namespace"`
	Logs      string `json:"logs"`
}
