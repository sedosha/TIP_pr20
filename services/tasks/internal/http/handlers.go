package http

import (
    "encoding/json"
    "net/http"
    "sync"
    "time"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
    "go.uber.org/zap"

    "example.com/pz4-monitoring/internal/metrics"
    "example.com/pz4-monitoring/services/tasks/internal/service"
    "example.com/pz4-monitoring/shared/middleware"
)

type TaskHandler struct {
    mu    sync.RWMutex
    tasks map[string]service.Task
    log   *zap.Logger
}

func NewTaskHandler(log *zap.Logger) *TaskHandler {
    return &TaskHandler{
        tasks: make(map[string]service.Task),
        log:   log,
    }
}

func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    requestID := middleware.GetRequestID(r.Context())

    var req service.CreateTaskRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.log.Warn("invalid request format", zap.String("request_id", requestID), zap.Error(err))
        http.Error(w, `{"error":"invalid request format"}`, http.StatusBadRequest)
        metrics.TaskRequestsTotal.WithLabelValues("create", "error").Inc()
        return
    }

    if req.Title == "" {
        h.log.Warn("title is required", zap.String("request_id", requestID))
        http.Error(w, `{"error":"title is required"}`, http.StatusBadRequest)
        metrics.TaskRequestsTotal.WithLabelValues("create", "error").Inc()
        return
    }

    h.mu.Lock()
    defer h.mu.Unlock()

    id := uuid.New().String()[:8]
    task := service.Task{
        ID:          id,
        Title:       req.Title,
        Description: req.Description,
        DueDate:     req.DueDate,
        Done:        false,
        CreatedAt:   time.Now(),
    }

    h.tasks[id] = task

    duration := time.Since(start).Seconds()
    metrics.TaskRequestDuration.WithLabelValues("create").Observe(duration)
    metrics.TaskRequestsTotal.WithLabelValues("create", "success").Inc()

    h.log.Info("task created",
        zap.String("request_id", requestID),
        zap.String("task_id", id),
        zap.String("title", req.Title),
    )

    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Request-ID", requestID)
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) GetAllTasks(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    requestID := middleware.GetRequestID(r.Context())

    h.mu.RLock()
    defer h.mu.RUnlock()

    tasks := make([]service.Task, 0, len(h.tasks))
    for _, task := range h.tasks {
        tasks = append(tasks, task)
    }

    duration := time.Since(start).Seconds()
    metrics.TaskRequestDuration.WithLabelValues("get_all").Observe(duration)
    metrics.TaskRequestsTotal.WithLabelValues("get_all", "success").Inc()

    h.log.Debug("all tasks retrieved",
        zap.String("request_id", requestID),
        zap.Int("count", len(tasks)),
    )

    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Request-ID", requestID)
    json.NewEncoder(w).Encode(tasks)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    requestID := middleware.GetRequestID(r.Context())
    id := chi.URLParam(r, "id")

    h.mu.RLock()
    defer h.mu.RUnlock()

    task, exists := h.tasks[id]
    if !exists {
        h.log.Warn("task not found",
            zap.String("request_id", requestID),
            zap.String("task_id", id),
        )
        http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
        metrics.TaskRequestsTotal.WithLabelValues("get", "not_found").Inc()
        return
    }

    duration := time.Since(start).Seconds()
    metrics.TaskRequestDuration.WithLabelValues("get").Observe(duration)
    metrics.TaskRequestsTotal.WithLabelValues("get", "success").Inc()

    h.log.Debug("task retrieved",
        zap.String("request_id", requestID),
        zap.String("task_id", id),
    )

    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Request-ID", requestID)
    json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    requestID := middleware.GetRequestID(r.Context())
    id := chi.URLParam(r, "id")

    var req service.UpdateTaskRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.log.Warn("invalid request format", zap.String("request_id", requestID), zap.Error(err))
        http.Error(w, `{"error":"invalid request format"}`, http.StatusBadRequest)
        metrics.TaskRequestsTotal.WithLabelValues("update", "error").Inc()
        return
    }

    h.mu.Lock()
    defer h.mu.Unlock()

    task, exists := h.tasks[id]
    if !exists {
        h.log.Warn("task not found for update",
            zap.String("request_id", requestID),
            zap.String("task_id", id),
        )
        http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
        metrics.TaskRequestsTotal.WithLabelValues("update", "not_found").Inc()
        return
    }

    if req.Title != nil {
        task.Title = *req.Title
    }
    if req.Description != nil {
        task.Description = *req.Description
    }
    if req.Done != nil {
        task.Done = *req.Done
    }

    h.tasks[id] = task

    duration := time.Since(start).Seconds()
    metrics.TaskRequestDuration.WithLabelValues("update").Observe(duration)
    metrics.TaskRequestsTotal.WithLabelValues("update", "success").Inc()

    h.log.Info("task updated",
        zap.String("request_id", requestID),
        zap.String("task_id", id),
    )

    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("X-Request-ID", requestID)
    json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    requestID := middleware.GetRequestID(r.Context())
    id := chi.URLParam(r, "id")

    h.mu.Lock()
    defer h.mu.Unlock()

    if _, exists := h.tasks[id]; !exists {
        h.log.Warn("task not found for delete",
            zap.String("request_id", requestID),
            zap.String("task_id", id),
        )
        http.Error(w, `{"error":"task not found"}`, http.StatusNotFound)
        metrics.TaskRequestsTotal.WithLabelValues("delete", "not_found").Inc()
        return
    }

    delete(h.tasks, id)

    duration := time.Since(start).Seconds()
    metrics.TaskRequestDuration.WithLabelValues("delete").Observe(duration)
    metrics.TaskRequestsTotal.WithLabelValues("delete", "success").Inc()

    h.log.Info("task deleted",
        zap.String("request_id", requestID),
        zap.String("task_id", id),
    )

    w.Header().Set("X-Request-ID", requestID)
    w.WriteHeader(http.StatusNoContent)
}
