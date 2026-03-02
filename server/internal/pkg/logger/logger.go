// Package logger provides structured logging for the server.
package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger wraps zap.Logger for type safety.
type Logger struct {
	*zap.Logger
	sugar *zap.SugaredLogger
}

// Config holds logger configuration.
type Config struct {
	Level  string // debug, info, warn, error
	Format string // json, console
	Output string // stdout, stderr, or file path
}

// New creates a new logger instance.
func New(cfg Config) (*Logger, error) {
	// Parse log level
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		level = zapcore.InfoLevel
	}

	// Build encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create encoder
	var encoder zapcore.Encoder
	if cfg.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Create write syncer
	var writeSyncer zapcore.WriteSyncer
	switch cfg.Output {
	case "stderr":
		writeSyncer = zapcore.AddSync(os.Stderr)
	case "stdout", "":
		writeSyncer = zapcore.AddSync(os.Stdout)
	default:
		file, err := os.OpenFile(cfg.Output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		writeSyncer = zapcore.AddSync(file)
	}

	// Create core
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// Create logger
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	return &Logger{
		Logger: zapLogger,
		sugar:  zapLogger.Sugar(),
	}, nil
}

// Sync flushes any buffered log entries.
func (l *Logger) Sync() error {
	if l.Logger != nil {
		return l.Logger.Sync()
	}
	return nil
}

// Debug logs a debug message.
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

// Info logs an info message.
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Warn logs a warning message.
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

// Error logs an error message.
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

// Errorw logs an error message with arbitrary key-value pairs (sugared style).
func (l *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	l.sugar.Errorw(msg, keysAndValues...)
}

// Warnw logs a warning message with arbitrary key-value pairs (sugared style).
func (l *Logger) Warnw(msg string, keysAndValues ...interface{}) {
	l.sugar.Warnw(msg, keysAndValues...)
}

// Infow logs an info message with arbitrary key-value pairs (sugared style).
func (l *Logger) Infow(msg string, keysAndValues ...interface{}) {
	l.sugar.Infow(msg, keysAndValues...)
}

// Debugw logs a debug message with arbitrary key-value pairs (sugared style).
func (l *Logger) Debugw(msg string, keysAndValues ...interface{}) {
	l.sugar.Debugw(msg, keysAndValues...)
}

// Fatal logs a fatal message and exits.
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

// NewNop creates a no-op logger that discards all output.
func NewNop() *Logger {
	return &Logger{
		Logger: zap.NewNop(),
		sugar:  zap.NewNop().Sugar(),
	}
}

// With creates a child logger with additional fields.
func (l *Logger) With(fields ...zap.Field) *zap.Logger {
	return l.Logger.With(fields...)
}

// Sugar returns the sugared logger for convenient logging.
func (l *Logger) Sugar() *zap.SugaredLogger {
	return l.sugar
}

// Named adds a sub-logger name.
func (l *Logger) Named(name string) *Logger {
	return &Logger{
		Logger: l.Logger.Named(name),
		sugar:  l.sugar.Named(name),
	}
}

// Global logger instance (for backward compatibility)
var (
	globalLogger *Logger
)

// Init initializes the global logger.
func Init(cfg *Config) error {
	var err error
	globalLogger, err = New(*cfg)
	return err
}

// L returns the global logger.
func L() *Logger {
	return globalLogger
}

// Global functions for backward compatibility
func Debug(msg string, fields ...zap.Field) {
	globalLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	globalLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	globalLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	globalLogger.Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	globalLogger.Fatal(msg, fields...)
}
