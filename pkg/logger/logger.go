package logger

import (
    "go.uber.org/zap"
    "go.uber.org/zap/zapcore"
)

func New() (*zap.Logger, error) {
    cfg := zap.NewProductionConfig()
    
    // Меняем уровень логирования на Debug
    cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
    
    cfg.OutputPaths = []string{"stdout"}
    cfg.ErrorOutputPaths = []string{"stderr"}
    
    return cfg.Build()
}
