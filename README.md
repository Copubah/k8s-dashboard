# Kubernetes Deployment Dashboard

A Kubernetes dashboard built with a Go/Gin backend, Kubernetes client-go, and a React frontend.

---

## Problem

Managing Kubernetes workloads using only kubectl is slow, repetitive, and error-prone, especially when handling multiple services and namespaces.

## Solution

This project provides a web-based Kubernetes dashboard that simplifies cluster operations through a REST API and UI, allowing users to view, manage, and control workloads visually.

---

## Tech Stack

- Backend: Go, Gin
- Kubernetes: client-go
- Frontend: React
- Auth: Basic Authentication
- Containerization: Docker
- Cluster: Kubernetes (Minikube / EKS compatible)

---

## Key Features

- Lists pods and deployments from your Kubernetes cluster
- Displays pod status, restart counts, and deployment replicas
- View pod logs directly from the UI
- Scale deployments dynamically
- Restart deployments using rollout strategy
- Basic authentication for API access

---

## Project Structure

    backend/   Go, Gin, client-go REST API
    frontend/  React dashboard UI
    k8s/       Kubernetes manifests and RBAC

---

## Environment Variables

### Backend

| Name               | Default               | Description                     |
|--------------------|-----------------------|---------------------------------|
| PORT               | 8080                  | API port                        |
| K8S_NAMESPACE      | empty                 | Default namespace (empty = all) |
| KUBECONFIG         | $HOME/.kube/config    | Local kubeconfig path           |
| KUBE_CONTEXT       | empty                 | Kubernetes context              |
| IN_CLUSTER         | false                 | Use in-cluster service account  |
| DASHBOARD_USERNAME | admin                 | Basic auth username             |
| DASHBOARD_PASSWORD | change-me             | Basic auth password             |
| ALLOWED_ORIGIN     | http://localhost:5173 | Allowed CORS origin             |

### Frontend

| Name         | Default               | Description          |
|--------------|-----------------------|----------------------|
| VITE_API_URL | http://localhost:8080 | Backend API base URL |

---

## Architecture

    Frontend -> Go API -> Kubernetes API Server -> Cluster Resources

---

## Run Locally With Minikube

1. Start Minikube:

        minikube start
        kubectl config use-context minikube

2. Start the application:

        docker compose up --build

3. Open the dashboard at http://localhost:5173

4. Sign in with username admin and password change-me

---

## Run Backend Without Docker

    cd backend
    cp .env.example .env
    go mod download
    PORT=8080 KUBE_CONTEXT=minikube DASHBOARD_PASSWORD=change-me go run ./cmd/server

---

## Run Frontend Without Docker

    cd frontend
    cp .env.example .env
    npm install
    npm run dev

---

## Deploy to Minikube

Build images inside Minikube Docker environment:

    eval $(minikube docker-env)
    docker build -t k8s-dashboard-backend:latest ./backend
    docker build --build-arg VITE_API_URL=http://localhost:8080 -t k8s-dashboard-frontend:latest ./frontend
    kubectl apply -k k8s

Port forward the services:

    kubectl -n k8s-dashboard port-forward service/dashboard-backend 8080:8080
    kubectl -n k8s-dashboard port-forward service/dashboard-frontend 5173:80

---

## API Endpoints

All endpoints require Basic Auth.

### List Pods

    GET /api/pods?namespace=default

### List Deployments

    GET /api/deployments?namespace=default

### View Pod Logs

    GET /api/pods/{namespace}/{pod}/logs?tailLines=100

### Scale Deployment

    POST /api/deployments/{namespace}/{name}/scale
    { "replicas": 3 }

### Restart Deployment

    POST /api/deployments/{namespace}/{name}/restart

---

## Authentication Notes

This project uses Basic Authentication for simplicity. In production, use TLS encryption, OIDC authentication, and strong RBAC policies.

---

## Best Practices Included

- Environment-based configuration
- Kubernetes access via client-go
- RBAC for cluster security
- Docker multi-stage builds
- Separation of frontend and backend
- Structured API error responses

---

## Future Improvements

- RBAC-based login system
- Multi-cluster support
- Live logs with WebSockets
- Metrics dashboard (CPU/Memory usage)
- Helm chart deployment
- Audit logging system
