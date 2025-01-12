package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// ProcessingSession represents a processing session with resume and job files
type ProcessingSession struct {
	ResumeLogFile string `json:"resume_log_file"`
	JobFile       string `json:"job_file"`
	Timestamp     string `json:"timestamp"`
	SessionID     string `json:"session_id"`
	FilePath      string `json:"file_path"`
}

// SaveProcessingSession saves the session information with resume and job files.
func SaveProcessingSession(resumeLog, jobFile string) (string, error) {
	if resumeLog != "" {
		if err := ValidateFilePath(filepath.Join("processed_texts", "resume", resumeLog)); err != nil {
			return "", fmt.Errorf("invalid resume file: %v", err)
		}
	}

	if jobFile != "" {
		if err := ValidateFilePath(filepath.Join("processed_texts", "job", jobFile)); err != nil {
			return "", fmt.Errorf("invalid job file: %v", err)
		}
	}

	session := ProcessingSession{
		ResumeLogFile: resumeLog,
		JobFile:       jobFile,
		Timestamp:     GetTimestamp(),
		SessionID:     fmt.Sprintf("session_%s", GetTimestamp()),
		FilePath:      "", // Set if needed
	}

	// Create sessions directory
	sessionsDir := filepath.Join("processed_texts", "sessions")
	if err := os.MkdirAll(sessionsDir, 0755); err != nil {
		log.Printf("Error creating sessions directory: %v", err) // Added logging
		return "", err
	}

	// Save session data
	sessionPath := filepath.Join(sessionsDir, session.SessionID+".json")
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		log.Printf("Error marshaling session data: %v", err) // Added logging
		return "", err
	}

	if err := os.WriteFile(sessionPath, data, 0644); err != nil {
		log.Printf("Error writing session file %s: %v", sessionPath, err) // Added logging
		return "", err
	}

	log.Printf("Session saved successfully with ID: %s", session.SessionID) // Added logging
	return session.SessionID, nil
}

// GetProcessingSession retrieves the session data based on the session ID.
func GetProcessingSession(sessionID string) (*ProcessingSession, error) {
	sessionPath := filepath.Join("processed_texts", "sessions", sessionID+".json")
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		return nil, err
	}

	var session ProcessingSession
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func GetTimestamp() string {
	return time.Now().Format("20060102_150405")
}

// ValidateFilePath checks if a file exists and is accessible
func ValidateFilePath(path string) error {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// Check if file exists
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil // It's OK if file doesn't exist yet
		}
		return fmt.Errorf("error accessing file: %s - %v", path, err)
	}
	return nil
}
