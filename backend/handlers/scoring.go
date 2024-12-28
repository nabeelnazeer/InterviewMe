package handlers

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ScoreResponse represents the scoring results
type ScoreResponse struct {
	OverallScore    float64            `json:"overall_score"`
	SkillsMatch     float64            `json:"skills_match"`
	ExperienceMatch float64            `json:"experience_match"`
	EducationMatch  float64            `json:"education_match"`
	DetailedScores  map[string]float64 `json:"detailed_scores"`
	Feedback        []string           `json:"feedback"`
}

// TextData represents the structure for storing processed texts
type TextData struct {
	ProcessedText string    `json:"processed_text"`
	Timestamp     time.Time `json:"timestamp"`
	Type          string    `json:"type"`
	ID            string    `json:"id"`
}

// ScoreResume handles the resume scoring endpoint
func ScoreResume(c *fiber.Ctx) error {
	var request struct {
		ResumeID string `json:"resume_id"`
		JobID    string `json:"job_id"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Load saved texts
	resumeText, err := loadProcessedText(request.ResumeID, "resume")
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Resume text not found",
		})
	}

	jobText, err := loadProcessedText(request.JobID, "job")
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Job description not found",
		})
	}

	// Prepare data for BERT model
	bertInput := prepareBertInput(resumeText, jobText)

	// TODO: Implement actual BERT model scoring
	// For now, return mock scoring
	mockScore := ScoreResponse{
		OverallScore:    0.85,
		SkillsMatch:     0.90,
		ExperienceMatch: 0.80,
		EducationMatch:  0.85,
		DetailedScores: map[string]float64{
			"technical_skills": 0.88,
			"soft_skills":      0.82,
			"qualifications":   0.85,
		},
		Feedback: []string{
			"Strong match in technical skills",
			"Consider highlighting more leadership experience",
		},
	}

	return c.JSON(mockScore)
}

// SaveProcessedText saves processed text with metadata
func SaveProcessedText(textType string, text string, id string) error {
	data := TextData{
		ProcessedText: text,
		Timestamp:     time.Now(),
		Type:          textType,
		ID:            id,
	}

	// Create directory if it doesn't exist
	outputDir := filepath.Join("processed_texts", textType)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	// Create filename with type and ID
	filename := fmt.Sprintf("%s_%s.json", textType, id)
	filePath := filepath.Join(outputDir, filename)

	// Marshal data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, jsonData, 0644)
}

// loadProcessedText retrieves saved processed text
func loadProcessedText(id string, textType string) (string, error) {
	filePath := filepath.Join("processed_texts", textType, fmt.Sprintf("%s_%s.json", textType, id))

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	var data TextData
	if err := json.Unmarshal(fileData, &data); err != nil {
		return "", err
	}

	return data.ProcessedText, nil
}

// prepareBertInput formats the text data for BERT model input
func prepareBertInput(resumeText, jobText string) map[string]interface{} {
	return map[string]interface{}{
		"text_pairs": [][]string{
			{resumeText, jobText},
		},
		"max_length": 512,
		"padding":    true,
		"truncation": true,
	}
}
