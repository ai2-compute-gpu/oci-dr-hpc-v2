// Package logger =============================================================================
// internal/logger/logger.go - Application logging
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
	logFile     *os.File
	logLevel    string = "info" // Default log level
)

// InitLogger initializes the logger with optional file output
func InitLogger(logFilePath string) error {
	return InitLoggerWithLevel(logFilePath, logLevel)
}

// InitLoggerWithLevel initializes the logger with specified log level
func InitLoggerWithLevel(logFilePath string, level string) error {
	logLevel = level
	if logFilePath != "" {
		// Check if the log file path exists as a directory and remove it
		if info, err := os.Stat(logFilePath); err == nil && info.IsDir() {
			if err := os.RemoveAll(logFilePath); err != nil {
				return fmt.Errorf("failed to remove existing directory at log path %s: %v", logFilePath, err)
			}
		}

		// Create parent directory if it doesn't exist
		if dir := filepath.Dir(logFilePath); dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create log directory %s: %v", dir, err)
			}
		}

		// Try to open log file
		var err error
		logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("failed to open log file %s: %v (ensure /var/log/oci-dr-hpc/ exists with write permissions)", logFilePath, err)
		}

		// Create multi-writers for both file and stdout/stderr
		infoWriter := io.MultiWriter(os.Stdout, logFile)
		errorWriter := io.MultiWriter(os.Stderr, logFile)
		debugWriter := io.MultiWriter(os.Stdout, logFile)

		infoLogger = log.New(infoWriter, "", 0)
		errorLogger = log.New(errorWriter, "", 0)
		debugLogger = log.New(debugWriter, "", 0)
	} else {
		// Default to stdout/stderr only
		infoLogger = log.New(os.Stdout, "", 0)
		errorLogger = log.New(os.Stderr, "", 0)
		debugLogger = log.New(os.Stdout, "", 0)
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

// getCallerInfo returns the filename, function name, and line number of the caller
func getCallerInfo() string {
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown:unknown:0"
	}

	funcName := "unknown"
	if fn := runtime.FuncForPC(pc); fn != nil {
		fullName := fn.Name()
		// Extract just the function name (remove package path)
		if lastDot := strings.LastIndex(fullName, "."); lastDot != -1 {
			funcName = fullName[lastDot+1:]
		} else {
			funcName = fullName
		}
	}

	fileName := filepath.Base(file)

	// Map nvidia-smi related functions to nvidia_smi.go
	if isNvidiaSMIFunction(funcName) {
		fileName = "nvidia_smi.go"
	}

	return fmt.Sprintf("%s:%s:%d", fileName, funcName, line)
}

// isNvidiaSMIFunction checks if the function name is nvidia-smi related
func isNvidiaSMIFunction(funcName string) bool {
	nvidiaSMIFunctions := []string{
		"CheckNvidiaSMI",
		"RunNvidiaSMIQuery",
		"GetGPUCount",
		"RunDiagnostics",
	}

	for _, nvidiaFunc := range nvidiaSMIFunctions {
		if funcName == nvidiaFunc {
			return true
		}
	}
	return false
}

// formatMessage creates a formatted log message with timestamp, level, caller info
func formatMessage(level string, msg string) string {
	now := time.Now().UTC()
	caller := getCallerInfo()
	return fmt.Sprintf("%s: %s %s: %s", level, now.Format("2006/01/02 15:04:05"), caller, msg)
}

// Info logs an info message
func Info(v ...interface{}) {
	if shouldLog("info") {
		msg := fmt.Sprint(v...)
		infoLogger.Println(formatMessage("INFO", msg))
	}
}

// Error logs an error message
func Error(v ...interface{}) {
	if shouldLog("error") {
		msg := fmt.Sprint(v...)
		errorLogger.Println(formatMessage("ERROR", msg))
	}
}

// Debug logs a debug message
func Debug(v ...interface{}) {
	if shouldLog("debug") {
		msg := fmt.Sprint(v...)
		debugLogger.Println(formatMessage("DEBUG", msg))
	}
}

// Infof logs a formatted info message
func Infof(format string, v ...interface{}) {
	if shouldLog("info") {
		msg := fmt.Sprintf(format, v...)
		infoLogger.Println(formatMessage("INFO", msg))
	}
}

// Errorf logs a formatted error message
func Errorf(format string, v ...interface{}) {
	if shouldLog("error") {
		msg := fmt.Sprintf(format, v...)
		errorLogger.Println(formatMessage("ERROR", msg))
	}
}

// Debugf logs a formatted debug message
func Debugf(format string, v ...interface{}) {
	if shouldLog("debug") {
		msg := fmt.Sprintf(format, v...)
		debugLogger.Println(formatMessage("DEBUG", msg))
	}
}

// shouldLog determines if a message should be logged based on the current log level
func shouldLog(level string) bool {
	currentLevel := strings.ToLower(logLevel)
	targetLevel := strings.ToLower(level)
	
	switch currentLevel {
	case "debug":
		return true // Log everything
	case "info":
		return targetLevel == "info" || targetLevel == "error"
	case "error":
		return targetLevel == "error"
	default:
		return true // Default to logging everything
	}
}

// SetLogLevel sets the current log level
func SetLogLevel(level string) {
	logLevel = level
}
