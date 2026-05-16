# Kubernetes Deployment Dashboard

A Kubernetes dashboard built with a Go/Gin backend, Kubernetes `client-go`, and a React frontend.

## What It Does

- Lists pods and deployments from your Kubernetes cluster.
- Shows pod status, restart counts, deployment replica counts, and ready replicas.
- Lets you view pod logs.
- Lets you scale deployments.
- Lets you restart deployments by updating the pod template annotation, the same idea used by `kubectl rollout restart`.
- Protects the API with HTTP Basic Authentication.

## Project Structure

```text
backend/   Go, Gin, client-go REST API
frontend/  React dashboard UI
k8s/       Kubernetes manifests and RBAC
```

## Environment Variables

Backend:

| Name | Default | Meaning |
| --- | --- | --- |
| `PORT` | `8080` | API port |
| `K8S_NAMESPACE` | empty | Namespace to read by default. Empty means all namespaces |
| `KUBECONFIG` | `$HOME/.kube/config` | Local kubeconfig path |
| `KUBE_CONTEXT` | empty | Optional kubeconfig context, for example `minikube` |
| `IN_CLUSTER` | `false` | Use Kubernetes service account credentials inside a pod |
| `DASHBOARD_USERNAME` | `admin` | Basic auth username |
| `DASHBOARD_PASSWORD` | `change-me` | Basic auth password |
| `ALLOWED_ORIGIN` | `http://localhost:5173` | Frontend origin allowed by CORS |

Frontend:

| Name | Default | Meaning |
| --- | --- | --- |
| `VITE_API_URL` | `http://localhost:8080` | Backend API base URL |

## Run Locally With Minikube

1. Start Minikube:

```bash
minikube start
kubectl config use-context minikube
```

2. Start the app with Docker Compose:

```bash
docker compose up --build
```

3. Open the dashboard:

```text
http://localhost:5173
```

4. Sign in with:

```text
username: admin
password: change-me
```

The backend reads your local kubeconfig through the Docker Compose volume mount.
The Compose file also mounts `~/.minikube` because Minikube kubeconfigs usually reference certificate files from that directory.

## Run Backend Without Docker

```bash
cd backend
cp .env.example .env
go mod download
PORT=8080 KUBE_CONTEXT=minikube DASHBOARD_PASSWORD=change-me go run ./cmd/server
```

## Run Frontend Without Docker

```bash
cd frontend
cp .env.example .env
npm install
npm run dev
```

## Deploy Into Minikube

Build the images inside Minikube's Docker environment:

```bash
eval $(minikube docker-env)
docker build -t k8s-dashboard-backend:latest ./backend
docker build --build-arg VITE_API_URL=http://localhost:8080 -t k8s-dashboard-frontend:latest ./frontend
kubectl apply -k k8s
```

Port-forward the services:

```bash
kubectl -n k8s-dashboard port-forward service/dashboard-backend 8080:8080
kubectl -n k8s-dashboard port-forward service/dashboard-frontend 5173:80
```

Then open `http://localhost:5173`.

## API Endpoints

All `/api` endpoints require Basic Auth.

### List Pods

```http
GET /api/pods?namespace=default
```

Sample response:

```json
[
  {
    "name": "nginx-7854ff8877-xk9q2",
    "namespace": "default",
    "status": "Running",
    "node": "minikube",
    "restarts": 0,
    "age": "12m30s"
  }
]
```

### List Deployments

```http
GET /api/deployments?namespace=default
```

Sample response:

```json
[
  {
    "name": "nginx",
    "namespace": "default",
    "replicas": 2,
    "readyReplicas": 2,
    "availableReplicas": 2,
    "updatedReplicas": 2,
    "age": "14m5s"
  }
]
```

### View Pod Logs

```http
GET /api/pods/default/nginx-7854ff8877-xk9q2/logs?tailLines=100
```

Sample response:

```json
{
  "pod": "nginx-7854ff8877-xk9q2",
  "namespace": "default",
  "logs": "10.244.0.1 - - [16/May/2026:10:00:00 +0000] \"GET / HTTP/1.1\" 200 615"
}
```

### Scale Deployment

```http
POST /api/deployments/default/nginx/scale
Content-Type: application/json

{ "replicas": 3 }
```

Sample response:

```json
{
  "name": "nginx",
  "namespace": "default",
  "replicas": 3
}
```

### Restart Deployment

```http
POST /api/deployments/default/nginx/restart
```

Sample response:

```json
{
  "name": "nginx",
  "namespace": "default",
  "restartedAt": "2026-05-16T10:00:00Z"
}
```

## Authentication Notes

This project uses Basic Auth to keep the example easy to understand. For production, use TLS, strong secrets, and an identity provider such as OIDC. Keep Kubernetes RBAC as narrow as your users need.

## Best Practices Included

- Configuration comes from environment variables.
- Kubernetes access uses `client-go`.
- In-cluster and local kubeconfig modes are both supported.
- Kubernetes permissions are declared with RBAC.
- API errors return JSON with clear messages.
- Dockerfiles use multi-stage builds.
- Frontend avoids hard-coded cluster data and talks only to the backend API.
