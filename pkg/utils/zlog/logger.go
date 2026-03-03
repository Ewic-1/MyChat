package zlog

import (
	"path/filepath"
	"runtime"

	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	l, err := zap.NewDevelopment()
	if err != nil {
		logger = zap.NewNop()
		return
	}
	logger = l
}

func getCallerInfoForLog() []zap.Field {
	pc, file, line, ok := runtime.Caller(2)
	if !ok {
		return nil
	}
	funcName := runtime.FuncForPC(pc).Name()
	funcName = filepath.Base(funcName)
	return []zap.Field{
		zap.String("func", funcName),
		zap.String("file", file),
		zap.Int("line", line),
	}
}

func Info(message string, fields ...zap.Field) {
	fields = append(fields, getCallerInfoForLog()...)
	logger.Info(message, fields...)
}

func Warn(message string, fields ...zap.Field) {
	fields = append(fields, getCallerInfoForLog()...)
	logger.Warn(message, fields...)
}

func Error(message string, fields ...zap.Field) {
	fields = append(fields, getCallerInfoForLog()...)
	logger.Error(message, fields...)
}

func Fatal(message string, fields ...zap.Field) {
	fields = append(fields, getCallerInfoForLog()...)
	logger.Fatal(message, fields...)
}

func Debug(message string, fields ...zap.Field) {
	fields = append(fields, getCallerInfoForLog()...)
	logger.Debug(message, fields...)
}
