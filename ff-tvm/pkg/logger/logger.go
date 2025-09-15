package logger

import (
	"fmt"
	"runtime"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func New() (*Logger, error) {
	// Настройка конфигурации логгера
	config := zap.NewProductionConfig()
	config.OutputPaths = []string{"stdout"}
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.StacktraceKey = "stacktrace"

	// Создание логгера
	zapLogger, err := config.Build()
	if err != nil {
		return nil, fmt.Errorf("failed to create logger: %w", err)
	}

	return &Logger{zapLogger}, nil
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
func (l *Logger) Info(msg string, fields ...zap.Field) {
	fields = append(fields, getCallerInfo())
	l.Logger.Info(msg, fields...)
}

// Error логирует сообщение об ошибке
func (l *Logger) Error(msg string, fields ...zap.Field) {
	fields = append(fields, getCallerInfo())
	l.Logger.Error(msg, fields...)
}

// Fatal логирует критическую ошибку и завершает программу
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	fields = append(fields, getCallerInfo())
	l.Logger.Fatal(msg, fields...)
}

// Debug логирует отладочное сообщение
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	fields = append(fields, getCallerInfo())
	l.Logger.Debug(msg, fields...)
}

// With создает новый логгер с дополнительными полями
func (l *Logger) With(fields ...zap.Field) *Logger {
	return &Logger{l.Logger.With(fields...)}
}

// Sync синхронизирует буфер логгера
func (l *Logger) Sync() error {
	return l.Logger.Sync()
}

// ErrorWithStack логирует ошибку со стектрейсом
func (l *Logger) ErrorWithStack(msg string, err error) {
	fields := []zap.Field{
		zap.Error(err),
		getCallerInfo(),
	}
	l.Logger.Error(msg, fields...)
}

// FatalWithStack логирует критическую ошибку со стектрейсом
func (l *Logger) FatalWithStack(msg string, err error) {
	fields := []zap.Field{
		zap.Error(err),
		getCallerInfo(),
	}
	l.Logger.Fatal(msg, fields...)
}
