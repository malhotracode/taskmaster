In the root `taskmaster-go` directory, create `README.md`:

```markdown
# TaskMaster Go - End-to-End Backend Project

This project demonstrates a complete end-to-end backend system for a simple Task Management API ("TaskMaster Go") built with Go. It includes:

- RESTful API (CRUD operations for tasks)
- PostgreSQL database
- Dockerized services for local development (Go app, PostgreSQL, OpenTelemetry Collector, Prometheus, Grafana)
- Kubernetes deployment manifests (for Minikube)
- Observability stack integration using OpenTelemetry, Prometheus, and Grafana.

## Architecture Overview
```

+-----------------+ +-----------------+ +-----------------+
| User / Client |----->| Load Balancer |----->| Go API Service |
| (e.g., curl) | | (K8s Service) | | (Kubernetes Pods)|
+-----------------+ +-----------------+ +-----------------+
+-----------------+ |
| (OTLP Traces & Metrics)
v
+---------------------+
| OpenTelemetry |
| Collector (K8s Svc) |
+---------------------+
/ \
 / \ (Prometheus format metrics)
v v
+----------------------+ (SQL) +-----------------+ (Scrape) +-----------------+
| PostgreSQL Database |<--------->| Go API Service |<----------| Prometheus |
| (K8s StatefulSet) | | (Business Logic,| | (K8s Service) |
+----------------------+ | DB Access) | +-----------------+
+-----------------+ |
| (Data Source)
v
+-----------------+
| Grafana |
| (K8s Service) |
+-----------------+

```

## Prerequisites

*   Go (version 1.20 or higher recommended)
*   Docker and Docker Compose
*   Minikube (or any Kubernetes cluster)
*   `kubectl` command-line tool

## Project Structure

```

taskmaster-go/
├── go-app/ # Go application source code
│ ├── main.go # Entry point, HTTP server
│ ├── handlers.go # HTTP request handlers
│ ├── store.go # Database interaction logic
│ ├── models.go # Data structures (Task model)
│ ├── telemetry.go # OpenTelemetry setup
│ ├── go.mod, go.sum # Go module files
│ └── Dockerfile # Dockerfile for the Go application
├── kubernetes/ # Kubernetes manifest YAML files
│ ├── namespace.yaml
│ ├── postgres-secret.yaml # Placeholder, create manually
│ ├── postgres-pvc.yaml
│ ├── postgres-statefulset.yaml
│ ├── postgres-service.yaml
│ ├── app-configmap.yaml
│ ├── app-deployment.yaml
│ ├── app-service.yaml
│ ├── otel-collector-config.yaml
│ ├── otel-collector-deployment.yaml
│ ├── otel-collector-service.yaml
│ ├── prometheus-rbac.yaml # Optional RBAC for Prometheus
│ ├── prometheus-config.yaml
│ ├── prometheus-deployment.yaml
│ ├── prometheus-service.yaml
│ ├── grafana-datasources-config.yaml
│ ├── grafana-deployment.yaml
│ └── grafana-service.yaml
├── docker-compose.yml # Docker Compose for local development environment
└── README.md # This file

````

## Setup and Running

### 1. Local Development with Docker Compose

This method runs the entire stack (Go app, PostgreSQL, OTel Collector, Prometheus, Grafana) locally using Docker.

**a. (Optional) Create `.env` file for Go app:**
   In the `go-app/` directory, you can create a `.env` file for local Go development outside Docker (though Docker Compose uses its own environment variables).
   ```env
   # go-app/.env (example)
   PORT=8080
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=taskuser
   DB_PASSWORD=taskpassword
   DB_NAME=taskdb
   OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317 # For local OTel collector if run separately
   APP_ENV=development
````

The `docker-compose.yml` overrides these for containerized environment.

**b. Fill Placeholder Configs (if not done during walkthrough):**
The `docker-compose.yml` file references configuration files from the `kubernetes/` directory for the observability stack. Ensure these files exist and have content (even if minimal for Docker Compose initially, they are fully defined for K8s):

- `kubernetes/otel-collector-config.yaml`
- `kubernetes/prometheus-config.yaml`
- `kubernetes/grafana-datasources-config.yaml`

**c. Start Services:**
Navigate to the project root (`taskmaster-go/`) and run:

```bash
docker-compose up --build -d
```

**d. Access Services (Local Docker Compose):**

- **Go Application API:** `http://localhost:8080` (e.g., `http://localhost:8080/tasks`)
- **Prometheus:** `http://localhost:9090`
- **Grafana:** `http://localhost:3000` (Login: `admin` / `admin`)

**e. Stop Services:**

```bash
docker-compose down
```

### 2. Deployment to Kubernetes (Minikube)

**a. Start Minikube:**

```bash
minikube start --cpus=4 --memory=4096 # Or your preferred settings
```

**(Optional) Use Minikube's Docker daemon for local image builds:**

```bash
eval $(minikube docker-env)
# If you do this, build your Go app image with a simple name, e.g., `taskmaster-go:latest`
# And set `imagePullPolicy: IfNotPresent` or `Never` in `kubernetes/app-deployment.yaml`.
```

**b. Build and Push/Load Docker Image for the Go App:**
Navigate to `go-app/`:

- **If using Minikube's Docker daemon:**
  ```bash
  docker build -t taskmaster-go:latest .
  # Ensure kubernetes/app-deployment.yaml uses image: taskmaster-go:latest
  # and imagePullPolicy: IfNotPresent or Never
  ```
- **If pushing to a Docker registry (e.g., Docker Hub):**
  ```bash
  docker build -t your-dockerhub-username/taskmaster-go:latest .
  docker push your-dockerhub-username/taskmaster-go:latest
  # Ensure kubernetes/app-deployment.yaml uses image: your-dockerhub-username/taskmaster-go:latest
  # and imagePullPolicy: Always (or IfNotPresent)
  ```
  Return to the project root (`cd ..`).

**c. Apply Kubernetes Manifests:**
Apply the manifests in order:

```bash
# 1. Namespace
kubectl apply -f kubernetes/namespace.yaml

# 2. PostgreSQL Secret (Manual Step - MUST BE DONE FIRST)
kubectl create secret generic postgres-secret \
  --from-literal=POSTGRES_USER=taskuser \
  --from-literal=POSTGRES_PASSWORD=taskpassword \
  --from-literal=POSTGRES_DB=taskdb \
  -n taskmaster

# 3. PostgreSQL Resources
kubectl apply -f kubernetes/postgres-pvc.yaml -n taskmaster
kubectl apply -f kubernetes/postgres-statefulset.yaml -n taskmaster
kubectl apply -f kubernetes/postgres-service.yaml -n taskmaster

# 4. Go Application Resources
kubectl apply -f kubernetes/app-configmap.yaml -n taskmaster
kubectl apply -f kubernetes/app-deployment.yaml -n taskmaster
kubectl apply -f kubernetes/app-service.yaml -n taskmaster

# 5. OpenTelemetry Collector Resources
kubectl apply -f kubernetes/otel-collector-config.yaml -n taskmaster
kubectl apply -f kubernetes/otel-collector-deployment.yaml -n taskmaster
kubectl apply -f kubernetes/otel-collector-service.yaml -n taskmaster

# 6. Prometheus Resources (RBAC is optional but recommended)
# kubectl apply -f kubernetes/prometheus-rbac.yaml # If using
kubectl apply -f kubernetes/prometheus-config.yaml -n taskmaster
kubectl apply -f kubernetes/prometheus-deployment.yaml -n taskmaster
kubectl apply -f kubernetes/prometheus-service.yaml -n taskmaster

# 7. Grafana Resources
kubectl apply -f kubernetes/grafana-datasources-config.yaml -n taskmaster
kubectl apply -f kubernetes/grafana-deployment.yaml -n taskmaster
kubectl apply -f kubernetes/grafana-service.yaml -n taskmaster
```

**d. Check Deployment Status:**

```bash
kubectl get pods -n taskmaster -w
kubectl get services -n taskmaster
```

Wait for all pods to be `Running`.

**e. Access Services on Minikube:**

- **Go Application API:** `minikube service taskmaster-app-svc -n taskmaster --url`
- **Prometheus:** `minikube service prometheus -n taskmaster --url` (or `http://$(minikube ip):30090`)
- **Grafana:** `minikube service grafana -n taskmaster --url` (or `http://$(minikube ip):30003`)
- Login to Grafana: `admin` / `admin`

## Testing the API

Use `curl` or a tool like Postman to interact with the API. Replace `YOUR_APP_URL` with the URL obtained from Docker Compose (`http://localhost:8080`) or Minikube.

- **Health Check:** `GET YOUR_APP_URL/healthz`
- **Create Task:** `POST YOUR_APP_URL/tasks`

```json
{
  "title": "My New Task",
  "description": "Details about the task.",
  "status": "pending"
}
```

- **Get All Tasks:** `GET YOUR_APP_URL/tasks`
- **Get Specific Task:** `GET YOUR_APP_URL/tasks/{id}`
- **Update Task:** `PUT YOUR_APP_URL/tasks/{id}`

```json
{
  "title": "Updated Task Title",
  "description": "Updated description.",
  "status": "completed"
}
```

- **Delete Task:** `DELETE YOUR_APP_URL/tasks/{id}`

## Observability

- **Traces & Metrics:** The Go application is instrumented with OpenTelemetry. It sends trace and metric data via OTLP to the OpenTelemetry Collector.
- **OpenTelemetry Collector:** Receives data from the app, processes it, and exports metrics to Prometheus. (Traces are currently logged by the collector; a Jaeger/Tempo exporter could be added).
- **Prometheus:** Scrapes metrics from the OpenTelemetry Collector. Explore metrics at the Prometheus URL (e.g., `http_server_duration_seconds_count`, `go_goroutines`).
- **Grafana:** Visualizes metrics from Prometheus. The Prometheus datasource is auto-provisioned. Create dashboards to monitor application performance (request rates, latencies, error rates, etc.).

## Cleanup

**Docker Compose:**

```bash
docker-compose down -v # -v also removes volumes (PostgreSQL data)
```

**Minikube:**

```bash
kubectl delete namespace taskmaster
# or to stop/delete Minikube cluster
# minikube stop
# minikube delete
```

```

---
```
