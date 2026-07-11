package logn

import (
	"fmt"
	"os"
	"path/filepath"
	"os/user"
)

const logFileName = "youtube_uploader_mcp.log"

var logFile *os.File

func init() {
	// Best-effort logging only. If we can't create a file (e.g. sandboxed
	// environment), we just disable file logging rather than crashing.
	logFile = openBestEffortLogFile()
}

func openBestEffortLogFile() *os.File {
	// Prefer an explicit directory if provided.
	if dir := os.Getenv("YOUTUBE_UPLOADER_MCP_LOG_DIR"); dir != "" {
		if f := tryOpenLogFile(dir); f != nil {
			return f
		}
	}

	// Next, try the current working directory (usually writable in CI/sandboxes).
	if wd, err := os.Getwd(); err == nil {
		if f := tryOpenLogFile(wd); f != nil {
			return f
		}
	}

	// Then, try OS temp dir.
	if f := tryOpenLogFile(os.TempDir()); f != nil {
		return f
	}

	// Finally, try the user's home directory.
	if u, err := user.Current(); err == nil && u.HomeDir != "" {
		if f := tryOpenLogFile(u.HomeDir); f != nil {
			return f
		}
	}

	return nil
}

func tryOpenLogFile(dir string) *os.File {
	if dir == "" {
		return nil
	}
	_ = os.MkdirAll(dir, 0o755)
	path := filepath.Join(dir, logFileName)
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0o644)
	if err != nil {
		return nil
	}
	return f
}

func Infof(message string, args ...any) {
	if logFile == nil {
		return
	}
	_, err := fmt.Fprintf(logFile, "[INFO] "+message+"\n", args...)
	if err != nil {
		panic("Error writing to log file: " + err.Error())
	}
}

func Debugf(message string, args ...any) {
	if logFile == nil {
		return
	}
	_, err := fmt.Fprintf(logFile, "[DEBUG] "+message+"\n", args...)
	if err != nil {
		panic("Error writing to log file: " + err.Error())
	}
}

func Errorf(message string, args ...any) {
	if logFile == nil {
		return
	}
	_, err := fmt.Fprintf(logFile, "[ERROR] "+message+"\n", args...)
	if err != nil {
		panic("Error writing to log file: " + err.Error())
	}
}
