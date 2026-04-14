package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    HttpRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "app_http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "path"},
    )

    HttpErrorsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "app_http_errors_total",
            Help: "Total number of HTTP error responses",
        },
        []string{"method", "path", "status_code"},
    )

    HttpRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "app_http_request_duration_seconds",
            Help:    "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )

    TaskRequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "app_task_requests_total",
            Help: "Total number of task operations",
        },
        []string{"operation", "status"},
    )

    TaskRequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "app_task_request_duration_seconds",
            Help:    "Task operation duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"operation"},
    )
)
