package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var tracer = otel.Tracer("taskmaster-go/handlers")

func jsonResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func errorResponse(w http.ResponseWriter, statusCode int, message string) {
	jsonResponse(w, statusCode, map[string]string{"error": message})
}

func getTasksHandler(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "getTasksHandler")
	defer span.End()

	tasks, err := GetTasks()
	if err != nil {
		span.RecordError(err)
		errorResponse(w, http.StatusInternalServerError, "Failed to fetch tasks")
		return
	}
	jsonResponse(w, http.StatusOK, tasks)
}

func getTaskHandler(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "getTaskHandler")
	defer span.End()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid task ID")
		return
	}
	span.SetAttributes(attribute.Int("task.id", id))

	task, err := GetTask(id)
	if err != nil {
		span.RecordError(err)
		errorResponse(w, http.StatusInternalServerError, "Failed to fetch task")
		return
	}
	if task == nil {
		errorResponse(w, http.StatusNotFound, "Task not found")
		return
	}
	jsonResponse(w, http.StatusOK, task)
}

func createTaskHandler(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "createTaskHandler")
	defer span.End()

	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if task.Title == "" {
		errorResponse(w, http.StatusBadRequest, "Title is required")
		return
	}
	if task.Status == "" {
		task.Status = "pending" // Default status
	}

	if err := CreateTask(&task); err != nil {
		span.RecordError(err)
		errorResponse(w, http.StatusInternalServerError, "Failed to create task")
		return
	}
	span.SetAttributes(attribute.Int("task.id", task.ID))
	jsonResponse(w, http.StatusCreated, task)
}

func updateTaskHandler(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "updateTaskHandler")
	defer span.End()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid task ID")
		return
	}
	span.SetAttributes(attribute.Int("task.id", id))

	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if task.Title == "" {
		errorResponse(w, http.StatusBadRequest, "Title is required")
		return
	}

	if err := UpdateTask(id, &task); err != nil {
		span.RecordError(err)
		errorResponse(w, http.StatusInternalServerError, "Failed to update task")
		return
	}
	task.ID = id // Ensure ID is set in response
	jsonResponse(w, http.StatusOK, task)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	_, span := tracer.Start(r.Context(), "deleteTaskHandler")
	defer span.End()

	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid task ID")
		return
	}
	span.SetAttributes(attribute.Int("task.id", id))

	if err := DeleteTask(id); err != nil {
		span.RecordError(err)
		errorResponse(w, http.StatusInternalServerError, "Failed to delete task")
		return
	}
	jsonResponse(w, http.StatusNoContent, nil)
}

// Make sure your routes are wrapped with otelhttp.NewHandler
func NewRouter() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/tasks", getTasksHandler).Methods("GET")
	r.HandleFunc("/tasks", createTaskHandler).Methods("POST")
	r.HandleFunc("/tasks/{id}", getTaskHandler).Methods("GET")
	r.HandleFunc("/tasks/{id}", updateTaskHandler).Methods("PUT")
	r.HandleFunc("/tasks/{id}", deleteTaskHandler).Methods("DELETE")

	// Wrap the router with the OTel HTTP middleware.
	// This will record metrics for all HTTP requests.
	// If you also want traces, you can pass "otelhttp.WithTracerProvider(otel.GetTracerProvider())"
	// but since we set a global tracer provider, it should pick it up.
	return otelhttp.NewHandler(r, "http-server")
}