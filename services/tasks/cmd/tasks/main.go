package main

import (
    "fmt"
    "net/http"
    "os"

    "github.com/prometheus/client_golang/prometheus/promhttp"
    "go.uber.org/zap"

    "example.com/pz4-monitoring/internal/httpapi"
    "example.com/pz4-monitoring/services/tasks/internal/grpcclient"
    httphandler "example.com/pz4-monitoring/services/tasks/internal/http"
    "example.com/pz4-monitoring/shared/middleware"
    applogger "example.com/pz4-monitoring/pkg/logger"
)

func main() {
    zapLogger, err := applogger.New()
    if err != nil {
        panic(err)
    }
    defer zapLogger.Sync()

    port := os.Getenv("TASKS_PORT")
    if port == "" {
        port = "8090"
    }

    grpcAddr := os.Getenv("AUTH_GRPC_ADDR")
    if grpcAddr == "" {
        grpcAddr = "localhost:50051"
        zapLogger.Info("using default gRPC address", zap.String("addr", grpcAddr))
    }

    authClient, err := grpcclient.NewAuthGRPCClient(grpcAddr)
    if err != nil {
        zapLogger.Fatal("failed to create auth gRPC client", zap.Error(err))
    }
    defer authClient.Close()

    handler := httphandler.NewTaskHandler(zapLogger)
    router := httphandler.NewRouter(handler, authClient, zapLogger)

    // Добавляем middleware для метрик
    metricsRouter := httpapi.MetricsMiddleware(router)

    // Создаем основной mux с /metrics эндпоинтом
    mux := http.NewServeMux()
    mux.Handle("/metrics", promhttp.Handler())
    mux.Handle("/", metricsRouter)

    rootHandler := middleware.RequestIDMiddleware(mux)

    addr := fmt.Sprintf(":%s", port)
    zapLogger.Info("Tasks service starting",
        zap.String("port", port),
        zap.String("grpc_addr", grpcAddr),
    )
    zapLogger.Info("Metrics endpoint available at", zap.String("url", fmt.Sprintf("http://localhost:%s/metrics", port)))

    if err := http.ListenAndServe(addr, rootHandler); err != nil {
        zapLogger.Fatal("failed to start tasks service", zap.Error(err))
    }
}
