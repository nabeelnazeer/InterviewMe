package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type PreprocessedData struct {
	ProcessedText   string            `json:"processed_text"`
	Text            string            `json:"text"`
	Entities        ExtractedEntities `json:"entities"`
	RawText         string            `json:"raw_text"`
	RawJSON         string            `json:"raw_json"`
	Name            string            `json:"name"`
	Email           string            `json:"email"`
	Phone           string            `json:"phone"`
	Requirements    JobRequirements   `json:"requirements,omitempty"`
	TechnicalSkills []string          `json:"technical_skills"`
	SoftSkills      []string          `json:"soft_skills"`
	Education       []Education       `json:"education"`
	Experience      []Experience      `json:"experience"`
	Projects        []Project         `json:"projects"`
	SessionID       string            `json:"session_id"`
	Filename        string            `json:"filename"`
	ProcessedAt     time.Time         `json:"processed_at"`
	ID              string            `json:"id"`
}

// Add LogEntry type definition
type LogEntry struct {
	Timestamp string           `json:"timestamp"`
	LogType   string           `json:"log_type"`
	Data      PreprocessedData `json:"data"`
}

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

func SaveLog(data PreprocessedData, logType string) error {
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.json", logType, timestamp)

	logData := LogEntry{
		Timestamp: timestamp,
		LogType:   logType,
		Data:      data,
	}

	jsonData, err := json.Marshal(logData)
	if err != nil {
		return fmt.Errorf("error marshaling log data: %v", err)
	}

	logDir := filepath.Join("logs", logType)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("error creating log directory: %v", err)
	}

	filePath := filepath.Join(logDir, filename)
	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing log file: %v", err)
	}

	return nil
}
