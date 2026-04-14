package middleware

import (
    "net/http"
    "time"

    "go.uber.org/zap"
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

func ZapLoggingMiddleware(log *zap.Logger, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        lrw := NewLoggingResponseWriter(w)
        requestID := GetRequestID(r.Context())
        if requestID == "" {
            requestID = "no-request-id"
        }

        log.Info("incoming request",
            zap.String("request_id", requestID),
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
            zap.String("remote_addr", r.RemoteAddr),
        )

        next.ServeHTTP(lrw, r)

        duration := time.Since(start)

        log.Info("request completed",
            zap.String("request_id", requestID),
            zap.String("method", r.Method),
            zap.String("path", r.URL.Path),
            zap.Int("status_code", lrw.StatusCode()),
            zap.Duration("duration", duration),
        )
    })
}
