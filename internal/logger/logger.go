package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"tree/internal/config"
)

type Logger struct {
	*log.Logger
	level Level
}

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	TraceLevel
)

var (
	globalLogger *Logger
)

func Init(cfg *config.Config) error {
	logDir := config.GetLogsDir()
	logFile := filepath.Join(logDir, "app.log")

	// Создаем директорию для логов
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Открываем файл лога
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Определяем уровень логирования
	level := parseLevel(cfg.LogLevel)

	// Настраиваем вывод: файл + консоль в debug-режиме
	var writers []io.Writer
	writers = append(writers, file)

	if level == DebugLevel {
		writers = append(writers, os.Stdout)
	}

	multiWriter := io.MultiWriter(writers...)
	globalLogger = &Logger{
		Logger: log.New(multiWriter, "", log.Ldate|log.Ltime|log.Lshortfile),
		level:  level,
	}

	Info("Logger initialized. Level: %s, Log file: %s",
		level.String(), logFile)
	return nil
}

func parseLevel(levelStr string) Level {
	switch strings.ToLower(levelStr) {
	case "debug":
		return DebugLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Debug logs a message at Debug level
func Debug(format string, v ...interface{}) {
	if globalLogger != nil && globalLogger.level <= DebugLevel {
		globalLogger.log(DebugLevel, format, v...)
	}
}

// Debugf logs a formatted message at Debug level
func Debugf(format string, v ...interface{}) {
	Debug(format, v...)
}

// Info logs a message at Info level
func Info(format string, v ...interface{}) {
	if globalLogger != nil && globalLogger.level <= InfoLevel {
		globalLogger.log(InfoLevel, format, v...)
	}
}

// Infof logs a formatted message at Info level
func Infof(format string, v ...interface{}) {
	Info(format, v...)
}

// Warn logs a message at Warn level
func Warn(format string, v ...interface{}) {
	if globalLogger != nil && globalLogger.level <= WarnLevel {
		globalLogger.log(WarnLevel, format, v...)
	}
}

// Warnf logs a formatted message at Warn level
func Warnf(format string, v ...interface{}) {
	Warn(format, v...)
}

// Error logs a message at Error level
func Error(format string, v ...interface{}) {
	if globalLogger != nil && globalLogger.level <= ErrorLevel {
		globalLogger.log(ErrorLevel, format, v...)
	}
}

// Errorf logs a formatted message at Error level
func Errorf(format string, v ...interface{}) {
	Error(format, v...)
}
func Trace(format string, v ...interface{}) {
	if globalLogger != nil && globalLogger.level <= TraceLevel {
		globalLogger.log(TraceLevel, format, v...)
	}
}
func Tracef(format string, v ...interface{}) {
	Trace(format, v...)
}
func (l *Logger) log(level Level, format string, v ...interface{}) {
	// Добавляем префикс уровня и временной метки
	prefix := fmt.Sprintf("[%s] [%s] ", time.Now().Format("15:04:05"), level.String())
	message := fmt.Sprintf(format, v...)
	l.Logger.Output(3, prefix+message)
}
