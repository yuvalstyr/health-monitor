package logger

import (
	"io"
	"os"
	stdlogger "log" // Standard log package for middleware compatibility
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Setup configures zerolog with appropriate settings
func Setup() {
	// Set up pretty console logging for development
	if os.Getenv("ENV") != "production" {
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			NoColor:    false,
		})
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		// In production, use JSON format
		zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Override global level if DEBUG env var is set
	if os.Getenv("DEBUG") == "true" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}
}

// Create a bridge from zerolog to standard logger for middleware
type zerologBridge struct{}

func (z zerologBridge) Write(p []byte) (n int, err error) {
	log.Info().Msg(string(p))
	return len(p), nil
}

// StdLogger returns a standard logger that uses zerolog as backend
func StdLogger() *stdlogger.Logger {
	return stdlogger.New(zerologBridge{}, "", 0)
}

// The following functions are convenient wrappers around zerolog

// Debug logs a debug message
func Debug() *zerolog.Event {
	return log.Debug()
}

// Info logs an info message
func Info() *zerolog.Event {
	return log.Info()
}

// Warn logs a warning message
func Warn() *zerolog.Event {
	return log.Warn()
}

// Error logs an error message
func Error() *zerolog.Event {
	return log.Error()
}

// Fatal logs a fatal message and then calls os.Exit(1)
func Fatal() *zerolog.Event {
	return log.Fatal()
}

// Writer returns a writer that logs at the given level
func Writer(level zerolog.Level) io.Writer {
	return zerolog.NewConsoleWriter(func(w *zerolog.ConsoleWriter) {
		w.Out = os.Stdout
		w.NoColor = false
	})
}
