package logger

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
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
func logWithCaller(level string, color string, callerDepth int, format string, args ...interface{}) {
	// Get the caller information
	_, file, line, ok := runtime.Caller(callerDepth)
	if !ok {
		file = "???"
		line = 0
	}

	// ファイル名を短縮（長すぎると見づらいため）
	shortFile := shortenFilePath(file)

	message := fmt.Sprintf(format, args...)
	logMessage := fmt.Sprintf("%s%s%s: %s:%d: %s", color, level, ColorReset, shortFile, line, message)
	Logger.Println(logMessage)
}

// ファイルパスを短縮する
func shortenFilePath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) > 3 {
		return strings.Join(parts[len(parts)-3:], "/")
	}
	return path
}

// 一般ログ
func Info(message string) {
	logWithCaller("INFO", ColorGreen, 3, "%s", message)
}

func Infof(format string, args ...interface{}) {
	logWithCaller("INFO", ColorGreen, 3, format, args...)
}

// エラーログ
func Error(err error) {
	if err != nil {
		logWithCaller("ERROR", ColorRed, 3, "%s", err.Error())
	}
}

func Errorf(format string, args ...interface{}) {
	logWithCaller("ERROR", ColorRed, 3, format, args...)
}

// デバッグログ
func Debug(message string) {
	logWithCaller("DEBUG", ColorBlue, 3, "%s", message)
}

func Debugf(format string, args ...interface{}) {
	logWithCaller("DEBUG", ColorBlue, 3, format, args...)
}

func Warn(message string) {
	logWithCaller("WARN", ColorYellow, 3, "%s", message)
}

func Warnf(format string, args ...interface{}) {
	logWithCaller("WARN", ColorYellow, 3, format, args...)
}

// **追加: エラーのスタックトレースを出力する**
func ErrorWithStack(err error) {
	if err != nil {
		stackTrace := getStackTrace()
		logWithCaller("ERROR", ColorRed, 3, "%s\nStack Trace:\n%s", err.Error(), stackTrace)
	}
}

// **スタックトレースを取得する関数**
func getStackTrace() string {
	stack := make([]byte, 1024)
	length := runtime.Stack(stack, false)
	return string(stack[:length])
}
