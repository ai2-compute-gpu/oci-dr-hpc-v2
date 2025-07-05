// Package logger =============================================================================
// internal/logger/logger.go - Application logging
package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
	logFile     *os.File
)

// InitLogger initializes the logger with optional file output
func InitLogger(logFilePath string) error {
	if logFilePath != "" {
		// Create log directory if it doesn't exist
		logDir := filepath.Dir(logFilePath)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			return err
		}

		// Open log file
		var err error
		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return err
		}

		// Create multi-writers for both file and stdout/stderr
		infoWriter := io.MultiWriter(os.Stdout, logFile)
		errorWriter := io.MultiWriter(os.Stderr, logFile)
		debugWriter := io.MultiWriter(os.Stdout, logFile)

		infoLogger = log.New(infoWriter, "INFO: ", log.Ldate|log.Ltime)
		errorLogger = log.New(errorWriter, "ERROR: ", log.Ldate|log.Ltime)
		debugLogger = log.New(debugWriter, "DEBUG: ", log.Ldate|log.Ltime)
	} else {
		// Default to stdout/stderr only
		infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
		errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime)
		debugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime)
	}

	return nil
}

func init() {
	// Initialize with default settings (stdout/stderr only)
	err := InitLogger("")
	if err != nil {
		log.Println("Failed to initialize logger:", err)
		return
	}
}

// CloseLogFile closes the log file if it's open
func CloseLogFile() error {
	if logFile != nil {
		return logFile.Close()
	}
	return nil
}

// Info logs an info message
func Info(v ...interface{}) {
	infoLogger.Println(v...)
}

// Error logs an error message
func Error(v ...interface{}) {
	errorLogger.Println(v...)
}

// Debug logs a debug message
func Debug(v ...interface{}) {
	debugLogger.Println(v...)
}

// Infof logs a formatted info message
func Infof(format string, v ...interface{}) {
	infoLogger.Printf(format, v...)
}

// Errorf logs a formatted error message
func Errorf(format string, v ...interface{}) {
	errorLogger.Printf(format, v...)
}

// Debugf logs a formatted debug message
func Debugf(format string, v ...interface{}) {
	debugLogger.Printf(format, v...)
}
