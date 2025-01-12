package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// SavePreprocessLog saves the raw cleaned JSON string to a log file
func SavePreprocessLog(jsonStr string, logType string) error {
	// Create logs directory
	logsDir := filepath.Join("processed_texts", "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	// Create filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.json", logType, timestamp)
	filePath := filepath.Join(logsDir, filename)

	// Format the JSON with indentation
	var jsonData interface{}
	if err := json.Unmarshal([]byte(jsonStr), &jsonData); err != nil {
		return err
	}

	prettyJSON, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(filePath, prettyJSON, 0644)
}

// SaveRawJSONLog saves only the cleaned JSON string directly to a file
func SaveRawJSONLog(jsonStr string, logType string) error {
	// Create logs directory
	logsDir := filepath.Join("processed_texts", "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	// Create filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.json", logType, timestamp)
	filePath := filepath.Join(logsDir, filename)

	// Write the raw JSON string directly to file
	return os.WriteFile(filePath, []byte(jsonStr), 0644)
}

// SaveCleanJSON saves only the cleaned JSON string to a file
func SaveCleanJSON(jsonStr string, logType string) error {
	// Create logs directory
	logsDir := filepath.Join("processed_texts", "logs")
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return err
	}

	// Create filename with timestamp
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.json", logType, timestamp)
	filePath := filepath.Join(logsDir, filename)

	// Write just the JSON string directly
	return os.WriteFile(filePath, []byte(jsonStr), 0644)
}

func SaveLogs(data PreprocessedData) error {
	logEntry := map[string]interface{}{
		"id":    data.ID,
		"name":  data.Name,
		"email": data.Email,
		"phone": data.Phone,
		// ...existing fields...
	}

	// Save the log entry
	// ...existing logging implementation...
}
