package utility

import (
	"path/filepath"
	"time"
)

type LogEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
	Time    time.Time  `json:"time"`
	Data    interface{} `json:"data, omitempty"`
}

const (
	logDir = "logs"
	latestFile = "latest"
	maxLogSize = 10 * 1024 * 1024 // 10 MB
	archiveDir = "archive"
)

var latestFilePath = filepath.Join(logDir, latestFile)

func initLogFolders() error {
	
}