package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

type JSONLog struct {
	Timestamp time.Time      `json:"timestamp"`
	Data      map[string]any `json:"data"`
	RawJSON   string         `json:"raw_json"`
}

// ProcessingSession struct removed as it's defined in filetransfer.go

func SaveJSONLog(data map[string]any, rawJSON string, prefix string) error {
	// Create logs directory if it doesn't exist
	logsDir := filepath.Join("processed_texts", "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	log := JSONLog{
		Timestamp: time.Now(),
		Data:      data,
		RawJSON:   rawJSON,
	}

	// Create filename with timestamp
	filename := fmt.Sprintf("%s_%s.json", prefix, time.Now().Format("20060102_150405"))
	filePath := filepath.Join(logsDir, filename)

	// Marshal to JSON with indentation
	jsonData, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, jsonData, 0644)
}

// InitializeJSONLogger sets up a JSON logger.
func InitializeJSONLogger() {
	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("JSON Logger initialized")
}
