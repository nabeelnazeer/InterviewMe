package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type ExperienceResponse struct {
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Duration         string   `json:"duration"`
	Responsibilities []string `json:"responsibilities"`
	MatchScore       float64  `json:"match_score"`
}

// Add new struct for raw JSON data
type RawSkillsData struct {
	Skills []string `json:"skills"`
}

// Add new function to handle raw JSON string
func ProcessRawExperience(c *fiber.Ctx) error {
	var data struct {
		JsonString string `json:"json_string"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Clean the JSON string
	cleanedJSON := cleanJSONString(data.JsonString)

	// Parse the cleaned JSON
	var skillsData RawSkillsData
	if err := json.Unmarshal([]byte(cleanedJSON), &skillsData); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to parse JSON string: " + err.Error(),
		})
	}

	// Categorize skills
	technicalSkills := FilterTechnicalSkills(skillsData.Skills)
	softSkills := filterSoftSkills(skillsData.Skills)

	// Initialize Gemini for skill analysis
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to initialize Gemini",
		})
	}
	defer client.Close()

	// Analyze skills with Gemini
	model := client.GenerativeModel("gemini-pro")
	skillsAnalysis, err := analyzeSkillsWithGemini(model, ctx, skillsData.Skills)
	if err != nil {
		log.Printf("Error analyzing skills: %v", err)
		// Continue without Gemini analysis
	}

	return c.JSON(fiber.Map{
		"raw_skills":       skillsData.Skills,
		"technical_skills": technicalSkills,
		"soft_skills":      softSkills,
		"skills_analysis":  skillsAnalysis,
	})
}

func analyzeSkillsWithGemini(model *genai.GenerativeModel, ctx context.Context, skills []string) (string, error) {
	prompt := fmt.Sprintf(`Analyze these skills and provide a brief summary of the candidate's technical profile:
    Skills: %v
    Please focus on:
    1. Main technical areas
    2. Experience level indicated by the skill set
    3. Potential roles suitable for this skill combination
    Provide a concise 2-3 sentence response.`, skills)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		return string(resp.Candidates[0].Content.Parts[0].(genai.Text)), nil
	}

	return "", fmt.Errorf("no content generated")
}

// Update main handler to support both file and raw JSON processing
func GetProcessedExperience(c *fiber.Ctx) error {
	// Check if raw JSON string is provided
	if c.Get("Content-Type") == "application/json" {
		return ProcessRawExperience(c)
	}

	// Get resume ID and job ID from query parameters
	resumeID := c.Query("resume_id")
	jobID := c.Query("job_id")

	if resumeID == "" || jobID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Resume ID and Job ID are required",
		})
	}

	// Read the processed resume file
	resumePath := "processed_texts/resume/resume_" + resumeID + ".json"
	resumeData, err := os.ReadFile(resumePath)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Resume data not found",
		})
	}

	// Read the processed job file
	jobPath := "processed_texts/job/job_" + jobID + ".json"
	jobData, err := os.ReadFile(jobPath)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Job data not found",
		})
	}

	// Parse the JSON data
	var resumeJSON map[string]interface{}
	var jobJSON map[string]interface{}

	if err := json.Unmarshal(resumeData, &resumeJSON); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to parse resume data",
		})
	}

	if err := json.Unmarshal(jobData, &jobJSON); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to parse job data",
		})
	}

	// Initialize Gemini client
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to initialize Gemini",
		})
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")

	// Extract experiences from resume
	experiences := resumeJSON["Entities"].(map[string]interface{})["Experience"].([]interface{})
	var processedExperiences []ExperienceResponse

	for _, exp := range experiences {
		experience := exp.(map[string]interface{})

		// Generate concise description using Gemini
		prompt := "Generate a concise one-sentence description for this job experience: " +
			experience["description"].(string)

		resp, err := model.GenerateContent(ctx, genai.Text(prompt))
		if err != nil {
			log.Printf("Error generating description: %v", err)
			continue
		}

		// Calculate match score based on job requirements
		matchScore := calculateMatchScore(experience, jobJSON)

		processedExp := ExperienceResponse{
			Title:            experience["title"].(string),
			Description:      string(resp.Candidates[0].Content.Parts[0].(genai.Text)),
			Duration:         experience["duration"].(string),
			Responsibilities: convertToStringSlice(experience["responsibilities"]),
			MatchScore:       matchScore,
		}

		processedExperiences = append(processedExperiences, processedExp)
	}

	return c.JSON(fiber.Map{
		"experiences": processedExperiences,
	})
}

func calculateMatchScore(experience map[string]interface{}, jobData map[string]interface{}) float64 {
	// Get job requirements
	requirements := jobData["Requirements"].(map[string]interface{})
	requiredSkills := requirements["skills"].([]interface{})

	// Get experience skills
	expSkills := experience["skills"].([]interface{})

	// Calculate match score based on skills overlap
	matchingSkills := 0
	for _, reqSkill := range requiredSkills {
		for _, expSkill := range expSkills {
			if strings.ToLower(reqSkill.(string)) == strings.ToLower(expSkill.(string)) {
				matchingSkills++
				break
			}
		}
	}

	return float64(matchingSkills) / float64(len(requiredSkills))
}

func convertToStringSlice(i interface{}) []string {
	if i == nil {
		return []string{}
	}

	interfaceSlice := i.([]interface{})
	stringSlice := make([]string, len(interfaceSlice))

	for i, v := range interfaceSlice {
		stringSlice[i] = v.(string)
	}

	return stringSlice
}

// AnalyzeExperience handles experience analysis
func AnalyzeExperience(c *fiber.Ctx) error {
	return GetProcessedExperience(c) // Reuse existing functionality
}

// ProcessExperience handles experience processing
func ProcessExperience(c *fiber.Ctx) error {
	return GetProcessedExperience(c) // Reuse existing functionality
}
