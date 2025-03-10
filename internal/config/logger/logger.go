package logger

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// logWrapper ensures logs are properly formatted even if the logger
// isn't initialized.
func logWrapper(level string, format string, args ...interface{}) {
	// Format the message once to keep things consistent.
	message := fmt.Sprintf(format, args...)

	// If the queue is not closed, add the message to the queue.
	if !queueClosed.Load() {
		select {
		// Add the message to the queue.
		case queue <- logOperation{level, message}:
		// Drop the message if the queue is full.
		default:
		}
		return
	}

	sublogWrapper(level, message)
}

func sublogWrapper(level string, message string) {
	// Choose the correct log level
	switch level {
	case "debug":
		log.Debug().Msg(message)
	case "info":
		log.Info().Msg(message)
	case "warn":
		log.Warn().Msg(message)
	case "error":
		log.Error().Msg(message)
	case "fatal":
		log.Fatal().Msg(message)
	case "panic":
		log.Panic().Msg(message)
	case "trace":
		log.Trace().Msg(message)
	default:
		// Default to Info level
		log.Info().Msg(message)
	}
}

// LogTrace logs a trace message.
func LogTrace(format string, args ...interface{}) {
	logWrapper("trace", format, args...)
}

// LogDebug logs a debug message.
func LogDebug(format string, args ...interface{}) {
	logWrapper("debug", format, args...)
}

// LogInfo logs an info message.
func LogInfo(format string, args ...interface{}) {
	logWrapper("info", format, args...)
}

// LogWarn logs a warning message.
func LogWarn(format string, args ...interface{}) {
	logWrapper("warn", format, args...)
}

// LogError logs an error message.
func LogError(format string, args ...interface{}) {
	logWrapper("error", format, args...)
}

// LogFatal logs a fatal message.
func LogFatal(format string, args ...interface{}) {
	logWrapper("fatal", format, args...)
}

// LogPanic logs a panic message.
func LogPanic(format string, args ...interface{}) {
	logWrapper("panic", format, args...)
}

// MarkInitialized marks the logger as initialized.
func MarkInitialized() {
	if !queueClosed.Load() {
		queueClosed.Store(true)
		close(queue)
		go processQueue()
	}
	LogInfo("Logger initialized")
}
