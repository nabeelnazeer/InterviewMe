package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TextData represents the processed text data structure
type TextData struct {
	ProcessedText   string            `json:"processed_text"`
	Timestamp       time.Time         `json:"timestamp"`
	Type            string            `json:"type"`
	ID              string            `json:"id"`
	Entities        ExtractedEntities `json:"entities"`
	Requirements    JobRequirements   `json:"requirements,omitempty"`
	SoftSkills      []string          `json:"soft_skills,omitempty"`
	TechnicalSkills []string          `json:"technical_skills,omitempty"`
	RawJSON         string            `json:"raw_json,omitempty"`
}

// ExtractedEntities represents the entities extracted from text
type ExtractedEntities struct {
	Name       string       `json:"name"`
	Email      []string     `json:"email"`
	Phone      string       `json:"phone"`
	Skills     []string     `json:"skills"`
	Education  []Education  `json:"education"`
	Projects   []Project    `json:"projects"`
	Experience []Experience `json:"experience"`
}

// Education represents educational background
type Education struct {
	Degree         string `json:"degree"`
	Institution    string `json:"institution"`
	Year           string `json:"year"`
	Location       string `json:"location"`
	Specialization string `json:"specialization"`
	GraduationDate string `json:"graduation_date"`
}

// Project represents a project entry
type Project struct {
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Technologies []string `json:"technologies"`
	Duration     string   `json:"duration"`
	Role         string   `json:"role"`
	Timeline     string   `json:"timeline"`
	Team         []string `json:"team"`
	Achievements []string `json:"achievements"`
	Status       string   `json:"status"`
}

// Experience represents work experience
type Experience struct {
	Title            string   `json:"title"`
	Company          string   `json:"company"`
	Duration         string   `json:"duration"`
	Location         string   `json:"location"`
	Description      string   `json:"description"`
	Skills           []string `json:"skills"`
	Responsibilities []string `json:"responsibilities"`
	Achievements     []string `json:"achievements"`
	TeamSize         int      `json:"team_size"`
	Level            string   `json:"level"`
	RoleDescription  string   `json:"role_description"`
}

// JobRequirements represents job requirements
type JobRequirements struct {
	Skills     []string `json:"skills"`
	Experience struct {
		MinYears int      `json:"min_years"`
		Level    string   `json:"level"`
		Areas    []string `json:"areas"`
	} `json:"experience"`
	Education struct {
		Degree         string   `json:"degree"`
		Fields         []string `json:"fields"`
		Qualifications []string `json:"qualifications"`
	} `json:"education"`
	Responsibilities []string `json:"responsibilities"`
}

type Entities struct {
	Name  string   `json:"name"`
	Email []string `json:"email"`
	Phone string   `json:"phone"`
	// ...existing fields...
}

// LoadTextData loads processed text data from file
func LoadTextData(id string, textType string) (*TextData, error) {
	// Clean the ID by removing any prefix if present
	cleanID := strings.TrimPrefix(id, textType+"_")
	cleanID = strings.TrimSuffix(cleanID, ".json")

	// Try different possible file paths
	possiblePaths := []string{
		filepath.Join("processed_texts", textType, fmt.Sprintf("%s_%s.json", textType, cleanID)),
		filepath.Join("processed_texts", textType, id),
		filepath.Join("processed_texts", textType, fmt.Sprintf("%s.json", id)),
	}

	var fileData []byte
	var err error
	var successPath string

	// Try each possible path
	for _, path := range possiblePaths {
		log.Printf("Attempting to load file: %s", path)
		if fileData, err = os.ReadFile(path); err == nil {
			successPath = path
			break
		}
	}

	if err != nil {
		log.Printf("Error reading file from all attempted paths: %v", err)
		return nil, err
	}

	log.Printf("Successfully loaded file from: %s", successPath)

	var data TextData
	if err := json.Unmarshal(fileData, &data); err != nil {
		return nil, err
	}

	return &data, nil
}
