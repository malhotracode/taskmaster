# TaskMaster Go

A production-ready, end-to-end backend system for simple task management, built with Go. This project demonstrates:

- **RESTful API** for CRUD operations on tasks
- **PostgreSQL** as the database
- **Docker Compose** for local development
- **Kubernetes** manifests for cloud-native deployment (Minikube-ready)
- **Observability** with OpenTelemetry, Prometheus, and Grafana

---

## 🚀 Features

- **CRUD API**: Create, read, update, and delete tasks
- **Database**: PostgreSQL with automatic schema creation
- **Observability**: Distributed tracing and metrics via OpenTelemetry
- **Monitoring**: Prometheus scrapes metrics, Grafana dashboards ready
- **Easy Local Dev**: One command to start the whole stack with Docker Compose
- **Cloud Native**: Kubernetes manifests for production-like deployments

---

## 🗺️ Architecture

```
+-----------------+     +-----------------+     +-----------------+
|   User/Client   | --> | Load Balancer   | --> | Go API Service  |
| (curl/Postman)  |     | (K8s Service)   |     | (Pods)          |
+-----------------+     +-----------------+     +-----------------+
         | OTLP Traces & Metrics
         v
+---------------------+
| OpenTelemetry       |
| Collector (K8s Svc) |
+---------------------+
         | Prometheus Metrics
         v
+-----------------+     +-----------------+
| Prometheus      | --> | Grafana         |
| (K8s Service)   |     | (K8s Service)   |
+-----------------+     +-----------------+
         ^
         | SQL
+----------------------+
| PostgreSQL Database  |
| (StatefulSet)        |
+----------------------+
```

---

## 📦 Project Structure

```
taskmaster-go/
├── go-app/                  # Go application source code
│   ├── main.go              # Entry point, HTTP server
│   ├── handlers.go          # HTTP request handlers
│   ├── store.go             # Database logic
│   ├── models.go            # Data structures
│   ├── telemetry.go         # OpenTelemetry setup
│   ├── go.mod, go.sum       # Go module files
│   └── Dockerfile           # Go app Dockerfile
├── kubernetes/              # Kubernetes manifests
│   ├── namespace.yaml
│   ├── postgres-*.yaml
│   ├── app-*.yaml
│   ├── otel-collector-*.yaml
│   ├── prometheus-*.yaml
│   ├── grafana-*.yaml
├── docker-compose.yml       # Local dev environment
└── README.md                # This file
```

---

## 🛠️ Prerequisites

- [Go](https://golang.org/) (1.20+)
- [Docker](https://www.docker.com/) & [Docker Compose](https://docs.docker.com/compose/)
- [Minikube](https://minikube.sigs.k8s.io/) (or any Kubernetes cluster)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

---

## ⚡ Quickstart: Local Development with Docker Compose

This will run the Go app, PostgreSQL, OpenTelemetry Collector, Prometheus, and Grafana locally.

1. **(Optional) Create a `.env` file for local Go development:**

   ```env
   # go-app/.env (example)
   PORT=8080
   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=taskuser
   DB_PASSWORD=taskpassword
   DB_NAME=taskdb
   OTEL_EXPORTER_OTLP_ENDPOINT=localhost:4317
   APP_ENV=development
   ```

2. **Ensure observability configs exist:**

   - [`kubernetes/otel-collector-config.yaml`](kubernetes/otel-collector-config.yaml)
   - [`kubernetes/prometheus-config.yaml`](kubernetes/prometheus-config.yaml)
   - [`kubernetes/grafana-config.yaml`](kubernetes/grafana-config.yaml)

3. **Start all services:**

   ```bash
   docker-compose up --build -d
   ```

4. **Access services:**

   - **API:** [http://localhost:8080](http://localhost:8080) (e.g., `/tasks`)
   - **Prometheus:** [http://localhost:9090](http://localhost:9090)
   - **Grafana:** [http://localhost:3000](http://localhost:3000) (login: `admin`/`admin`)

5. **Stop services:**

   ```bash
   docker-compose down
   ```

---

## ☸️ Deploy to Kubernetes (Minikube)

1. **Start Minikube:**

   ```bash
   minikube start --cpus=4 --memory=4096
   ```

2. **Build and load the Go app image:**

   - If using Minikube’s Docker daemon:
     ```bash
     eval $(minikube docker-env)
     cd go-app
     docker build -t taskmaster-go:latest .
     ```
     Ensure [`kubernetes/app-deployment.yaml`](kubernetes/app-deployment.yaml) uses `image: taskmaster-go:latest` and `imagePullPolicy: IfNotPresent`.

   - Or push to Docker Hub and update the image reference in the deployment manifest.

3. **Apply manifests (in order):**

   ```bash
   kubectl apply -f kubernetes/namespace.yaml
   kubectl create secret generic postgres-secret \
     --from-literal=POSTGRES_USER=taskuser \
     --from-literal=POSTGRES_PASSWORD=taskpassword \
     --from-literal=POSTGRES_DB=taskdb \
     -n taskmaster
   kubectl apply -f kubernetes/postgres-pvc.yaml -n taskmaster
   kubectl apply -f kubernetes/postgres-statefulset.yaml -n taskmaster
   kubectl apply -f kubernetes/postgres-service.yaml -n taskmaster
   kubectl apply -f kubernetes/app-configmap.yaml -n taskmaster
   kubectl apply -f kubernetes/app-deployment.yaml -n taskmaster
   kubectl apply -f kubernetes/app-service.yaml -n taskmaster
   kubectl apply -f kubernetes/otel-collector-config.yaml -n taskmaster
   kubectl apply -f kubernetes/otel-collector-deployment.yaml -n taskmaster
   kubectl apply -f kubernetes/otel-collector-service.yaml -n taskmaster
   kubectl apply -f kubernetes/prometheus-config.yaml -n taskmaster
   kubectl apply -f kubernetes/prometheus-deployment.yaml -n taskmaster
   kubectl apply -f kubernetes/prometheus-service.yaml -n taskmaster
   kubectl apply -f kubernetes/grafana-config.yaml -n taskmaster
   kubectl apply -f kubernetes/grafana-deployment.yaml -n taskmaster
   kubectl apply -f kubernetes/grafana-service.yaml -n taskmaster
   ```

4. **Check status:**

   ```bash
   kubectl get pods -n taskmaster
   kubectl get services -n taskmaster
   ```

5. **Access services:**

   - **API:** `minikube service taskmaster-app-svc -n taskmaster --url`
   - **Prometheus:** `minikube service prometheus -n taskmaster --url`
   - **Grafana:** `minikube service grafana -n taskmaster --url` (login: `admin`/`admin`)

---

## 🧪 API Usage

Use `curl` or Postman. Replace `YOUR_APP_URL` with your API endpoint.

- **Create Task**
  ```bash
  curl -X POST YOUR_APP_URL/tasks -H "Content-Type: application/json" -d '{"title":"My Task","description":"Details","status":"pending"}'
  ```

- **Get All Tasks**
  ```bash
  curl YOUR_APP_URL/tasks
  ```

- **Get Task by ID**
  ```bash
  curl YOUR_APP_URL/tasks/1
  ```

- **Update Task**
  ```bash
  curl -X PUT YOUR_APP_URL/tasks/1 -H "Content-Type: application/json" -d '{"title":"Updated","description":"Updated desc","status":"completed"}'
  ```

- **Delete Task**
  ```bash
  curl -X DELETE YOUR_APP_URL/tasks/1
  ```

---

## 📊 Observability

- **Traces & Metrics:** The Go app sends OpenTelemetry traces and metrics to the OTel Collector.
- **Prometheus:** Scrapes metrics from the OTel Collector.
- **Grafana:** Visualizes Prometheus metrics. Preconfigured datasource.

---

## 🧹 Cleanup

- **Docker Compose:**
  ```bash
  docker-compose down -v
  ```

- **Minikube:**
  ```bash
  kubectl delete namespace taskmaster
  # minikube stop
  # minikube delete
  ```

---

## 🤝 Contributing

Pull requests and issues are welcome!

---

## 📄 License

MIT
