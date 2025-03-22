package logger

import (
	"fmt"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
)

func init() {
	// Настройка конфигурации логгера
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.StacktraceKey = "stacktrace"

	// Создание логгера
	var err error
	Logger, err = config.Build()
	if err != nil {
		panic(err)
	}
}

// getCallerInfo возвращает информацию о вызывающем коде
func getCallerInfo() zap.Field {
	pc, file, line, ok := runtime.Caller(2) // 2 - пропускаем текущую функцию и функцию логгера
	if !ok {
		return zap.Skip()
	}
	fn := runtime.FuncForPC(pc)
	return zap.String("caller", fmt.Sprintf("%s:%d:%s", file, line, fn.Name()))
}

// Info логирует информационное сообщение
func Info(msg string, fields ...zap.Field) {
	fields = append(fields, getCallerInfo())
	Logger.Info(msg, fields...)
}

// Error логирует сообщение об ошибке
func Error(msg string, fields ...zap.Field) {
	fields = append(fields, getCallerInfo())
	Logger.Error(msg, fields...)
}

// Fatal логирует критическую ошибку и завершает программу
func Fatal(msg string, fields ...zap.Field) {
	fields = append(fields, getCallerInfo())
	Logger.Fatal(msg, fields...)
}

// Debug логирует отладочное сообщение
func Debug(msg string, fields ...zap.Field) {
	fields = append(fields, getCallerInfo())
	Logger.Debug(msg, fields...)
}

// With создает новый логгер с дополнительными полями
func With(fields ...zap.Field) *zap.Logger {
	return Logger.With(fields...)
}

// Sync синхронизирует буфер логгера
func Sync() error {
	return Logger.Sync()
}

// ErrorWithStack логирует ошибку со стектрейсом
func ErrorWithStack(msg string, err error) {
	fields := []zap.Field{
		zap.Error(err),
		getCallerInfo(),
	}
	Logger.Error(msg, fields...)
}

// FatalWithStack логирует критическую ошибку со стектрейсом
func FatalWithStack(msg string, err error) {
	fields := []zap.Field{
		zap.Error(err),
		getCallerInfo(),
	}
	Logger.Fatal(msg, fields...)
}
