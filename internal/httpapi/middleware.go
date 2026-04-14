package httpapi

import (
    "net/http"
    "strconv"
    "time"

    "example.com/pz4-monitoring/internal/metrics"
)

type LoggingResponseWriter struct {
    http.ResponseWriter
    statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
    return &LoggingResponseWriter{
        ResponseWriter: w,
        statusCode:     http.StatusOK,
    }
}

func (lrw *LoggingResponseWriter) WriteHeader(code int) {
    lrw.statusCode = code
    lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *LoggingResponseWriter) StatusCode() int {
    return lrw.statusCode
}

func MetricsMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        lrw := NewLoggingResponseWriter(w)

        next.ServeHTTP(lrw, r)

        duration := time.Since(start).Seconds()
        path := r.URL.Path

        metrics.HttpRequestsTotal.WithLabelValues(r.Method, path).Inc()
        metrics.HttpRequestDuration.WithLabelValues(r.Method, path).Observe(duration)

        if lrw.StatusCode() >= 400 {
            metrics.HttpErrorsTotal.WithLabelValues(
                r.Method,
                path,
                strconv.Itoa(lrw.StatusCode()),
            ).Inc()
        }
    })
}
