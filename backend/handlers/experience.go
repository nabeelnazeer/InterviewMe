package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type ProcessedExperience struct {
	Title          string   `json:"title"`
	Company        string   `json:"company"`
	Duration       string   `json:"duration"`
	Description    string   `json:"description"`
	RelevantSkills []string `json:"relevant_skills"`
	JobFitSummary  string   `json:"job_fit_summary"`
}

type ExperienceResponse struct {
	TotalYearsExperience float64               `json:"total_years_experience"`
	Experiences          []ProcessedExperience `json:"experiences"`
	OverallFit           string                `json:"overall_fit"`
}

// Update the experience handler to use filenames
func GetProcessedExperience(c *fiber.Ctx) error {
	resumeFilename := c.Query("resume_file")
	jobFilename := c.Query("job_file")

	log.Printf("Processing experience with files - Resume: %s, Job: %s", resumeFilename, jobFilename)

	if resumeFilename == "" || jobFilename == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Both resume_file and job_file are required",
		})
	}

	// Construct file paths directly
	resumePath := filepath.Join("processed_texts", "resume", resumeFilename)
	jobPath := filepath.Join("processed_texts", "job", jobFilename)

	log.Printf("Attempting to read files - Resume: %s, Job: %s", resumePath, jobPath)

	// Read files
	resumeData, err := os.ReadFile(resumePath)
	if err != nil {
		log.Printf("Error reading resume file: %v", err)
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("Resume file not found: %v", err),
		})
	}

	jobData, err := os.ReadFile(jobPath)
	if err != nil {
		log.Printf("Error reading job file: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to read job data: %v", err),
		})
	}

	// Parse resume and job data
	var resumeMap map[string]interface{}
	if err := json.Unmarshal(resumeData, &resumeMap); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to parse resume data",
		})
	}

	var jobMap map[string]interface{}
	if err := json.Unmarshal(jobData, &jobMap); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to parse job data",
		})
	}

	// Initialize Gemini
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to initialize AI",
		})
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")

	// Get raw experiences
	rawJSON := resumeMap["raw_json"].(string)
	var rawData map[string]interface{}
	json.Unmarshal([]byte(rawJSON), &rawData)
	experiences := rawData["experience"].([]interface{})

	// Process each experience
	var processedExperiences []ProcessedExperience
	var totalMonths float64

	jobDesc := jobMap["ProcessedText"].(string)
	jobReqs := jobMap["Requirements"].(map[string]interface{})

	for _, exp := range experiences {
		experience := exp.(map[string]interface{})

		// Calculate duration
		duration := experience["duration"].(string)
		months := calculateDurationInMonths(duration)
		totalMonths += months

		// Generate enhanced description with Gemini
		description := experience["description"].(string)
		enhancedDesc := generateEnhancedDescription(model, ctx, description, jobDesc)

		processed := ProcessedExperience{
			Title:          experience["title"].(string),
			Company:        experience["company"].(string),
			Duration:       duration,
			Description:    enhancedDesc,
			RelevantSkills: extractRelevantSkills(description, jobReqs),
			JobFitSummary:  analyzeJobFit(model, ctx, description, jobDesc),
		}

		processedExperiences = append(processedExperiences, processed)
	}

	// Generate overall fit analysis
	overallFit := analyzeOverallFit(model, ctx, processedExperiences, jobDesc)

	response := ExperienceResponse{
		TotalYearsExperience: totalMonths / 12.0,
		Experiences:          processedExperiences,
		OverallFit:           overallFit,
	}

	// Add the missing return statement to complete the function
	return c.JSON(response)
}

func analyzeJobFit(model *genai.GenerativeModel, ctx context.Context, expDesc, jobDesc string) string {
	prompt := fmt.Sprintf(
		`Analyze how well this experience matches the job requirements and provide a brief one-sentence summary:
        
        Job Description: %s
        Experience: %s`, jobDesc, expDesc)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "Analysis not available"
	}

	return string(resp.Candidates[0].Content.Parts[0].(genai.Text))
}

func analyzeOverallFit(model *genai.GenerativeModel, ctx context.Context, experiences []ProcessedExperience, jobDesc string) string {
	prompt := fmt.Sprintf(
		`Analyze the overall fit of the candidate's experience for this job and provide a concise summary:
        
        Job Description: %s
        
        Total Experience: %.1f years
        Key Roles: %s`,
		jobDesc,
		float64(len(experiences)),
		formatExperienceSummary(experiences))

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "Overall analysis not available"
	}

	return string(resp.Candidates[0].Content.Parts[0].(genai.Text))
}

func formatExperienceSummary(experiences []ProcessedExperience) string {
	var roles []string
	for _, exp := range experiences {
		roles = append(roles, fmt.Sprintf("%s at %s", exp.Title, exp.Company))
	}
	return strings.Join(roles, ", ")
}

func getMostRecentFile(dir string, prefix string) (string, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}

	var latestFile string
	var latestTime time.Time

	for _, file := range files {
		if strings.HasPrefix(file.Name(), prefix) {
			timeStr := strings.TrimPrefix(strings.TrimSuffix(file.Name(), ".json"), prefix+"_")
			fileTime, err := time.Parse("20060102_150405", timeStr)
			if err == nil && (latestFile == "" || fileTime.After(latestTime)) {
				latestTime = fileTime
				latestFile = file.Name()
			}
		}
	}

	if latestFile == "" {
		return "", fmt.Errorf("no files found with prefix %s", prefix)
	}

	return latestFile, nil
}

func calculateDurationInMonths(duration string) float64 {
	// Parse duration string (e.g., "January 2020 - Present" or "2019 - 2021")
	parts := strings.Split(duration, "-")
	if len(parts) != 2 {
		return 0
	}

	startDate, endDate := parseDate(strings.TrimSpace(parts[0])), parseDate(strings.TrimSpace(parts[1]))
	if startDate.IsZero() || endDate.IsZero() {
		return 0
	}

	months := endDate.Sub(startDate).Hours() / 24 / 30
	return months
}

func parseDate(dateStr string) time.Time {
	if strings.ToLower(dateStr) == "present" {
		return time.Now()
	}

	formats := []string{
		"January 2006",
		"Jan 2006",
		"2006",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t
		}
	}

	return time.Time{}
}

func extractRelevantSkills(description string, jobReqs map[string]interface{}) []string {
	var relevantSkills []string

	// Extract required skills from job requirements
	if skills, ok := jobReqs["skills"].([]interface{}); ok {
		requiredSkills := make(map[string]bool)
		for _, skill := range skills {
			if skillStr, ok := skill.(string); ok {
				requiredSkills[strings.ToLower(skillStr)] = true
			}
		}

		// Extract skills from experience description
		words := strings.Fields(strings.ToLower(description))
		for _, word := range words {
			if requiredSkills[word] {
				relevantSkills = append(relevantSkills, word)
			}
		}
	}

	return relevantSkills
}

func generateEnhancedDescription(model *genai.GenerativeModel, ctx context.Context, description string, jobDesc string) string {
	prompt := fmt.Sprintf(
		`Enhance this experience description to better align with the job requirements. 
        Make it more impactful and quantifiable where possible:
        
        Job Description: %s
        Experience Description: %s`, jobDesc, description)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return description // Return original description if enhancement fails
	}

	return string(resp.Candidates[0].Content.Parts[0].(genai.Text))
}

func ProcessExperience(data PreprocessedData) error {
	// Access extracted entities
	name := data.Name
	emails := data.Email
	phone := data.Phone

	// Use these entities in experience processing
	// ...existing experience processing logic...

	return nil
}
