package logutil

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/tw93/mole/cmd/platform"
)

// Logger provides leveled logging with file rotation support.
type Logger struct {
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	opsLogger   *log.Logger
	logFile     *os.File
}

type Config struct {
	Level     string
	LogDir    string
	DebugMode bool
	OpLog     bool
}

// New creates a new Logger, writing to the platform's log directory.
func New(cfg Config) (*Logger, error) {
	p := platform.Current
	if p == nil {
		// Fallback for tests if Current is not initialized
		return &Logger{
			debugLogger: log.New(os.Stdout, "DEBUG: ", log.Ltime),
			infoLogger:  log.New(os.Stdout, "INFO: ", log.Ltime),
			warnLogger:  log.New(os.Stderr, "WARN: ", log.Ltime),
			opsLogger:   log.New(os.Stdout, "OPS: ", log.Ltime),
		}, nil
	}

	// Use platform log directory
	logDir := p.MoleLogDir()
	
	// Create logs dir if it doesn't exist
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath := filepath.Join(logDir, "mole.log")
	
	// Open log file, append mode
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Multiwriter to write to both stdout/stderr and file
	// For CLI we often don't want debug in stdout, but for Mole we'll just write to file for now
	// to avoid cluttering BubbleTea UI.
	
	return &Logger{
		debugLogger: log.New(f, "DEBUG: ", log.Ldate|log.Ltime),
		infoLogger:  log.New(f, "INFO: ", log.Ldate|log.Ltime),
		warnLogger:  log.New(io.MultiWriter(os.Stderr, f), "WARN: ", log.Ldate|log.Ltime),
		opsLogger:   log.New(f, "OPS: ", log.Ldate|log.Ltime),
		logFile:     f,
	}, nil
}

func NewConsoleLogger() *Logger {
	return &Logger{
		debugLogger: log.New(os.Stdout, "DEBUG: ", log.Ltime),
		infoLogger:  log.New(os.Stdout, "INFO: ", log.Ltime),
		warnLogger:  log.New(os.Stderr, "WARN: ", log.Ltime),
		opsLogger:   log.New(os.Stdout, "OPS: ", log.Ltime),
	}
}

func NewStreamLogger(out io.Writer) *Logger {
	return &Logger{
		debugLogger: log.New(out, "DEBUG: ", log.Ltime),
		infoLogger:  log.New(out, "INFO: ", log.Ltime),
		warnLogger:  log.New(out, "WARN: ", log.Ltime),
		opsLogger:   log.New(out, "OPS: ", log.Ltime),
	}
}


// Close closes the underlying log file.
func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

func (l *Logger) Debug(format string, v ...interface{}) {
	l.debugLogger.Printf(format, v...)
}

func (l *Logger) Info(format string, v ...interface{}) {
	l.infoLogger.Printf(format, v...)
}

func (l *Logger) Warning(format string, v ...interface{}) {
	l.warnLogger.Printf(format, v...)
}

func (l *Logger) Error(format string, v ...interface{}) {
	l.warnLogger.Printf("ERROR: "+format, v...)
}

func (l *Logger) Success(format string, v ...interface{}) {
	l.infoLogger.Printf("SUCCESS: "+format, v...)
}

// Ops records a destructive operation (like a deletion).
func (l *Logger) Ops(operation, path string, size int64) {
	l.opsLogger.Printf("[%s] OP=%s PATH=%s SIZE=%d", time.Now().Format(time.RFC3339), operation, path, size)
}
