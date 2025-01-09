package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const (
	maxTokenLength = 800 // Maximum length for text sent to Gemini
	cooldownPeriod = 1   // Seconds between API calls
)

type ExperienceAnalysisRequest struct {
	ResumeID string `json:"resume_id"`
	JobID    string `json:"job_id"`
}

type ExperienceAnalysis struct {
	TotalYears        float64           `json:"total_years"`
	Roles             []ExperienceRole  `json:"roles"`
	SkillsGained      []string          `json:"skills_gained"`
	JobFitAnalysis    string            `json:"job_fit_analysis"`
	ResponsibilityMap map[string]string `json:"responsibility_map"`
}

type ExperienceRole struct {
	Title            string   `json:"title"`
	Company          string   `json:"company"`
	Duration         string   `json:"duration"`
	Skills           []string `json:"skills"`
	Responsibilities []string `json:"responsibilities"`
}

func AnalyzeExperience(c *fiber.Ctx) error {
	var req ExperienceAnalysisRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Validate IDs
	if req.ResumeID == "" || req.JobID == "" {
		return c.JSON(ExperienceAnalysis{}) // Return empty analysis if IDs missing
	}

	// Load data with null checks
	resumeData, err := loadTextData(req.ResumeID, "resume")
	if err != nil || resumeData == nil {
		return c.JSON(ExperienceAnalysis{})
	}

	jobData, err := loadTextData(req.JobID, "job")
	if err != nil || jobData == nil {
		return c.JSON(ExperienceAnalysis{})
	}

	// Check for empty experience
	if len(resumeData.Entities.Experience) == 0 {
		return c.JSON(ExperienceAnalysis{
			TotalYears: 0,
			Roles:      []ExperienceRole{},
		})
	}

	// Initialize Gemini with API key from environment
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		return c.JSON(ExperienceAnalysis{})
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")
	model.SetTemperature(0.7)

	analysis, err := performExperienceAnalysis(ctx, model, resumeData, jobData)
	if err != nil {
		return c.JSON(ExperienceAnalysis{})
	}

	return c.JSON(analysis)
}

func performExperienceAnalysis(ctx context.Context, model *genai.GenerativeModel, resumeData, jobData *TextData) (ExperienceAnalysis, error) {
	var analysis ExperienceAnalysis

	// Extract years (with truncated prompt)
	yearsPrompt := truncateText(fmt.Sprintf(
		"Calculate total years of experience from: %s",
		formatExperienceForPrompt(resumeData.Entities.Experience),
	), maxTokenLength)

	yearsResp, err := model.GenerateContent(ctx, genai.Text(yearsPrompt))
	if err == nil && len(yearsResp.Candidates) > 0 {
		analysis.TotalYears = extractNumber(yearsResp.Candidates[0].Content.Parts[0].(genai.Text))
	}

	// Process roles (with cooldown and truncation)
	var roles []ExperienceRole
	for _, exp := range resumeData.Entities.Experience {
		time.Sleep(time.Second * cooldownPeriod)

		rolePrompt := truncateText(fmt.Sprintf(`
            Analyze experience and list skills and responsibilities:
            Title: %s
            Company: %s
            Description: %s
            Format: {"skills":[],"responsibilities":[]}
        `, exp.Title, exp.Company, exp.Description), maxTokenLength)

		roleResp, err := model.GenerateContent(ctx, genai.Text(rolePrompt))
		if err != nil {
			continue
		}

		var roleAnalysis struct {
			Skills           []string `json:"skills"`
			Responsibilities []string `json:"responsibilities"`
		}

		if respText, ok := roleResp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			json.Unmarshal([]byte(respText), &roleAnalysis)
		}

		roles = append(roles, ExperienceRole{
			Title:            exp.Title,
			Company:          exp.Company,
			Duration:         exp.Duration,
			Skills:           roleAnalysis.Skills,
			Responsibilities: roleAnalysis.Responsibilities,
		})

		analysis.SkillsGained = append(analysis.SkillsGained, roleAnalysis.Skills...)
	}

	analysis.Roles = roles

	// Job fit analysis (truncated)
	matchingPrompt := truncateText(fmt.Sprintf(
		"Compare experience:%s with job:%s. List strengths and gaps.",
		formatExperienceForPrompt(resumeData.Entities.Experience),
		jobData.ProcessedText,
	), maxTokenLength)

	matchingResp, err := model.GenerateContent(ctx, genai.Text(matchingPrompt))
	if err == nil && len(matchingResp.Candidates) > 0 {
		analysis.JobFitAnalysis = string(matchingResp.Candidates[0].Content.Parts[0].(genai.Text))
	}

	// Map responsibilities (using existing helper)
	analysis.ResponsibilityMap = mapExperienceToResponsibilities(
		resumeData.Entities.Experience,
		jobData.Requirements.Responsibilities,
	)

	return analysis, nil
}

// Add new helper function for text truncation
func truncateText(text string, maxLength int) string {
	if len(text) <= maxLength {
		return text
	}
	return text[:maxLength] + "..."
}

func formatExperienceForPrompt(experience []Experience) string {
	var parts []string
	for _, exp := range experience {
		parts = append(parts, fmt.Sprintf(
			"Title: %s\nCompany: %s\nDuration: %s\nDescription: %s",
			exp.Title, exp.Company, exp.Duration, exp.Description,
		))
	}
	return strings.Join(parts, "\n\n")
}

func extractNumber(text genai.Text) float64 {
	var number float64
	fmt.Sscanf(string(text), "%f", &number)
	return number
}

func mapExperienceToResponsibilities(experience []Experience, jobResponsibilities []string) map[string]string {
	mapping := make(map[string]string)
	for _, resp := range jobResponsibilities {
		bestMatch := findBestExperienceMatch(resp, experience)
		if bestMatch != "" {
			mapping[resp] = bestMatch
		}
	}
	return mapping
}

func findBestExperienceMatch(responsibility string, experience []Experience) string {
	var bestMatch string
	maxSimilarity := 0.0

	for _, exp := range experience {
		similarity := calculateSimilarity(responsibility, exp.Description)
		if similarity > maxSimilarity {
			maxSimilarity = similarity
			bestMatch = fmt.Sprintf("%s at %s: %s", exp.Title, exp.Company, exp.Description)
		}
	}

	if maxSimilarity < 0.3 { // Threshold for considering it a match
		return ""
	}
	return bestMatch
}

func calculateSimilarity(text1, text2 string) float64 {
	// Simplified similarity calculation using word overlap
	words1 := strings.Fields(strings.ToLower(text1))
	words2 := strings.Fields(strings.ToLower(text2))

	wordSet := make(map[string]bool)
	for _, word := range words1 {
		wordSet[word] = true
	}

	matches := 0
	for _, word := range words2 {
		if wordSet[word] {
			matches++
		}
	}

	return float64(matches) / float64(len(words1)+len(words2)-matches)
}
