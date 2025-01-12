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
	"regexp"
	"strconv"
	"strings"
	"time"

	"interviewme/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/google/generative-ai-go/genai"
	"github.com/jdkato/prose/v2"
	"google.golang.org/api/option"
)

// Add new types for job description
type JobDescription struct {
	RawText       string            `json:"raw_text"`
	ProcessedText string            `json:"processed_text"`
	Requirements  []string          `json:"requirements"`
	Skills        []string          `json:"required_skills"`
	Experience    map[string]string `json:"experience_requirements"`
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
	timestamp := time.Now().Unix()
	resumeID := fmt.Sprintf("%d", timestamp)
	filename := fmt.Sprintf("resume_%s.json", resumeID) // Update filename format
	filePath := filepath.Join("processed_texts", "resume", filename)

	// Ensure directory exists before saving
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		log.Printf("Error creating directory structure: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create directory structure",
		})
	}

	// Extract entities using Gemini API first
	entities, err := extractEntitiesWithGemini(processedText)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Entity extraction failed: " + err.Error(),
		})
	}

	// Validate and clean extracted entities
	validateExtractedEntities(&entities)

	// Save processed text with entities
	if err := SaveProcessedText("resume", processedText, resumeID, entities); err != nil {
		log.Printf("Error saving resume text: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save processed text",
		})
	}

	// Wait briefly to ensure file is written
	time.Sleep(100 * time.Millisecond)

	// Verify file exists after saving
	if _, err := os.Stat(filePath); err != nil {
		log.Printf("Error verifying saved file: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to verify saved file: %v", err),
		})
	}

	// After ensuring file exists, create session
	sessionID, err := utils.SaveProcessingSession(filename, "")
	if err != nil {
		log.Printf("Error saving resume session: %v", err)
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save resume session",
		})
	}

	// Update the result initialization to match PreprocessedData structure
	result := PreprocessedData{
		ProcessedText:   processedText,
		Text:            processedText,
		Entities:        entities,
		RawText:         extractedText,
		TechnicalSkills: FilterTechnicalSkills(entities.Skills),
		SoftSkills:      filterSoftSkills(entities.Skills),
		Education:       entities.Education,
		Experience:      entities.Experience,
		Projects:        entities.Projects,
		SessionID:       sessionID,
		Filename:        filename,
		ProcessedAt:     time.Now(),
		ID:              resumeID,
	}

	// Ensure Name is never undefined
	if result.Entities.Name == "" {
		result.Entities.Name = "Not provided"
	}

	// Log the response for debugging
	log.Printf("Sending response: %+v", result)

	// Return file information and session ID
	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}

// preprocessText performs basic text preprocessing.
func preprocessText(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	// Remove non-printable characters
	text = strings.Map(func(r rune) rune {
		if r < 32 || r > 126 {
			return -1
		}
		return r
	}, text)

	// Add NLP preprocessing
	doc, _ := prose.NewDocument(text)

	// Tokenization and lemmatization
	var processed []string
	for _, token := range doc.Tokens() {
		processed = append(processed, token.Text)
	}

	// Remove stopwords
	processed = removeStopwords(processed)

	return strings.Join(processed, " ")
}

func removeStopwords(tokens []string) []string {
	stopwords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true,
		"or": true, "but": true, "in": true, "on": true,
		// Add more stopwords...
	}

	var filtered []string
	for _, token := range tokens {
		if !stopwords[strings.ToLower(token)] {
			filtered = append(filtered, token)
		}
	}
	return filtered
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
4. Technical Skills (programming languages, tools, technologies)
5. Soft Skills (leadership, communication, etc.)
6. Education (with degree, institution, year, specialization)
7. Projects (with name, description, technologies used, duration, role)
8. Experience (with title, company, duration, location, responsibilities)

Provide the output in JSON format with the keys: 
name, email, phone, technical_skills, soft_skills, education, projects, experience.
Each skill type should be an array of strings.
Education should include degree, institution, year, location, specialization.
Projects should be an array of objects with name, description, technologies, duration, role.
Experience should be an array of objects with title, company, duration, location, description, responsibilities.

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

	// Save just the cleaned JSON string
	if err := SaveCleanJSON(jsonStr, "resume"); err != nil {
		log.Printf("Error saving resume JSON log: %v", err)
	}
	log.Printf("Cleaned JSON: %s", jsonStr)

	// Create a temporary struct for initial parsing
	var rawResponse map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawResponse); err != nil {
		log.Printf("Error parsing raw JSON: %v", err)
		return entities, err
	}

	// Safely get string values with type checking
	if name, ok := rawResponse["name"].(string); ok {
		entities.Name = name
	} else if name, ok := rawResponse["Name"].(string); ok { // Add check for uppercase key
		entities.Name = name
	}

	// Ensure name is not empty
	if entities.Name == "" {
		entities.Name = "Not provided"
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

	// Update parsing to handle different skill types
	if techSkills, ok := rawResponse["technical_skills"].([]interface{}); ok {
		entities.Skills = make([]string, 0, len(techSkills))
		for _, skill := range techSkills {
			if skillStr, ok := skill.(string); ok {
				entities.Skills = append(entities.Skills, skillStr)
			}
		}
	}

	// Add soft skills extraction
	if softSkills, ok := rawResponse["soft_skills"].([]interface{}); ok {
		for _, skill := range softSkills {
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

				// Clean up empty fields
				if education.Year == "" && education.GraduationDate != "" {
					education.Year = education.GraduationDate
				}

				// Only add education entry if essential fields are present
				if education.Degree != "" || education.Institution != "" {
					entities.Education = append(entities.Education, education)
				}
			}
		}
	}

	// Extract experience before projects
	if expArray, ok := rawResponse["experience"].([]interface{}); ok {
		entities.Experience = make([]Experience, 0, len(expArray))
		for _, exp := range expArray {
			if expMap, ok := exp.(map[string]interface{}); ok {
				experience := Experience{
					Title:           getString(expMap, "title"),
					Company:         getString(expMap, "company"),
					Duration:        getString(expMap, "duration"),
					Location:        getString(expMap, "location"),
					Description:     getString(expMap, "description"),
					RoleDescription: getString(expMap, "role_description"),
					Level:           getString(expMap, "level"),
				}

				// Extract responsibilities array
				if resp, ok := expMap["responsibilities"].([]interface{}); ok {
					experience.Responsibilities = make([]string, 0, len(resp))
					for _, r := range resp {
						if rStr, ok := r.(string); ok {
							experience.Responsibilities = append(experience.Responsibilities, rStr)
						}
					}
				}

				// Extract skills array
				if skills, ok := expMap["skills"].([]interface{}); ok {
					experience.Skills = make([]string, 0, len(skills))
					for _, s := range skills {
						if sStr, ok := s.(string); ok {
							experience.Skills = append(experience.Skills, sStr)
						}
					}
				}

				entities.Experience = append(entities.Experience, experience)
			}
		}
	}

	// Then extract projects separately
	if projArray, ok := rawResponse["projects"].([]interface{}); ok {
		entities.Projects = make([]Project, 0, len(projArray))
		for _, proj := range projArray {
			if projMap, ok := proj.(map[string]interface{}); ok {
				project := Project{
					Name:        getString(projMap, "name"),
					Description: getString(projMap, "description"),
					Duration:    getString(projMap, "duration"),
					Role:        getString(projMap, "role"),
					Timeline:    getString(projMap, "timeline"),
					Status:      getString(projMap, "status"),
				}

				// Extract technologies array
				if techs, ok := projMap["technologies"].([]interface{}); ok {
					project.Technologies = make([]string, 0, len(techs))
					for _, t := range techs {
						if tStr, ok := t.(string); ok {
							project.Technologies = append(project.Technologies, tStr)
						}
					}
				}

				entities.Projects = append(entities.Projects, project)
			}
		}
	}

	// Log the processed entities
	logData := map[string]any{
		"entities":       entities,
		"processed_text": text,
	}

	if err := utils.SaveJSONLog(logData, jsonStr, "resume"); err != nil {
		log.Printf("Error saving JSON log: %v", err)
	}

	// Validate before returning
	validateExtractedEntities(&entities)

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
	jobID := fmt.Sprintf("%d", time.Now().Unix())

	// Create empty entities for job description
	emptyEntities := ExtractedEntities{}

	if err := SaveProcessedText("job", processedText, jobID, emptyEntities); err != nil {
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
    2. Experience requirements (years, level, and specific areas)
    3. Educational requirements
    4. Key responsibilities and duties
    5. Preferred qualifications
    6. Project requirements or experience

    Format the output as a clean JSON object with these exact keys:
    {
        "skills": ["skill1", "skill2", ...],
        "experience": {
            "min_years": number,
            "level": "entry/mid/senior",
            "areas": ["area1", "area2", ...],
            "preferred": ["preferred exp1", "preferred exp2", ...]
        },
        "education": {
            "degree": "required degree",
            "fields": ["field1", "field2", ...],
            "qualifications": ["qualification1", ...]
        },
        "responsibilities": ["responsibility1", "responsibility2", ...],
        "project_requirements": {
            "types": ["type1", "type2", ...],
            "skills": ["skill1", "skill2", ...],
            "experience": ["exp1", "exp2", ...]
        }
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

	// Save just the cleaned JSON string
	if err := SaveCleanJSON(jsonStr, "job"); err != nil {
		log.Printf("Error saving job JSON log: %v", err)
	}

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

	// Categorize skills
	technicalSkills := FilterTechnicalSkills(requirements.Skills)
	softSkills := filterSoftSkills(requirements.Skills)

	// Create a complete job description data structure
	jobData := TextData{
		ProcessedText:   processedText,
		Timestamp:       time.Now(),
		Type:            "job",
		ID:              jobID,
		Requirements:    requirements,
		SoftSkills:      softSkills,
		TechnicalSkills: technicalSkills,
	}

	// Save the complete job data
	outputDir := filepath.Join("processed_texts", "job")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create directory",
		})
	}

	filename := fmt.Sprintf("job_%s.json", jobID)
	filePath := filepath.Join(outputDir, filename)

	jsonData, err := json.Marshal(jobData)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to marshal job data",
		})
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save job data",
		})
	}

	// Log the processed job description
	logData := map[string]any{
		"requirements":     requirements,
		"processed_text":   processedText,
		"technical_skills": technicalSkills,
		"soft_skills":      softSkills,
	}

	if err := utils.SaveJSONLog(logData, jsonStr, "job"); err != nil {
		log.Printf("Error saving JSON log: %v", err)
	}

	// After saving job file
	sessionIDJob, err := utils.SaveProcessingSession("", jobID)
	if err != nil {
		log.Printf("Error saving job session: %v", err)
	}

	return c.JSON(fiber.Map{
		"requirements":   requirements,
		"id":             jobID,
		"session_id_job": sessionIDJob,
		"filename":       filename,
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

// Add new function to validate extracted entities
func validateExtractedEntities(entities *ExtractedEntities) {
	// Ensure name is initialized
	if entities.Name == "" {
		entities.Name = "Not provided"
	}

	// Ensure skills arrays are initialized
	if entities.Skills == nil {
		entities.Skills = []string{}
	}

	// Ensure education array is initialized
	if entities.Education == nil {
		entities.Education = []Education{}
	}

	// Deduplicate skills
	skillsMap := make(map[string]bool)
	var uniqueSkills []string
	for _, skill := range entities.Skills {
		if !skillsMap[strings.ToLower(skill)] {
			skillsMap[strings.ToLower(skill)] = true
			uniqueSkills = append(uniqueSkills, skill)
		}
	}
	entities.Skills = uniqueSkills

	// Initialize other required fields if nil
	if entities.Email == nil {
		entities.Email = []string{}
	}

	if entities.Phone == "" {
		entities.Phone = "Not provided"
	}
}

// Add new function to extract education details
func extractEducationDetails(education []Education) []string {
	var details []string
	for _, edu := range education {
		detail := fmt.Sprintf("%s in %s from %s (%s)",
			edu.Degree,
			edu.Specialization,
			edu.Institution,
			edu.Year)
		details = append(details, detail)
	}
	return details
}

// Add these helper functions before extractProjects function

func splitIntoSentences(text string) []string {
	// Simple sentence splitting based on common delimiters
	delimiters := regexp.MustCompile(`[.!?]\s+`)
	sentences := delimiters.Split(text, -1)

	var cleaned []string
	for _, sentence := range sentences {
		if trimmed := strings.TrimSpace(sentence); trimmed != "" {
			cleaned = append(cleaned, trimmed)
		}
	}
	return cleaned
}

func extractProjectName(text string) string {
	// Look for project name patterns
	patterns := []string{
		`project[:|\s+]([^\.]+)`,
		`developed[\s+]([^\.]+)`,
		`created[\s+]([^\.]+)`,
		`built[\s+]([^\.]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func extractDescription(text string) string {
	// Remove common project indicators from the text
	indicators := []string{
		"project:", "developed", "created", "built",
		"implemented", "designed", "architected", "deployed",
	}

	description := text
	for _, indicator := range indicators {
		description = strings.ReplaceAll(strings.ToLower(description), indicator, "")
	}

	return strings.TrimSpace(description)
}

func extractTechnologies(text string) []string {
	// Look for technology keywords
	techKeywords := []string{
		"using", "with", "technologies:", "tech stack:",
		"built with", "developed using", "powered by",
	}

	var technologies []string
	text = strings.ToLower(text)

	for _, keyword := range techKeywords {
		if idx := strings.Index(text, keyword); idx != -1 {
			techText := text[idx+len(keyword):]
			// Split by common separators
			techs := strings.FieldsFunc(techText, func(r rune) bool {
				return r == ',' || r == ';' || r == '/' || r == '|'
			})
			for _, tech := range techs {
				if cleaned := strings.TrimSpace(tech); cleaned != "" {
					technologies = append(technologies, cleaned)
				}
			}
		}
	}

	return technologies
}

func extractDuration(text string) string {
	// Look for duration patterns
	patterns := []string{
		`(\d+)\s*(?:month|year)s?`,
		`(?:jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)[a-z]*[\s-]+\d{4}`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if match := re.FindString(text); match != "" {
			return strings.TrimSpace(match)
		}
	}
	return ""
}

func extractRole(text string) string {
	// Look for role patterns
	rolePatterns := []string{
		`role[:|\s+]([^\.]+)`,
		`position[:|\s+]([^\.]+)`,
		`as\s+a\s+([^\.]+)`,
	}

	for _, pattern := range rolePatterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func extractCompany(text string) string {
	// Look for company patterns
	patterns := []string{
		`at\s+([^,\.]+)`,
		`with\s+([^,\.]+)`,
		`for\s+([^,\.]+)`,
		`company[:|\s+]([^,\.]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func extractSkillsFromText(text string) []string {
	// Extract skills using both technical and soft skills keywords
	var skills []string

	// Combine technical and soft skills keywords
	skillsKeywords := map[string]bool{
		// Technical keywords
		"programming": true, "software": true, "development": true,
		"java": true, "python": true, "golang": true,
		"javascript": true, "react": true, "node": true,
		// Soft skills keywords
		"leadership": true, "communication": true, "teamwork": true,
		"management": true, "coordination": true, "planning": true,
	}

	words := strings.Fields(strings.ToLower(text))
	for _, word := range words {
		if skillsKeywords[word] {
			skills = append(skills, word)
		}
	}

	return skills
}

// Add new function to extract projects using spago
func extractProjects(text string) []Project {
	// Split text into sentences
	sentences := splitIntoSentences(text)

	var projects []Project
	var currentProject *Project

	for _, sentence := range sentences {
		// Look for project indicators
		if containsProjectIndicators(sentence) {
			if currentProject != nil {
				projects = append(projects, *currentProject)
			}
			currentProject = &Project{}

			// Extract project details
			currentProject.Name = extractProjectName(sentence)
			currentProject.Description = extractDescription(sentence)
			currentProject.Technologies = extractTechnologies(sentence)
			currentProject.Duration = extractDuration(sentence)
			currentProject.Role = extractRole(sentence)
			currentProject.Timeline = extractTimeline(sentence)
			currentProject.Team = extractTeamMembers(sentence)
			currentProject.Achievements = extractAchievements(sentence)
			currentProject.Status = extractProjectStatus(sentence)
		} else if currentProject != nil {
			// Append additional details to current project
			currentProject.Description += " " + sentence
		}
	}

	// Add the last project if exists
	if currentProject != nil {
		projects = append(projects, *currentProject)
	}

	return projects
}

// Update the extractExperience function
func extractExperience(text string) []Experience {
	sections := splitIntoExperienceSections(text)
	var experiences []Experience

	for _, section := range sections {
		if containsExperienceIndicators(section) {
			exp := Experience{}

			// Extract basic details
			exp.Title = extractTitle(section)
			exp.Company = extractCompany(section)
			exp.Duration = extractDuration(section)
			exp.Location = extractLocation(section)

			// Extract role description
			exp.RoleDescription = extractRoleDescription(section)
			if exp.RoleDescription == "" {
				exp.RoleDescription = "null"
			}

			// Extract responsibilities
			exp.Responsibilities = extractDetailedResponsibilities(section)
			if len(exp.Responsibilities) == 0 {
				exp.Responsibilities = []string{"null"}
			}

			// Extract other details
			exp.Skills = extractSkillsFromText(section)
			exp.Achievements = extractAchievementsFromExp(section)
			exp.TeamSize = extractTeamSize(section)
			exp.Level = extractPositionLevel(section)

			if exp.Title != "" || exp.Company != "" {
				experiences = append(experiences, exp)
			}
		}
	}

	return experiences
}

// Add new function to split text into experience sections
func splitIntoExperienceSections(text string) []string {
	// Common patterns that indicate new experience sections
	patterns := []string{
		`(?i)(?:\d{4}\s*-\s*(?:\d{4}|present))`,                                 // Date ranges
		`(?i)(?:jan|feb|mar|apr|may|jun|jul|aug|sep|oct|nov|dec)[a-z]*\s+\d{4}`, // Month Year
		`(?i)(?:work experience|professional experience|employment history)`,    // Section headers
		`(?i)(?:company|organization|firm):`,                                    // Company indicators
	}

	var sections []string
	currentSection := ""
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		isNewSection := false
		for _, pattern := range patterns {
			if matched, _ := regexp.MatchString(pattern, line); matched {
				if currentSection != "" {
					sections = append(sections, strings.TrimSpace(currentSection))
				}
				currentSection = line
				isNewSection = true
				break
			}
		}
		if !isNewSection {
			currentSection += "\n" + line
		}
	}

	if currentSection != "" {
		sections = append(sections, strings.TrimSpace(currentSection))
	}

	return sections
}

// Add new function to extract detailed role description
func extractRoleDescription(text string) string {
	// Patterns that typically introduce role descriptions
	patterns := []string{
		`(?i)role overview[:|\s+]([^\.]+\.)`,
		`(?i)position summary[:|\s+]([^\.]+\.)`,
		`(?i)job description[:|\s+]([^\.]+\.)`,
		`(?i)primary responsibilities(?:\s+include)?[:|\s+]([^\.]+\.)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}

	// If no specific pattern matches, try to extract the first relevant sentence
	sentences := splitIntoSentences(text)
	for _, sentence := range sentences {
		if containsRoleIndicators(sentence) {
			return strings.TrimSpace(sentence)
		}
	}

	return ""
}

// Add new function to extract detailed responsibilities
func extractDetailedResponsibilities(text string) []string {
	var responsibilities []string

	// Look for bullet points or numbered lists first
	bulletPoints := extractListItems(text)
	if len(bulletPoints) > 0 {
		for _, point := range bulletPoints {
			if containsResponsibilityIndicators(point) {
				responsibilities = append(responsibilities, strings.TrimSpace(point))
			}
		}
	}

	// If no bullet points found, look for responsibility-specific sentences
	if len(responsibilities) == 0 {
		sentences := splitIntoSentences(text)
		for _, sentence := range sentences {
			if containsResponsibilityIndicators(sentence) {
				responsibilities = append(responsibilities, strings.TrimSpace(sentence))
			}
		}
	}

	return responsibilities
}

// Add helper function to check for role indicators
func containsRoleIndicators(text string) bool {
	indicators := []string{
		"responsible for", "role involved", "position entailed",
		"worked as", "served as", "functioned as",
		"primary focus", "main duties", "key responsibilities",
	}

	text = strings.ToLower(text)
	for _, indicator := range indicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}
	return false
}

// Add helper function to check for responsibility indicators
func containsResponsibilityIndicators(text string) bool {
	indicators := []string{
		"developed", "implemented", "managed", "led", "created",
		"designed", "maintained", "improved", "coordinated",
		"responsible for", "handled", "oversaw", "spearheaded",
		"established", "initiated", "launched", "executed",
	}

	text = strings.ToLower(text)
	for _, indicator := range indicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}
	return false
}

// Add helper functions for extraction
func containsProjectIndicators(text string) bool {
	indicators := []string{
		"project", "developed", "created", "built", "implemented",
		"designed", "architected", "deployed", "managed project",
	}
	text = strings.ToLower(text)
	for _, indicator := range indicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}
	return false
}

func containsExperienceIndicators(text string) bool {
	indicators := []string{
		"experience", "work", "position", "role", "job",
		"employed", "worked at", "company", "organization",
	}
	text = strings.ToLower(text)
	for _, indicator := range indicators {
		if strings.Contains(text, indicator) {
			return true
		}
	}
	return false
}

// Helper functions for extraction
func splitIntoSections(text string) []string {
	// Split text into sections based on common delimiters
	delimiters := []string{"\n\n", "•", "\\*", "\\-"}
	sections := []string{text}

	for _, delimiter := range delimiters {
		var newSections []string
		for _, section := range sections {
			split := strings.Split(section, delimiter)
			for _, s := range split {
				if trimmed := strings.TrimSpace(s); trimmed != "" {
					newSections = append(newSections, trimmed)
				}
			}
		}
		sections = newSections
	}

	return sections
}

func extractTitle(text string) string {
	// Use NER to identify job titles
	// This is a simplified version
	titlePatterns := []string{
		"software engineer", "developer", "architect",
		"team lead", "manager", "consultant",
	}

	text = strings.ToLower(text)
	for _, pattern := range titlePatterns {
		if strings.Contains(text, pattern) {
			return strings.Title(pattern)
		}
	}
	return ""
}

// Add similar extraction functions for other fields...

// ...existing code...

// Add new function to extract location
func extractLocation(text string) string {
	// Look for location patterns
	patterns := []string{
		`in\s+([^,\.]+)`,
		`at\s+([^,\.]+)`,
		`location[:|\s+]([^,\.]+)`,
		`based\s+in\s+([^,\.]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

// Add new function to extract responsibilities from job description
func extractResponsibilities(text string) []string {
	sections := splitIntoSections(text)
	var responsibilities []string

	// Keywords that indicate responsibilities section
	indicators := []string{
		"responsibilities", "duties", "what you'll do",
		"what you will do", "role overview", "job duties",
	}

	for _, section := range sections {
		for _, indicator := range indicators {
			if strings.Contains(strings.ToLower(section), indicator) {
				// Extract bullet points or numbered items
				items := extractListItems(section)
				responsibilities = append(responsibilities, items...)
				break
			}
		}
	}

	return responsibilities
}

// Helper function to extract list items
func extractListItems(text string) []string {
	var items []string

	// Split by common list indicators
	patterns := []string{
		`[•\-\*]\s+([^\n]+)`, // Bullet points
		`\d+\.\s+([^\n]+)`,   // Numbered items
		`(?m)^-\s+([^\n]+)`,  // Dash items at start of line
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				items = append(items, strings.TrimSpace(match[1]))
			}
		}
	}

	return items
}

// Add helper functions for new project fields
func extractTimeline(text string) string {
	patterns := []string{
		`timeline[:|\s+]([^\.]+)`,
		`duration[:|\s+]([^\.]+)`,
		`completed in[:|\s+]([^\.]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func extractTeamMembers(text string) []string {
	patterns := []string{
		`team[:|\s+]([^\.]+)`,
		`team size[:|\s+]([^\.]+)`,
		`collaborated with[:|\s+]([^\.]+)`,
	}

	var members []string
	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			members = append(members, strings.Split(match[1], ",")...)
		}
	}
	return members
}

func extractAchievements(text string) []string {
	patterns := []string{
		`achieved[:|\s+]([^\.]+)`,
		`results[:|\s+]([^\.]+)`,
		`improved[:|\s+]([^\.]+)`,
		`implemented[:|\s+]([^\.]+)`,
	}

	var achievements []string
	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			achievements = append(achievements, strings.TrimSpace(match[1]))
		}
	}
	return achievements
}

func extractProjectStatus(text string) string {
	patterns := []string{
		`status[:|\s+]([^\.]+)`,
		`currently[:|\s+]([^\.]+)`,
		`completed[:|\s+]([^\.]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.TrimSpace(match[1])
		}
	}
	return ""
}

func extractTeamSize(text string) int {
	patterns := []string{
		`team of (\d+)`,
		`(\d+)[- ]person team`,
		`managing (\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			if size, err := strconv.Atoi(match[1]); err == nil {
				return size
			}
		}
	}
	return 0
}

func extractPositionLevel(text string) string {
	patterns := []string{
		`(junior|senior|lead|principal|staff)`,
		`level[:|\s+]([\w\s]+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		if match := re.FindStringSubmatch(text); len(match) > 1 {
			return strings.Title(strings.TrimSpace(match[1]))
		}
	}
	return ""
}

// Add the missing function
func extractAchievementsFromExp(text string) []string {
	// Look for achievement patterns
	patterns := []string{
		`achieved\s+([^\.]+)`,
		`accomplished\s+([^\.]+)`,
		`resulted in\s+([^\.]+)`,
		`successfully\s+([^\.]+)`,
	}

	var achievements []string
	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		matches := re.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				achievements = append(achievements, strings.TrimSpace(match[1]))
			}
		}
	}

	return achievements
}

// ...rest of existing code...

// Add new function to extract responsibilities from experience sections
func extractResponsibilitiesFromExp(text string) []string {
	// Look for responsibility-related patterns
	patterns := []string{
		`responsible for\s+([^\.]+)`,
		`duties included\s+([^\.]+)`,
		`key responsibilities\s*[:|\s+]([^\.]+)`,
	}

	var responsibilities []string

	// First try to find explicitly stated responsibilities
	for _, pattern := range patterns {
		re := regexp.MustCompile(`(?i)` + pattern)
		matches := re.FindAllStringSubmatch(text, -1)
		for _, match := range matches {
			if len(match) > 1 {
				responsibilities = append(responsibilities, strings.TrimSpace(match[1]))
			}
		}
	}

	// If no explicit responsibilities found, look for bullet points or numbered lists
	if len(responsibilities) == 0 {
		responsibilities = extractListItems(text)
	}

	// If still no responsibilities found, try to extract action-oriented sentences
	if len(responsibilities) == 0 {
		actionWords := []string{
			`(developed|implemented|managed|led|created|designed|maintained|improved|coordinated)`,
		}
		for _, pattern := range actionWords {
			re := regexp.MustCompile(`(?i)` + pattern + `\s+([^\.]+)`)
			matches := re.FindAllStringSubmatch(text, -1)
			for _, match := range matches {
				if len(match) > 1 {
					responsibilities = append(responsibilities, strings.TrimSpace(match[0]))
				}
			}
		}
	}

	return responsibilities
}

// ...rest of existing code...

// PreprocessJobRequest represents the request body for job preprocessing
type PreprocessJobRequest struct {
	Description string `json:"description"`
	JobID       string `json:"job_id"`
}

// PreprocessJob handles job description preprocessing
func PreprocessJob(c *fiber.Ctx) error {
	var req PreprocessJobRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot parse JSON",
		})
	}

	// Validate JobID format
	matched, err := regexp.MatchString(`^job_\d+$`, req.JobID)
	if err != nil || !matched {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid JobID format",
		})
	}

	// Process the job description
	// ...existing processing code...

	// Example: Store the job data with JobID
	jobData := fiber.Map{
		"job_id":      req.JobID,
		"description": req.Description,
		// ...other processed fields...
	}

	// TODO: Implement storage logic (e.g., save to database)

	return c.JSON(jobData)
}

// Example fix for an incomplete function
// ...existing code...

func extractEmail(raw interface{}) string {
	if emails, ok := raw.([]interface{}); ok {
		if len(emails) > 0 {
			if emailStr, ok := emails[0].(string); ok {
				return emailStr
			}
		}
	}
	return "" // Add empty string return value
}
