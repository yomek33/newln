package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

var Logger *log.Logger

func init() {
	Logger = log.New(os.Stdout, "SLOG: ", log.Ldate|log.Ltime)
}

// logWithCaller logs a message with the file name and line number of the caller
func logWithCaller(level string, color string, format string, args ...interface{}) {
	// Get the caller information
	_, file, line, ok := runtime.Caller(2)
	if !ok {
		file = "???"
		line = 0
	}
	message := fmt.Sprintf(format, args...)
	logMessage := fmt.Sprintf("%s%s%s: %s:%d: %s", color, level, ColorReset, file, line, message)
	Logger.Println(logMessage)
}

func Info(message string) {
	logWithCaller("INFO", ColorGreen, "%s", message)
}

func Infof(format string, args ...interface{}) {
	logWithCaller("INFO", ColorGreen, format, args...)
}

func Error(err error) {
	if err != nil {
		logWithCaller("ERROR", ColorRed, "%s", err.Error())
	}
}

func Errorf(format string, args ...interface{}) {
	logWithCaller("ERROR", ColorRed, format, args...)
}

func Debug(message string) {
	logWithCaller("DEBUG", ColorBlue, "%s", message)
}

func Debugf(format string, args ...interface{}) {
	logWithCaller("DEBUG", ColorBlue, format, args...)
}
