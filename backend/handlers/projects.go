package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

const (
	maxPromptLength = 1000 // Maximum length for project and job descriptions
	maxProjects     = 5    // Maximum number of projects to analyze
	promptCooldown  = 2    // Seconds between Gemini API calls
)

type ProjectAnalysisRequest struct {
	ResumeID string `json:"resume_id"`
	JobID    string `json:"job_id"`
}

type ProjectAnalysis struct {
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	TechStack      []string `json:"tech_stack"`
	RelevanceToJob string   `json:"relevance_to_job"`
	MatchingSkills []string `json:"matching_skills"`
	// Removed RelevanceScore field
}

type ProjectAnalysisResponse struct {
	Projects []ProjectAnalysis `json:"projects"`
}

func AnalyzeProjects(c *fiber.Ctx) error {
	var req ProjectAnalysisRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error":   "Invalid request body",
			"details": err.Error(),
		})
	}

	// Log received data
	log.Printf("Received request with resume_id: %s, job_id: %s", req.ResumeID, req.JobID)

	// Handle empty IDs
	if req.ResumeID == "" || req.JobID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Missing resume_id or job_id",
		})
	}

	// Use IDs exactly as received - remove prefix handling code
	resumeData, err := loadTextData(req.ResumeID, "resume")
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Resume data not found",
		})
	}

	jobData, err := loadTextData(req.JobID, "job")
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Job data not found",
		})
	}

	// Check if there are any projects
	if len(resumeData.Entities.Projects) == 0 {
		return c.JSON(ProjectAnalysisResponse{
			Projects: []ProjectAnalysis{
				{
					Name:           "None",
					Description:    "No projects found in resume",
					TechStack:      []string{"None"},
					RelevanceToJob: "No projects to analyze",
					MatchingSkills: []string{"None"},
					// Removed RelevanceScore
				},
			},
		})
	}

	// Initialize Gemini with safety checks
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(getGeminiAPIKey()))
	if err != nil {
		log.Printf("Failed to create Gemini client: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to initialize AI service",
		})
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")

	// Add safety settings
	model.SetTemperature(0.7) // Balanced between creativity and consistency
	model.SetTopP(0.8)        // Reduce randomness
	model.SetTopK(40)         // Limit token choices

	var analysisResults []ProjectAnalysis
	processedProjects := 0

	// Sort projects by relevance to job description before analysis
	sortedProjects := preprocessProjects(resumeData.Entities.Projects, jobData.ProcessedText)

	for _, project := range sortedProjects {
		if processedProjects >= maxProjects {
			break // Limit the number of projects analyzed
		}

		// Clean and truncate project description
		cleanDesc := sanitizeText(project.Description, maxPromptLength)
		cleanJobDesc := sanitizeText(jobData.ProcessedText, maxPromptLength)

		// Enhanced prompt with better structure and constraints
		prompt := buildAnalysisPrompt(cleanDesc, cleanJobDesc)

		// Add cooldown between API calls
		time.Sleep(time.Second * promptCooldown)

		analysis, err := analyzeProjectWithRetry(ctx, model, prompt, project)
		if err != nil {
			log.Printf("Error analyzing project: %v", err)
			continue
		}

		// Validate and enhance the analysis
		analysis = validateAndEnhanceAnalysis(analysis, project, jobData)

		analysisResults = append(analysisResults, analysis)
		processedProjects++
	}

	return c.JSON(ProjectAnalysisResponse{
		Projects: analysisResults,
	})
}

// New helper functions

func sanitizeText(text string, maxLength int) string {
	// Remove special characters and excessive whitespace
	text = regexp.MustCompile(`[^\w\s-.,]`).ReplaceAllString(text, " ")
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
	text = strings.TrimSpace(text)

	// Truncate if too long
	if len(text) > maxLength {
		return text[:maxLength] + "..."
	}
	return text
}

func preprocessProjects(projects []Project, jobDesc string) []Project {
	// Calculate initial relevance scores
	type projectScore struct {
		project Project
		score   float64
	}

	var scoredProjects []projectScore

	for _, p := range projects {
		// Calculate TF-IDF similarity between project and job
		score := calculateTFIDFSimilarity(p.Description, jobDesc)
		scoredProjects = append(scoredProjects, projectScore{p, score})
	}

	// Sort by relevance score
	sort.Slice(scoredProjects, func(i, j int) bool {
		return scoredProjects[i].score > scoredProjects[j].score
	})

	// Return sorted projects
	result := make([]Project, len(scoredProjects))
	for i, ps := range scoredProjects {
		result[i] = ps.project
	}
	return result
}

func buildAnalysisPrompt(projectDesc, jobDesc string) string {
	return fmt.Sprintf(`Analyze this project concisely:
Project Description: %s

Key Job Requirements: %s

Provide a JSON response with:
{
	"description": "one clear sentence about what the project does",
	"tech_stack": ["only mentioned technologies", "max 5 key ones"],
	"relevance": "one sentence about job fit"
}`, projectDesc, extractKeyRequirements(jobDesc))
}

func extractKeyRequirements(jobDesc string) string {
	// Extract main requirements using keyword detection
	keywords := []string{"required", "must have", "key", "essential", "requirements"}
	sentences := strings.Split(jobDesc, ".")
	var relevant []string

	for _, sentence := range sentences {
		for _, keyword := range keywords {
			if strings.Contains(strings.ToLower(sentence), keyword) {
				relevant = append(relevant, strings.TrimSpace(sentence))
				break
			}
		}
	}

	// Return truncated requirements
	return strings.Join(relevant[:min(3, len(relevant))], ". ")
}

func analyzeProjectWithRetry(ctx context.Context, model *genai.GenerativeModel, prompt string, project Project) (ProjectAnalysis, error) {
	maxRetries := 3
	var lastErr error

	for i := 0; i < maxRetries; i++ {
		analysis, err := analyzeProject(ctx, model, prompt, project)
		if err == nil {
			return analysis, nil
		}
		lastErr = err
		time.Sleep(time.Second * time.Duration(i+1))
	}

	// Return fallback analysis if all retries fail
	return createFallbackAnalysis(project), lastErr
}

func validateAndEnhanceAnalysis(analysis ProjectAnalysis, project Project, jobData *TextData) ProjectAnalysis {
	// Ensure we have a valid description
	if analysis.Description == "" {
		analysis.Description = project.Description
	}

	// Validate and clean tech stack
	analysis.TechStack = validateTechStack(analysis.TechStack, project.Technologies)

	// Enhance matching skills analysis
	analysis.MatchingSkills = enhanceSkillMatching(analysis.TechStack, jobData.Requirements.Skills)

	// Removed relevance score calculation

	return analysis
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func findMatchingSkills(projectSkills []string, jobSkills []string) []string {
	var matching []string
	projectSkillsMap := make(map[string]bool)

	for _, skill := range projectSkills {
		projectSkillsMap[strings.ToLower(skill)] = true
	}

	for _, jobSkill := range jobSkills {
		if projectSkillsMap[strings.ToLower(jobSkill)] {
			matching = append(matching, jobSkill)
		}
	}

	return matching
}

func getGeminiAPIKey() string {
	key := os.Getenv("GEMINI_API_KEY")
	if key == "" {
		log.Printf("Warning: GEMINI_API_KEY not set in environment variables")
		return ""
	}
	return key
}

func calculateTFIDFSimilarity(text1, text2 string) float64 {
	// Simple TF-IDF implementation
	words1 := strings.Fields(strings.ToLower(text1))
	words2 := strings.Fields(strings.ToLower(text2))

	// Create word frequency maps
	freq1 := make(map[string]int)
	freq2 := make(map[string]int)
	for _, word := range words1 {
		freq1[word]++
	}
	for _, word := range words2 {
		freq2[word]++
	}

	// Calculate similarity
	var similarity float64
	for word, count1 := range freq1 {
		if count2, exists := freq2[word]; exists {
			similarity += float64(count1 * count2)
		}
	}

	// Normalize
	total := float64(len(words1) * len(words2))
	if total > 0 {
		return similarity / total
	}
	return 0
}

func analyzeProject(ctx context.Context, model *genai.GenerativeModel, prompt string, project Project) (ProjectAnalysis, error) {
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return ProjectAnalysis{}, err
	}

	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return ProjectAnalysis{}, fmt.Errorf("no response from model")
	}

	// Get text from the response
	var responseText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if textValue, ok := part.(genai.Text); ok {
			responseText = string(textValue)
			break
		}
	}

	// Parse the response into structured data
	var analysis ProjectAnalysis
	analysis.Name = project.Name
	if analysis.Name == "" {
		analysis.Name = strings.Split(project.Description, ".")[0]
	}

	// Parse the JSON response
	var geminiResponse struct {
		Description string   `json:"description"`
		TechStack   []string `json:"tech_stack"`
		Relevance   string   `json:"relevance"`
	}

	if err := json.Unmarshal([]byte(responseText), &geminiResponse); err != nil {
		return analysis, err
	}

	analysis.Description = geminiResponse.Description
	analysis.TechStack = geminiResponse.TechStack
	analysis.RelevanceToJob = geminiResponse.Relevance

	return analysis, nil
}

func createFallbackAnalysis(project Project) ProjectAnalysis {
	return ProjectAnalysis{
		Name:           project.Name,
		Description:    project.Description,
		TechStack:      project.Technologies,
		RelevanceToJob: "Could not analyze relevance",
		MatchingSkills: []string{},
	}
}

func validateTechStack(techStack []string, fallbackTech []string) []string {
	if len(techStack) == 0 {
		return fallbackTech
	}

	// Remove duplicates and empty strings
	uniqueTech := make(map[string]bool)
	var validTech []string

	for _, tech := range techStack {
		tech = strings.TrimSpace(tech)
		if tech != "" && !uniqueTech[tech] {
			uniqueTech[tech] = true
			validTech = append(validTech, tech)
		}
	}

	return validTech
}

func enhanceSkillMatching(projectSkills, jobSkills []string) []string {
	var enhancedMatches []string
	projectSkillsLower := make(map[string]bool)

	// Convert project skills to lowercase for matching
	for _, skill := range projectSkills {
		projectSkillsLower[strings.ToLower(skill)] = true
	}

	// Find matches including partial matches
	for _, jobSkill := range jobSkills {
		jobSkillLower := strings.ToLower(jobSkill)
		if projectSkillsLower[jobSkillLower] {
			enhancedMatches = append(enhancedMatches, jobSkill)
			continue
		}

		// Check for partial matches
		for projectSkill := range projectSkillsLower {
			if strings.Contains(projectSkill, jobSkillLower) ||
				strings.Contains(jobSkillLower, projectSkill) { // Fixed: contains -> Contains
				enhancedMatches = append(enhancedMatches, jobSkill)
				break
			}
		}
	}

	return enhancedMatches
}
