package utility

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
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
	// First create the log folder
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		if err := os.Mkdir(logDir, 0755); err != nil {
			return fmt.Errorf("failed to create logs directory: %w", err)
		}
	}

	// Then create the archive folder
	arc := filepath.Join(logDir, archiveDir)
	if _, err := os.Stat(arc); os.IsNotExist(err) {
		if err := os.Mkdir(arc, 0755); err != nil {
			return  fmt.Errorf("failed to create an archive directory: %w", err)
		}
	}

	return  nil
}

func rotateLogs() error {
	// Check file size
	info, err := os.Stat(latestFilePath)
	if os.IsNotExist(err) {
		return nil // No log file yet, nothing to rotate
	}

	if err != nil {
		return  fmt.Errorf("stat error: %w", err)
	}

	if info.Size() < maxLogSize {
		return nil // No need to rotate
	}

	// Rotate the log
	timestamp := time.Now().Format("20060102_150405")
	rotatedName := filepath.Join(logDir, archiveDir, fmt.Sprintf("%s.log.gz", timestamp))

	oldFile, err := os.Open(latestFilePath)
	if err != nil {
		return fmt.Errorf("failed to open latest log for rotation: %w", err)
	}
	defer oldFile.Close()

	// Create gzip archive
	outFile, err := os.Create(rotatedName)
	if err != nil {
		return fmt.Errorf("failed to create rotated log file: %w", err)
	}
	gz := gzip.NewWriter(outFile)

	// Copy contents to gzip writer
	if _, err := io.Copy(gz, oldFile); err != nil {
		gz.Close()
		outFile.Close()
		return fmt.Errorf("failed to write to gzip file: %w", err)
	}

	gz.Close()
	outFile.Close()

	os.Remove(latestFilePath)

	// Create new empty latest log file
	_, err = os.Create(latestFilePath)
	if err != nil {
		return fmt.Errorf("failed to create new latest log file: %w", err)
	}

	return nil
}