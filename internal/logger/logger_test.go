package logger

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestInitLogger(t *testing.T) {
	tests := []struct {
		name        string
		logFilePath string
		wantErr     bool
	}{
		{
			name:        "empty path - stdout/stderr only",
			logFilePath: "",
			wantErr:     false,
		},
		{
			name:        "valid file path",
			logFilePath: filepath.Join(t.TempDir(), "test.log"),
			wantErr:     false,
		},
		{
			name:        "nested directory path",
			logFilePath: filepath.Join(t.TempDir(), "logs", "test.log"),
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing file handle
			if logFile != nil {
				logFile.Close()
				logFile = nil
			}

			err := InitLogger(tt.logFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitLogger() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify loggers are initialized
			if infoLogger == nil || errorLogger == nil || debugLogger == nil {
				t.Error("Loggers not properly initialized")
			}

			// If file path was provided, verify file exists
			if tt.logFilePath != "" && !tt.wantErr {
				if _, err := os.Stat(tt.logFilePath); os.IsNotExist(err) {
					t.Errorf("Log file was not created: %s", tt.logFilePath)
				}
			}

			// Clean up
			if logFile != nil {
				logFile.Close()
				logFile = nil
			}
		})
	}
}

func TestCloseLogFile(t *testing.T) {
	tests := []struct {
		name        string
		setupFile   bool
		wantErr     bool
	}{
		{
			name:        "no file open",
			setupFile:   false,
			wantErr:     false,
		},
		{
			name:        "file open",
			setupFile:   true,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFile {
				tempFile := filepath.Join(t.TempDir(), "test.log")
				err := InitLogger(tempFile)
				if err != nil {
					t.Fatalf("Failed to setup test file: %v", err)
				}
			} else {
				logFile = nil
			}

			err := CloseLogFile()
			if (err != nil) != tt.wantErr {
				t.Errorf("CloseLogFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetCallerInfo(t *testing.T) {
	caller := getCallerInfo()
	
	// Should return format: filename:functionname:linenumber
	parts := strings.Split(caller, ":")
	if len(parts) != 3 {
		t.Errorf("getCallerInfo() returned unexpected format: %s", caller)
	}

	// First part should be filename or could be asm file for runtime calls
	if !strings.HasSuffix(parts[0], ".go") && !strings.HasSuffix(parts[0], ".s") {
		t.Errorf("Expected filename to end with .go or .s, got: %s", parts[0])
	}

	// Third part should be a line number (digits)
	if matched, _ := regexp.MatchString(`^\d+$`, parts[2]); !matched {
		t.Errorf("Expected line number to be digits, got: %s", parts[2])
	}
}

func TestIsNvidiaSMIFunction(t *testing.T) {
	tests := []struct {
		name     string
		funcName string
		want     bool
	}{
		{
			name:     "CheckNvidiaSMI function",
			funcName: "CheckNvidiaSMI",
			want:     true,
		},
		{
			name:     "RunNvidiaSMIQuery function",
			funcName: "RunNvidiaSMIQuery",
			want:     true,
		},
		{
			name:     "GetGPUCount function",
			funcName: "GetGPUCount",
			want:     true,
		},
		{
			name:     "RunDiagnostics function",
			funcName: "RunDiagnostics",
			want:     true,
		},
		{
			name:     "non-nvidia function",
			funcName: "SomeOtherFunction",
			want:     false,
		},
		{
			name:     "empty function name",
			funcName: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNvidiaSMIFunction(tt.funcName); got != tt.want {
				t.Errorf("isNvidiaSMIFunction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatMessage(t *testing.T) {
	msg := formatMessage("TEST", "sample message")
	
	// Message should contain level, timestamp, caller info, and message
	if !strings.Contains(msg, "TEST:") {
		t.Error("Message should contain level")
	}
	if !strings.Contains(msg, "sample message") {
		t.Error("Message should contain the actual message")
	}
	
	// Check timestamp format (YYYY/MM/DD HH:MM:SS)
	timestampRegex := regexp.MustCompile(`\d{4}/\d{2}/\d{2} \d{2}:\d{2}:\d{2}`)
	if !timestampRegex.MatchString(msg) {
		t.Error("Message should contain properly formatted timestamp")
	}
	
	// Check caller info format (filename:function:line)
	callerRegex := regexp.MustCompile(`\w+\.go:\w+:\d+`)
	if !callerRegex.MatchString(msg) {
		t.Error("Message should contain caller info")
	}
}


func TestInfo(t *testing.T) {
	// Test by logging to a file
	tempDir := t.TempDir()
	logFilePath := filepath.Join(tempDir, "test.log")

	err := InitLogger(logFilePath)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	Info("test message")

	CloseLogFile()

	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, "INFO:") {
		t.Error("Info message should contain INFO level")
	}
	if !strings.Contains(output, "test message") {
		t.Error("Info message should contain the actual message")
	}
}

func TestError(t *testing.T) {
	tempDir := t.TempDir()
	logFilePath := filepath.Join(tempDir, "test.log")

	err := InitLogger(logFilePath)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	Error("error message")

	CloseLogFile()

	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, "ERROR:") {
		t.Error("Error message should contain ERROR level")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message should contain the actual message")
	}
}

func TestDebug(t *testing.T) {
	tempDir := t.TempDir()
	logFilePath := filepath.Join(tempDir, "test.log")

	err := InitLogger(logFilePath)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	Debug("debug message")

	CloseLogFile()

	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, "DEBUG:") {
		t.Error("Debug message should contain DEBUG level")
	}
	if !strings.Contains(output, "debug message") {
		t.Error("Debug message should contain the actual message")
	}
}

func TestInfof(t *testing.T) {
	tempDir := t.TempDir()
	logFilePath := filepath.Join(tempDir, "test.log")

	err := InitLogger(logFilePath)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	Infof("formatted %s %d", "message", 42)

	CloseLogFile()

	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, "INFO:") {
		t.Error("Infof message should contain INFO level")
	}
	if !strings.Contains(output, "formatted message 42") {
		t.Error("Infof message should contain the formatted message")
	}
}

func TestErrorf(t *testing.T) {
	tempDir := t.TempDir()
	logFilePath := filepath.Join(tempDir, "test.log")

	err := InitLogger(logFilePath)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	Errorf("formatted %s %d", "error", 42)

	CloseLogFile()

	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, "ERROR:") {
		t.Error("Errorf message should contain ERROR level")
	}
	if !strings.Contains(output, "formatted error 42") {
		t.Error("Errorf message should contain the formatted message")
	}
}

func TestDebugf(t *testing.T) {
	tempDir := t.TempDir()
	logFilePath := filepath.Join(tempDir, "test.log")

	err := InitLogger(logFilePath)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	Debugf("formatted %s %d", "debug", 42)

	CloseLogFile()

	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	output := string(content)
	if !strings.Contains(output, "DEBUG:") {
		t.Error("Debugf message should contain DEBUG level")
	}
	if !strings.Contains(output, "formatted debug 42") {
		t.Error("Debugf message should contain the formatted message")
	}
}

func TestLoggerWithFile(t *testing.T) {
	tempDir := t.TempDir()
	logFilePath := filepath.Join(tempDir, "test.log")

	// Initialize logger with file
	err := InitLogger(logFilePath)
	if err != nil {
		t.Fatalf("Failed to initialize logger with file: %v", err)
	}

	// Log some messages
	Info("test info message")
	Error("test error message")
	Debug("test debug message")

	// Close the log file
	err = CloseLogFile()
	if err != nil {
		t.Fatalf("Failed to close log file: %v", err)
	}

	// Read the log file
	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logContent := string(content)
	
	// Verify all messages are in the file
	if !strings.Contains(logContent, "INFO:") {
		t.Error("Log file should contain INFO message")
	}
	if !strings.Contains(logContent, "ERROR:") {
		t.Error("Log file should contain ERROR message")
	}
	if !strings.Contains(logContent, "DEBUG:") {
		t.Error("Log file should contain DEBUG message")
	}
	if !strings.Contains(logContent, "test info message") {
		t.Error("Log file should contain info message content")
	}
	if !strings.Contains(logContent, "test error message") {
		t.Error("Log file should contain error message content")
	}
	if !strings.Contains(logContent, "test debug message") {
		t.Error("Log file should contain debug message content")
	}
}

func TestLoggerMultipleArgs(t *testing.T) {
	tempDir := t.TempDir()
	logFilePath := filepath.Join(tempDir, "test.log")

	err := InitLogger(logFilePath)
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	Info("multiple", "args", 123, true)

	CloseLogFile()

	content, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	output := string(content)
	expected := "multipleargs123 true"
	if !strings.Contains(output, expected) {
		t.Errorf("Info with multiple args should contain '%s', got: %s", expected, output)
	}
}

func TestTimestampFormat(t *testing.T) {
	msg := formatMessage("TEST", "message")
	
	// Extract timestamp from message
	parts := strings.Split(msg, " ")
	if len(parts) < 3 {
		t.Fatalf("Message format unexpected: %s", msg)
	}
	
	// Combine date and time parts
	dateTime := parts[1] + " " + parts[2]
	
	// Parse the timestamp
	_, err := time.Parse("2006/01/02 15:04:05", dateTime)
	if err != nil {
		t.Errorf("Timestamp format is invalid: %s, error: %v", dateTime, err)
	}
}