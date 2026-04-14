package http

import (
    "context"
    "fmt"
    "net/http"
    
    "go.uber.org/zap"
    
    "example.com/pz4-monitoring/services/tasks/internal/grpcclient"
)

func AuthGRPCMiddleware(client *grpcclient.AuthGRPCClient, log *zap.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                log.Warn("missing authorization header",
                    zap.String("path", r.URL.Path),
                )
                http.Error(w, `{"error":"missing authorization header"}`, http.StatusUnauthorized)
                return
            }
            
            var token string
            _, err := fmt.Sscanf(authHeader, "Bearer %s", &token)
            if err != nil || token == "" {
                log.Warn("invalid authorization format",
                    zap.String("path", r.URL.Path),
                    zap.String("auth_header", authHeader),
                )
                http.Error(w, `{"error":"invalid authorization format"}`, http.StatusUnauthorized)
                return
            }
            
            valid, subject, err := client.VerifyToken(r.Context(), token)
            if err != nil {
                log.Error("auth service unavailable",
                    zap.String("path", r.URL.Path),
                    zap.Error(err),
                )
                http.Error(w, `{"error":"auth service unavailable"}`, http.StatusServiceUnavailable)
                return
            }
            
            if !valid {
                log.Warn("invalid token",
                    zap.String("path", r.URL.Path),
                )
                http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
                return
            }
            
            log.Debug("token validated",
                zap.String("path", r.URL.Path),
                zap.String("subject", subject),
            )
            
            ctx := context.WithValue(r.Context(), "username", subject)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
