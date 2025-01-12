package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Only keep experience-specific types here
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

func GetProcessedExperience(c *fiber.Ctx) error {
	resumeFilename := c.Query("resume_file")
	jobFilename := c.Query("job_file")

	log.Printf("GetProcessedExperience - Received filenames: resume=%s, job=%s",
		resumeFilename, jobFilename)

	if resumeFilename == "" || jobFilename == "" {
		log.Printf("GetProcessedExperience - Missing required files")
		return c.Status(400).JSON(fiber.Map{
			"error": "Both resume_file and job_file are required",
		})
	}

	// Ensure filenames have correct extensions
	if !strings.HasSuffix(resumeFilename, ".json") {
		resumeFilename += ".json"
	}
	if !strings.HasSuffix(jobFilename, ".json") {
		jobFilename += ".json"
	}

	// Clean up any duplicate prefixes
	resumeFilename = strings.TrimPrefix(resumeFilename, "resume_resume_")
	resumeFilename = strings.TrimPrefix(resumeFilename, "upload-")
	if !strings.HasPrefix(resumeFilename, "resume_") {
		resumeFilename = "resume_" + resumeFilename
	}

	jobFilename = strings.TrimPrefix(jobFilename, "job_job_")
	if !strings.HasPrefix(jobFilename, "job_") {
		jobFilename = "job_" + jobFilename
	}

	// Construct the full file paths
	resumePath := fmt.Sprintf("processed_texts/resume/%s", resumeFilename)
	jobPath := fmt.Sprintf("processed_texts/job/%s", jobFilename)

	log.Printf("GetProcessedExperience - Processing files: resume=%s, job=%s",
		resumePath, jobPath)

	fmt.Printf("Looking for resume at: %s\n", resumePath)
	fmt.Printf("Looking for job at: %s\n", jobPath)

	// Load the resume data using the full path
	resumeData, err := LoadTextDataFromPath(resumePath)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("Resume file not found at %s: %v", resumePath, err),
		})
	}

	// Verify that we have experience data
	if len(resumeData.Entities.Experience) == 0 {
		return c.Status(200).JSON(ExperienceResponse{
			TotalYearsExperience: 0,
			Experiences:          []ProcessedExperience{},
			OverallFit:           "No experience data found in resume",
		})
	}

	// Load the job data using the full path
	jobData, err := LoadTextDataFromPath(jobPath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to read job data at %s: %v", jobPath, err),
		})
	}

	// Initialize Gemini
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to initialize AI client",
		})
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")

	// Process each experience
	var processedExperiences []ProcessedExperience
	var totalMonths float64

	// Log the experience data we're processing
	fmt.Printf("Processing %d experiences\n", len(resumeData.Entities.Experience))

	for _, exp := range resumeData.Entities.Experience {
		fmt.Printf("Processing experience: %s at %s\n", exp.Title, exp.Company)

		months := calculateDurationInMonths(exp.Duration)
		totalMonths += months

		// Extract skills from the experience description
		relevantSkills := extractRelevantSkills(exp.Description, jobData.Requirements)

		processed := ProcessedExperience{
			Title:          exp.Title,
			Company:        exp.Company,
			Duration:       exp.Duration,
			Description:    generateEnhancedDescription(model, ctx, exp.Description, jobData.ProcessedText),
			RelevantSkills: relevantSkills,
			JobFitSummary:  analyzeJobFit(model, ctx, exp.Description, jobData.ProcessedText),
		}

		processedExperiences = append(processedExperiences, processed)
	}

	// Generate overall fit analysis
	overallFit := analyzeOverallFit(model, ctx, processedExperiences, jobData.ProcessedText)

	response := ExperienceResponse{
		TotalYearsExperience: totalMonths / 12.0, // Convert months to years
		Experiences:          processedExperiences,
		OverallFit:           overallFit,
	}

	// Log the response before sending
	fmt.Printf("Sending response with %d experiences\n", len(processedExperiences))

	return c.JSON(response)
}

func AnalyzeExperience(c *fiber.Ctx) error {
	resumeId := c.Query("resume_id")
	jobId := c.Query("job_id")

	log.Printf("AnalyzeExperience - Received IDs: resume_id=%s, job_id=%s",
		resumeId, jobId)

	if resumeId == "" || jobId == "" {
		log.Printf("AnalyzeExperience - Missing required IDs")
		return c.Status(400).JSON(fiber.Map{
			"error": "Both resume_id and job_id are required",
		})
	}

	// Construct complete filenames
	resumeFilename := fmt.Sprintf("resume_%s.json", resumeId)
	jobFilename := fmt.Sprintf("job_%s.json", jobId)

	log.Printf("AnalyzeExperience - Constructed filenames: resume=%s, job=%s",
		resumeFilename, jobFilename)

	// Load data using existing LoadTextData function
	resumeData, err := LoadTextData(resumeFilename, "resume")
	if err != nil {
		log.Printf("AnalyzeExperience - Error loading resume: %v", err)
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("Resume not found: %v", err),
		})
	}

	jobData, err := LoadTextData(jobFilename, "job")
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": fmt.Sprintf("Job not found: %v", err),
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

	var processedExperiences []ProcessedExperience
	var totalMonths float64

	// Process each experience
	for _, exp := range resumeData.Entities.Experience {
		months := calculateDurationInMonths(exp.Duration)
		totalMonths += months

		// Generate enhanced description using Gemini
		enhancedDesc := generateEnhancedDescription(model, ctx, exp.Description, jobData.ProcessedText)

		// Extract relevant skills
		relevantSkills := extractRelevantSkills(exp.Description, jobData.Requirements)

		// Analyze job fit
		jobFit := analyzeJobFit(model, ctx, exp.Description, jobData.ProcessedText)

		processed := ProcessedExperience{
			Title:          exp.Title,
			Company:        exp.Company,
			Duration:       exp.Duration,
			Description:    enhancedDesc,
			RelevantSkills: relevantSkills,
			JobFitSummary:  jobFit,
		}

		processedExperiences = append(processedExperiences, processed)
	}

	// Generate overall fit analysis
	overallFit := analyzeOverallFit(model, ctx, processedExperiences, jobData.ProcessedText)

	response := ExperienceResponse{
		TotalYearsExperience: totalMonths / 12.0,
		Experiences:          processedExperiences,
		OverallFit:           overallFit,
	}

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

// Update extractRelevantSkills to work with JobRequirements instead of map
func extractRelevantSkills(description string, jobReqs JobRequirements) []string {
	var relevantSkills []string
	requiredSkills := make(map[string]bool)

	// Add all job requirements skills to the map
	for _, skill := range jobReqs.Skills {
		requiredSkills[strings.ToLower(skill)] = true
	}

	// Extract skills from experience description
	words := strings.Fields(strings.ToLower(description))
	for _, word := range words {
		if requiredSkills[word] {
			relevantSkills = append(relevantSkills, word)
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

// Add this new function to load data directly from path
func LoadTextDataFromPath(filepath string) (*ProcessedText, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	var processedText ProcessedText
	if err := json.Unmarshal(data, &processedText); err != nil {
		return nil, err
	}

	return &processedText, nil
}
