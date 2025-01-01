package handlers

import "time"

// TextData represents the structure for storing processed texts
type TextData struct {
	ProcessedText   string            `json:"processed_text"`
	Timestamp       time.Time         `json:"timestamp"`
	Type            string            `json:"type"`
	ID              string            `json:"id"`
	Entities        ExtractedEntities `json:"entities"`
	Requirements    JobRequirements   `json:"requirements"`
	SoftSkills      []string          `json:"soft_skills"`
	TechnicalSkills []string          `json:"technical_skills"`
}
