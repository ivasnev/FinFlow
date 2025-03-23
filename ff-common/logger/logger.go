package logger

import (
	"os"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// FFLogger — интерфейс логгера для всех сервисов
type FFLogger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	With(fields ...zap.Field) FFLogger
}

// ZapLogger — реализация логгера на базе zap
type ZapLogger struct {
	logger *zap.Logger
}

// NewLogger создаёт новый zap-логгер
func NewLogger(logFile string, logLevel string) FFLogger {
	level := getLogLevel(logLevel)

	writeSyncer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    10, // MB
		MaxBackups: 5,  // Количество файлов
		MaxAge:     30, // Дней хранения
		Compress:   true,
	})

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:      "timestamp",
		LevelKey:     "level",
		MessageKey:   "message",
		CallerKey:    "caller",
		EncodeTime:   customTimeEncoder,
		EncodeLevel:  zapcore.CapitalColorLevelEncoder,
		EncodeCaller: zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(writeSyncer, zapcore.AddSync(os.Stdout)),
		level,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return &ZapLogger{logger: logger}
}

// Реализация методов интерфейса
func (l *ZapLogger) Debug(msg string, fields ...zap.Field) { l.logger.Debug(msg, fields...) }
func (l *ZapLogger) Info(msg string, fields ...zap.Field)  { l.logger.Info(msg, fields...) }
func (l *ZapLogger) Warn(msg string, fields ...zap.Field)  { l.logger.Warn(msg, fields...) }
func (l *ZapLogger) Error(msg string, fields ...zap.Field) { l.logger.Error(msg, fields...) }
func (l *ZapLogger) Fatal(msg string, fields ...zap.Field) { l.logger.Fatal(msg, fields...) }

// With добавляет дополнительные поля в лог
func (l *ZapLogger) With(fields ...zap.Field) FFLogger {
	newLogger := l.logger.With(fields...)
	return &ZapLogger{logger: newLogger}
}

// getLogLevel преобразует строку в zapcore.Level
func getLogLevel(logLevel string) zapcore.Level {
	switch logLevel {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// customTimeEncoder задаёт формат времени
func customTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}
