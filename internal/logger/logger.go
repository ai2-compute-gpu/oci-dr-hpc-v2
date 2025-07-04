// =============================================================================
// internal/logger/logger.go - Application logging
package logger

import (
	"log"
	"os"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
)

func init() {
	infoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime)
	errorLogger = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime)
	debugLogger = log.New(os.Stdout, "DEBUG: ", log.Ldate|log.Ltime)
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
