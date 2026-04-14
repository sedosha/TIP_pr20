package main

import (
    "log"
    "net"
    "os"
    
    "google.golang.org/grpc"
    "go.uber.org/zap"
    
    pb "example.com/pz4-monitoring/proto"
    authgrpc "example.com/pz4-monitoring/services/auth/internal/grpc"
    applogger "example.com/pz4-monitoring/pkg/logger"
)

func main() {
    zapLogger, err := applogger.New()
    if err != nil {
        log.Fatal(err)
    }
    defer zapLogger.Sync()
    
    grpcPort := os.Getenv("AUTH_GRPC_PORT")
    if grpcPort == "" {
        grpcPort = "50051"
    }
    
    lis, err := net.Listen("tcp", ":"+grpcPort)
    if err != nil {
        zapLogger.Fatal("failed to listen", zap.Error(err))
    }
    
    s := grpc.NewServer()
    pb.RegisterAuthServiceServer(s, &authgrpc.AuthServer{})
    
    zapLogger.Info("Auth gRPC server starting",
        zap.String("port", grpcPort),
    )
    
    if err := s.Serve(lis); err != nil {
        zapLogger.Fatal("failed to serve", zap.Error(err))
    }
}
