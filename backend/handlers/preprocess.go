package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// PreprocessedData represents the extracted data from the resume.
type PreprocessedData struct {
	Text      string            `json:"text"`
	Entities  ExtractedEntities `json:"entities"`
	Skills    []string          `json:"skills"`
	Education []string          `json:"education"`
	ID        string            `json:"id"` // Add this field
}

// Update ExtractedEntities struct to match the actual response format
type ExtractedEntities struct {
	Name      string      `json:"name"`
	Email     []string    `json:"email"` // Keep this as []string to handle multiple emails
	Phone     string      `json:"phone"`
	Skills    []string    `json:"skills"`
	Education []Education `json:"education"`
}

type Education struct {
	Degree         string `json:"degree"`
	Institution    string `json:"institution"`
	Year           string `json:"year"`
	Location       string `json:"location"`
	Specialization string `json:"specialization"`  // Add this field
	GraduationDate string `json:"graduation_date"` // Add this field
}

// Add new types for job description
type JobDescription struct {
	RawText       string            `json:"raw_text"`
	ProcessedText string            `json:"processed_text"`
	Requirements  []string          `json:"requirements"`
	Skills        []string          `json:"required_skills"`
	Experience    map[string]string `json:"experience_requirements"`
}

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

// PreprocessResume handles resume preprocessing.
func PreprocessResume(c *fiber.Ctx) error {
	// Get the file from the request
	file, err := c.FormFile("resume")
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	// Get the uploaded file path
	uploadedFilePath := filepath.Join("uploads", "upload-"+file.Filename)

	// Check if file exists
	if _, err := os.Stat(uploadedFilePath); os.IsNotExist(err) {
		return c.Status(404).JSON(fiber.Map{
			"error": "File not found in uploads directory",
		})
	}

	// Read the uploaded file
	fileContent, err := os.ReadFile(uploadedFilePath)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Could not read uploaded file",
		})
	}

	// Create a buffer with the file content
	buf := bytes.NewBuffer(fileContent)

	fileExt := strings.ToLower(filepath.Ext(file.Filename))
	var extractedText string

	// Extract text based on file type
	switch fileExt {
	case ".pdf":
		cmd := exec.Command("pdftotext", "-", "-")
		cmd.Stdin = bytes.NewReader(buf.Bytes())
		var out bytes.Buffer
		cmd.Stdout = &out
		if err := cmd.Run(); err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Could not extract text from PDF",
			})
		}
		extractedText = out.String()
	case ".docx":
		extractedText = string(buf.Bytes())
	default:
		return c.Status(400).JSON(fiber.Map{
			"error": "Unsupported file format",
		})
	}

	// Log the extracted text
	log.Printf("Extracted text from resume: %s", extractedText)

	// Save extracted text
	err = saveProcessedText("resume", extractedText)
	if err != nil {
		log.Printf("Error saving resume text: %v", err)
	}

	// Preprocess text
	processedText := preprocessText(extractedText)

	// Generate a unique ID for the resume
	resumeID := fmt.Sprintf("resume_%d", time.Now().Unix())

	// Save processed text with the new function
	err = SaveProcessedText("resume", processedText, resumeID)
	if err != nil {
		log.Printf("Error saving resume text: %v", err)
	}

	// Extract entities using Gemini API
	entities, err := extractEntitiesWithGemini(processedText)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Entity extraction failed: " + err.Error(),
		})
	}

	// Extract skills and education using Random Forest models
	skills := extractSkillsWithModel(processedText)
	education := extractEducationWithModel(processedText)

	// Create response
	result := PreprocessedData{
		Text:      processedText,
		Entities:  entities,
		Skills:    skills,
		Education: education,
		ID:        resumeID, // Add this field to PreprocessedData struct
	}

	return c.JSON(result)
}

// preprocessText performs basic text preprocessing.
func preprocessText(text string) string {
	// Convert to lowercase
	text = strings.ToLower(text)
	// Remove extra whitespace
	text = strings.Join(strings.Fields(text), " ")
	// Remove non-printable characters
	text = strings.Map(func(r rune) rune {
		if r < 32 || r > 126 {
			return -1
		}
		return r
	}, text)
	return text
}

// Define structs to match Gemini's response structure
type Content struct {
	Parts []string `json:"Parts"`
	Role  string   `json:"Role"`
}

type Candidate struct {
	Content *Content `json:"Content"`
}

type ContentResponse struct {
	Candidates []Candidate `json:"Candidates"`
}

// Update the GeminiContent struct to match the actual response
type GeminiContent struct {
	Parts []string `json:"Parts"`
	Role  string   `json:"Role"`
}

type GeminiCandidate struct {
	Index        int           `json:"Index"`
	Content      GeminiContent `json:"Content"`
	FinishReason int           `json:"FinishReason"`
}

type GeminiResponse struct {
	Candidates []GeminiCandidate `json:"Candidates"`
}

// extractEntitiesWithGemini uses Google Gemini API for entity extraction.
func extractEntitiesWithGemini(text string) (ExtractedEntities, error) {
	var entities ExtractedEntities

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return entities, fiber.NewError(500, "GEMINI_API_KEY not found in environment")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Printf("Error initializing Gemini client: %v", err)
		return entities, err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-pro")
	prompt := `Extract the following entities from the text:
1. Name
2. Email
3. Phone
4. Skills
5. Education

Provide the output in JSON format with the keys: name, email, phone, skills, education.

Text: ` + text

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Error generating content from Gemini: %v", err)
		return entities, err
	}

	// Extract the JSON content from the response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return entities, fiber.NewError(500, "Empty response from Gemini")
	}

	// Convert genai.Part to string
	content := resp.Candidates[0].Content.Parts[0].(genai.Text)
	jsonStr := cleanJSONString(string(content))

	// Log the cleaned JSON string
	log.Printf("Cleaned JSON string: %s", jsonStr)

	// Create a temporary struct for initial parsing
	var rawResponse map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawResponse); err != nil {
		log.Printf("Error parsing raw JSON: %v", err)
		return entities, err
	}

	// Safely get string values with type checking
	if name, ok := rawResponse["name"].(string); ok {
		entities.Name = name
	}

	// Handle email field which can be either string or array
	if emailRaw, ok := rawResponse["email"].([]interface{}); ok {
		// Handle array of emails
		entities.Email = make([]string, 0, len(emailRaw))
		for _, e := range emailRaw {
			if emailStr, ok := e.(string); ok {
				entities.Email = append(entities.Email, emailStr)
			}
		}
	} else if emailStr, ok := rawResponse["email"].(string); ok {
		// Handle single email as string
		entities.Email = []string{emailStr}
	}

	if phone, ok := rawResponse["phone"].(string); ok {
		entities.Phone = phone
	}

	// Handle skills array with type checking
	if skillsRaw, ok := rawResponse["skills"].([]interface{}); ok {
		entities.Skills = make([]string, 0, len(skillsRaw))
		for _, skill := range skillsRaw {
			if skillStr, ok := skill.(string); ok {
				entities.Skills = append(entities.Skills, skillStr)
			}
		}
	}

	// Handle education array with type checking
	if eduRaw, ok := rawResponse["education"].([]interface{}); ok {
		entities.Education = make([]Education, 0, len(eduRaw))
		for _, edu := range eduRaw {
			if eduMap, ok := edu.(map[string]interface{}); ok {
				education := Education{
					Degree:         getString(eduMap, "degree"),
					Institution:    getString(eduMap, "institution"),
					Year:           getString(eduMap, "year"),
					Location:       getString(eduMap, "location"),
					Specialization: getString(eduMap, "specialization"),
					GraduationDate: getString(eduMap, "graduation_date"),
				}
				// Only add education entry if at least degree or institution is present
				if education.Degree != "" || education.Institution != "" {
					entities.Education = append(entities.Education, education)
				}
			}
		}
	}

	return entities, nil
}

// cleanJSONString removes markdown code block and extracts clean JSON
func cleanJSONString(content string) string {
	// Find the start and end of the JSON content
	startIndex := strings.Index(content, "{")
	endIndex := strings.LastIndex(content, "}")

	if startIndex == -1 || endIndex == -1 {
		log.Printf("Could not find valid JSON markers in content: %s", content)
		return ""
	}

	// Extract the JSON part
	jsonContent := content[startIndex : endIndex+1]

	// Remove any escaped newlines and normalize whitespace
	jsonContent = strings.ReplaceAll(jsonContent, "\\n", " ")
	jsonContent = strings.ReplaceAll(jsonContent, "\n", " ")

	return jsonContent
}

// saveProcessedText saves the processed text to a file
func saveProcessedText(prefix string, text string) error {
	// Create a timestamp for the filename
	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.txt", prefix, timestamp)

	// Create the output directory if it doesn't exist
	outputDir := "processed_texts"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	// Write the text to file
	return os.WriteFile(filepath.Join(outputDir, filename), []byte(text), 0644)
}

// extractSkillsWithModel extracts skills from text using a simple keyword approach for now
func extractSkillsWithModel(text string) []string {
	// TODO: Implement actual model-based extraction
	// For now, return an empty slice
	return []string{}
}

// extractEducationWithModel extracts education details from text using a simple keyword approach for now
func extractEducationWithModel(text string) []string {
	// TODO: Implement actual model-based extraction
	// For now, return an empty slice
	return []string{}
}

// PreprocessJobDescription handles job description preprocessing
func PreprocessJobDescription(c *fiber.Ctx) error {
	var data struct {
		Description string `json:"description"`
	}

	if err := c.BodyParser(&data); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Preprocess text
	processedText := preprocessText(data.Description)

	// Generate a unique ID for the job description
	jobID := fmt.Sprintf("job_%d", time.Now().Unix())

	// Save processed text
	err = SaveProcessedText("job", processedText, jobID)
	if err != nil {
		log.Printf("Error saving job description: %v", err)
	}

	// Initialize Gemini client
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return c.Status(500).JSON(fiber.Map{
			"error": "GEMINI_API_KEY not found in environment",
		})
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to initialize Gemini client",
		})
	}
	defer client.Close()

	// Extract requirements using Gemini API
	model := client.GenerativeModel("gemini-pro")
	prompt := `Analyze the following job description and extract:
    1. Required skills (both technical and soft skills)
    2. Experience requirements (years and level)
    3. Educational requirements
    4. Key responsibilities

    Format the output as a clean JSON object with these exact keys:
    {
        "skills": ["skill1", "skill2", ...],
        "experience": {
            "min_years": number,
            "level": "entry/mid/senior",
            "areas": ["area1", "area2", ...]
        },
        "education": {
            "degree": "required degree",
            "fields": ["field1", "field2", ...],
            "qualifications": ["qualification1", ...]
        },
        "responsibilities": ["responsibility1", "responsibility2", ...]
    }

    Job Description: ` + data.Description

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to process job description",
		})
	}

	// Extract content from response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return c.Status(500).JSON(fiber.Map{
			"error": "Empty response from Gemini",
		})
	}

	// Get the text content and clean it
	content := resp.Candidates[0].Content.Parts[0].(genai.Text)
	jsonStr := cleanJSONString(string(content))

	// Log the cleaned JSON for debugging
	log.Printf("Cleaned JSON string: %s", jsonStr)

	var requirements JobRequirements
	if err := json.Unmarshal([]byte(jsonStr), &requirements); err != nil {
		log.Printf("Error parsing requirements: %v, JSON: %s", err, jsonStr)
		// Try to provide a default structure if parsing fails
		requirements = JobRequirements{
			Skills:           []string{},
			Responsibilities: []string{},
		}
		requirements.Experience.Level = "entry"
		requirements.Experience.MinYears = 0
		requirements.Experience.Areas = []string{}
		requirements.Education.Degree = ""
		requirements.Education.Fields = []string{}
		requirements.Education.Qualifications = []string{}
	}

	return c.JSON(fiber.Map{
		"requirements": requirements,
		"id":           jobID, // Include the ID in the response
	})
}

// Helper function to safely get string values from map
func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}
