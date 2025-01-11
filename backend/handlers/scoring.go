package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2" // For semantic search
	"github.com/jdkato/prose/v2"  // For NER
	// For matrix operations
	// For BERT tokenization
)

// ScoreResponse represents the scoring results
type ScoreResponse struct {
	OverallScore       float64            `json:"overall_score"`
	SkillsMatch        float64            `json:"skills_match"`
	ExperienceMatch    float64            `json:"experience_match"`
	EducationMatch     float64            `json:"education_match"`
	DetailedScores     map[string]float64 `json:"detailed_scores"`
	Feedback           []string           `json:"feedback"`
	MatchedSkills      SkillMatches       `json:"matched_skills"`
	SoftSkillsAnalysis SoftSkillsData     `json:"soft_skills_analysis"`
}

// Add new type for skill matches
type SkillMatches struct {
	ExactMatches   []string       `json:"exact_matches"`
	PartialMatches []PartialMatch `json:"partial_matches"`
	MissingSkills  []string       `json:"missing_skills"`
}

type PartialMatch struct {
	JobSkill    string  `json:"job_skill"`
	ResumeSkill string  `json:"resume_skill"`
	Similarity  float64 `json:"similarity"`
}

// Add new type for soft skills data
type SoftSkillsData struct {
	Score            float64  `json:"score"`
	ExtractedSkills  []string `json:"extracted_skills"`
	ExperienceSkills []string `json:"experience_based_skills"`
}

// Remove TextData struct as it's now in models.go

// Add new scoring constants
const (
	similarityThreshold = 0.75
	contextWeight       = 0.3
	skillWeight         = 0.4
	experienceWeight    = 0.3
	maxScore            = 100.0
	semanticWeight      = 0.4
	keywordWeight       = 0.3
	entityWeight        = 0.3
	wordSimThreshold    = 0.7
)

// Add semantic search model struct
type SemanticModel struct {
	vocabulary map[string]int
	docVectors [][]float64
}

// Initialize semantic search model
func initSemanticModel() *SemanticModel {
	return &SemanticModel{
		vocabulary: make(map[string]int),
		docVectors: make([][]float64, 0),
	}
}

// Add safety checks for score calculations
func safeFloat64(value float64) float64 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0.0
	}
	return value
}

// Add new helper functions at the beginning of the file
func getLatestFileID(textType string) (string, error) {
	dir := filepath.Join("processed_texts", textType)
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no %s files found", textType)
	}

	// Sort files by modification time
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime().After(files[j].ModTime())
	})

	// Extract ID from filename (filename format: "type_id.json")
	latestFile := files[0].Name()
	fileID := strings.TrimPrefix(latestFile, textType+"_")
	fileID = strings.TrimSuffix(fileID, ".json")

	return fileID, nil
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

	// If IDs are not provided, use the most recent files
	var err error
	if request.ResumeID == "" {
		request.ResumeID, err = getLatestFileID("resume")
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "No resume files found",
			})
		}
	}

	if request.JobID == "" {
		request.JobID, err = getLatestFileID("job")
		if err != nil {
			return c.Status(404).JSON(fiber.Map{
				"error": "No job files found",
			})
		}
	}

	// Log the IDs being used
	log.Printf("Scoring Resume - Using Resume ID: %s, Job ID: %s", request.ResumeID, request.JobID)

	// Log the request
	log.Printf("Scoring Request - ResumeID: %s, JobID: %s", request.ResumeID, request.JobID)

	// Load both text and entities
	resumeData, err := loadTextData(request.ResumeID, "resume")
	if err != nil {
		log.Printf("Error loading resume data: %v", err)
		return c.Status(404).JSON(fiber.Map{
			"error": "Resume data not found",
		})
	}
	log.Printf("Loaded Resume Data - Skills: %v, Education: %v",
		resumeData.Entities.Skills,
		resumeData.Entities.Education)

	jobData, err := loadTextData(request.JobID, "job")
	if err != nil {
		log.Printf("Error loading job data: %v", err)
		return c.Status(404).JSON(fiber.Map{
			"error": "Job description data not found",
		})
	}
	log.Printf("Loaded Job Data - Skills: %v, Requirements: %v",
		jobData.Requirements.Skills,
		jobData.Requirements)

	// Calculate normalized scores (0-100 scale)
	skillsScore := math.Min(safeFloat64(calculateSkillsMatch(resumeData.Entities.Skills, jobData.Requirements.Skills)*maxScore), maxScore)
	experienceScore := math.Min(safeFloat64(calculateExperienceMatch(resumeData.Entities, jobData.Requirements)*maxScore), maxScore)
	educationScore := math.Min(safeFloat64(calculateEducationMatch(resumeData.Entities.Education, jobData.Requirements.Education)*maxScore), maxScore)
	technicalScore := math.Min(safeFloat64(calculateTechnicalSkillsScore(resumeData, jobData)*maxScore), maxScore)

	// Calculate overall score using technical skills instead of education (weighted average)
	overallScore := math.Min(safeFloat64(
		skillsScore*0.4+
			experienceScore*0.3+
			technicalScore*0.3), maxScore)

	// Generate feedback based on scores
	feedback := generateFeedback(skillsScore, experienceScore, educationScore, resumeData, jobData)

	softSkillsScore, softSkillsData := calculateSoftSkillsScore(resumeData, jobData)

	scoreResponse := ScoreResponse{
		OverallScore:    overallScore,
		SkillsMatch:     skillsScore,
		ExperienceMatch: experienceScore,
		EducationMatch:  educationScore,
		DetailedScores: map[string]float64{
			"technical_skills": technicalScore,
			"soft_skills":      math.Min(safeFloat64(softSkillsScore*maxScore), maxScore),
			"qualifications":   educationScore,
		},
		Feedback:           feedback,
		SoftSkillsAnalysis: softSkillsData,
	}

	// Calculate skill matches
	exactMatches, partialMatches, missingSkills := analyzeSkillMatches(
		resumeData.Entities.Skills,
		jobData.Requirements.Skills,
	)

	scoreResponse.MatchedSkills = SkillMatches{
		ExactMatches:   exactMatches,
		PartialMatches: partialMatches,
		MissingSkills:  missingSkills,
	}

	// Log the final score response
	log.Printf("Score Response: %+v", scoreResponse)

	return c.JSON(scoreResponse)
}

func loadTextData(id string, textType string) (*TextData, error) {
	filePath := filepath.Join("processed_texts", textType, fmt.Sprintf("%s_%s.json", textType, id))

	// Add debug logging
	log.Printf("Attempting to load file: %s", filePath)

	fileData, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file: %v", err)
		return nil, err
	}

	var data TextData
	if err := json.Unmarshal(fileData, &data); err != nil {
		return nil, err
	}

	return &data, nil
}

// Enhanced calculateSkillsMatch with semantic search
func calculateSkillsMatch(resumeSkills []string, jobSkills []string) float64 {
	// Add logging
	log.Printf("Calculating skills match - Resume Skills: %v", resumeSkills)
	log.Printf("Calculating skills match - Job Skills: %v", jobSkills)

	if len(jobSkills) == 0 {
		return 1.0
	}

	// Initialize models
	semanticModel := initSemanticModel()
	nerModel, _ := prose.NewDocument("")

	// Calculate different similarity scores
	semanticScore := calculateSemanticSimilarity(resumeSkills, jobSkills, semanticModel)
	keywordScore := calculateKeywordMatch(resumeSkills, jobSkills)
	entityScore := calculateEntityMatch(resumeSkills, jobSkills, nerModel)

	// Add score logging
	log.Printf("Skills match scores - Semantic: %.2f, Keyword: %.2f, Entity: %.2f",
		semanticScore, keywordScore, entityScore)

	// Weighted combination (removing maxScore multiplication as it's applied later)
	return semanticScore*semanticWeight + keywordScore*keywordWeight + entityScore*entityWeight
}

// Calculate semantic similarity using TF-IDF and cosine similarity
func calculateSemanticSimilarity(resumeSkills, jobSkills []string, model *SemanticModel) float64 {
	// Build vocabulary
	allTerms := make(map[string]bool)
	for _, skill := range append(resumeSkills, jobSkills...) {
		words := strings.Fields(strings.ToLower(skill))
		for _, word := range words {
			allTerms[word] = true
		}
	}

	// Create term-document matrix
	vocabulary := make([]string, 0, len(allTerms))
	for term := range allTerms {
		vocabulary = append(vocabulary, term)
	}
	sort.Strings(vocabulary)

	// Calculate TF-IDF vectors
	resumeVector := calculateTfIdf(resumeSkills, vocabulary)
	jobVector := calculateTfIdf(jobSkills, vocabulary)

	// Calculate cosine similarity
	return cosineSimilarityVec(resumeVector, jobVector)
}

// Calculate TF-IDF vector for a document
func calculateTfIdf(doc []string, vocabulary []string) []float64 {
	// Calculate term frequencies
	tf := make(map[string]float64)
	for _, skill := range doc {
		words := strings.Fields(strings.ToLower(skill))
		for _, word := range words {
			tf[word]++
		}
	}

	// Create TF-IDF vector
	vector := make([]float64, len(vocabulary))
	for i, term := range vocabulary {
		if freq, ok := tf[term]; ok {
			vector[i] = freq * math.Log(2.0) // Simple IDF approximation
		}
	}

	return vector
}

// Update cosineSimilarityVec to prevent NaN
func cosineSimilarityVec(v1, v2 []float64) float64 {
	if len(v1) != len(v2) || len(v1) == 0 {
		return 0.0
	}

	var dotProduct, norm1, norm2 float64
	for i := range v1 {
		dotProduct += v1[i] * v2[i]
		norm1 += v1[i] * v1[i]
		norm2 += v2[i] * v2[i]
	}

	norm1 = math.Sqrt(norm1)
	norm2 = math.Sqrt(norm2)

	if norm1 <= 0 || norm2 <= 0 {
		return 0.0
	}

	similarity := dotProduct / (norm1 * norm2)
	return safeFloat64(similarity)
}

func calculateExperienceMatch(resumeEntities ExtractedEntities, jobReqs JobRequirements) float64 {
	// Extract experience-related sentences from resume
	resumeExp := extractExperienceStatements(resumeEntities)

	// Get required experience areas and years
	requiredExp := jobReqs.Experience

	// Calculate years match
	yearsScore := calculateYearsMatch(resumeExp, requiredExp.MinYears)

	// Calculate area match using BERT similarity
	areaScore := calculateAreaMatch(resumeExp, requiredExp.Areas)

	// Calculate level match
	levelScore := calculateLevelMatch(resumeExp, requiredExp.Level)

	// Add semantic analysis of experience descriptions
	semanticScore := calculateExperienceSemanticMatch(
		resumeEntities.Skills,
		jobReqs.Experience.Areas,
	)

	// Weighted combination of scores
	return (yearsScore*0.3 + areaScore*0.3 + levelScore*0.2 + semanticScore*0.2)
}

// New function for semantic matching of experience
func calculateExperienceSemanticMatch(resumeExp []string, requiredAreas []string) float64 {
	if len(requiredAreas) == 0 {
		return 1.0
	}

	model := initSemanticModel()
	return calculateSemanticSimilarity(resumeExp, requiredAreas, model)
}

func calculateEducationMatch(resumeEducation []Education, jobEducation struct {
	Degree         string   `json:"degree"`
	Fields         []string `json:"fields"`
	Qualifications []string `json:"qualifications"`
}) float64 {
	if len(resumeEducation) == 0 {
		return 0.0
	}

	var scores []float64
	for _, edu := range resumeEducation {
		// Calculate degree match using BERT similarity
		degreeScore := calculateBertSimilarity(edu.Degree, jobEducation.Degree)

		// Calculate field match
		fieldScore := calculateFieldMatch(edu.Specialization, jobEducation.Fields)

		// Calculate qualifications match
		qualScore := calculateQualificationsMatch(edu, jobEducation.Qualifications)

		// Combine scores
		totalScore := (degreeScore*0.4 + fieldScore*0.4 + qualScore*0.2)
		scores = append(scores, totalScore)
	}

	// Return highest education match
	sort.Float64s(scores)
	if len(scores) > 0 {
		return scores[len(scores)-1]
	}
	return 0.0
}

func generateFeedback(skillsScore, experienceScore, educationScore float64, resumeData *TextData, jobData *TextData) []string {
	var feedback []string

	// Generate specific skill gap feedback
	skillGaps := identifySkillGaps(resumeData.Entities.Skills, jobData.Requirements.Skills)
	if len(skillGaps) > 0 {
		feedback = append(feedback, fmt.Sprintf("Consider developing these skills: %s", strings.Join(skillGaps, ", ")))
	}

	// Experience feedback - changed threshold to 70 on 100 scale
	if experienceScore < 70 {
		gaps := identifyExperienceGaps(resumeData.Entities, jobData.Requirements)
		feedback = append(feedback, gaps...)
	}

	// Education feedback - changed threshold to 70 on 100 scale
	if educationScore < 70 {
		eduFeedback := generateEducationFeedback(resumeData.Entities.Education, jobData.Requirements.Education)
		feedback = append(feedback, eduFeedback...)
	}

	return feedback
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

func calculateTechnicalSkillsScore(resumeData *TextData, jobData *TextData) float64 {
	techSkills := FilterTechnicalSkills(resumeData.Entities.Skills)
	requiredTechSkills := FilterTechnicalSkills(jobData.Requirements.Skills)

	// Ensure score is between 0 and 1 before maxScore multiplication
	return math.Min(calculateSkillsMatch(techSkills, requiredTechSkills), 1.0)
}

// Modify calculateSoftSkillsScore to return both score and skills
func calculateSoftSkillsScore(resumeData *TextData, jobData *TextData) (float64, SoftSkillsData) {
	// Extract soft skills from both resume and job description
	softSkills := filterSoftSkills(resumeData.Entities.Skills)
	requiredSoftSkills := filterSoftSkills(jobData.Requirements.Skills)

	// Log extracted soft skills
	log.Printf("Extracted soft skills from resume: %v", softSkills)
	log.Printf("Required soft skills from job: %v", requiredSoftSkills)

	// Extract experience-based soft skills
	expSkills, expScore := extractSoftSkillsFromExperience(resumeData.Entities)
	log.Printf("Experience-based soft skills: %v", expSkills)

	if len(requiredSoftSkills) == 0 {
		return 1.0, SoftSkillsData{
			Score:            100.0,
			ExtractedSkills:  softSkills,
			ExperienceSkills: expSkills,
		}
	}

	// Calculate scores
	semanticScore := calculateSemanticSimilarity(softSkills, requiredSoftSkills, initSemanticModel())
	keywordScore := calculateKeywordMatch(softSkills, requiredSoftSkills)

	// Weighted combination of scores
	weightedScore := (semanticScore * 0.4) + (keywordScore * 0.4) + (expScore * 0.2)
	finalScore := math.Min(weightedScore, 1.0)

	return finalScore, SoftSkillsData{
		Score:            finalScore * 100,
		ExtractedSkills:  softSkills,
		ExperienceSkills: expSkills,
	}
}

// Update extractSoftSkillsFromExperience to return both score and skills
func extractSoftSkillsFromExperience(entities ExtractedEntities) ([]string, float64) {
	softSkillIndicators := map[string]bool{
		"led": true, "managed": true, "coordinated": true,
		"collaborated": true, "mentored": true, "trained": true,
		"facilitated": true, "organized": true, "presented": true,
		"negotiated": true, "resolved": true, "improved": true,
	}

	var extractedSkills []string
	score := 0.0
	totalIndicators := float64(len(softSkillIndicators))

	for _, skill := range entities.Skills {
		skillLower := strings.ToLower(skill)
		for indicator := range softSkillIndicators {
			if strings.Contains(skillLower, indicator) {
				score++
				extractedSkills = append(extractedSkills, skill)
				break
			}
		}
	}

	return extractedSkills, score / totalIndicators
}

func filterSoftSkills(skills []string) []string {
	// Expanded list of soft skills keywords
	softKeywords := map[string]bool{
		"communication": true, "leadership": true, "teamwork": true,
		"problem solving": true, "analytical": true, "creative": true,
		"interpersonal": true, "organization": true, "time management": true,
		"adaptability": true, "collaboration": true, "management": true,
		"critical thinking": true, "emotional intelligence": true,
		"conflict resolution": true, "negotiation": true, "presentation": true,
		"decision making": true, "flexibility": true, "multitasking": true,
		"self-motivated": true, "work ethic": true, "attention to detail": true,
		"team player": true, "project management": true, "mentoring": true,
		"strategic thinking": true, "coaching": true, "public speaking": true,
	}

	var soft []string
	for _, skill := range skills {
		skill = strings.ToLower(skill)
		for keyword := range softKeywords {
			if strings.Contains(skill, keyword) {
				soft = append(soft, skill)
				break
			}
		}
	}
	return soft
}

// Helper functions for BERT-based similarity
func getBertEmbeddings(texts []string) [][]float32 {
	// For now, return dummy embeddings
	embeddings := make([][]float32, len(texts))
	for i := range texts {
		embeddings[i] = make([]float32, 384) // Standard BERT embedding size
	}
	return embeddings
}

// Update cosineSimilarity to prevent NaN
func cosineSimilarity(a, b []float32) float64 {
	if len(a) != len(b) || len(a) == 0 {
		return 0.0
	}

	var dotProduct float32
	var normA, normB float32

	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	if normA <= 0 || normB <= 0 {
		return 0.0
	}

	return safeFloat64(float64(dotProduct / (sqrt(normA) * sqrt(normB))))
}

func identifySkillGaps(resumeSkills []string, jobSkills []string) []string {
	var gaps []string
	resumeEmbeddings := getBertEmbeddings(resumeSkills)
	jobEmbeddings := getBertEmbeddings(jobSkills)

	for i, jobSkill := range jobSkills {
		matched := false
		for _, resumeEmb := range resumeEmbeddings {
			if cosineSimilarity(jobEmbeddings[i], resumeEmb) > similarityThreshold {
				matched = true
				break
			}
		}
		if !matched {
			gaps = append(gaps, jobSkill)
		}
	}
	return gaps
}

func identifyExperienceGaps(resumeEntities ExtractedEntities, jobReqs JobRequirements) []string {
	var gaps []string

	// Check experience years
	if yearsOfExperience := extractYearsOfExperience(resumeEntities); yearsOfExperience < float64(jobReqs.Experience.MinYears) {
		gaps = append(gaps, fmt.Sprintf("Need %d more years of experience", jobReqs.Experience.MinYears-int(yearsOfExperience)))
	}

	// Check experience areas
	missingAreas := findMissingExperienceAreas(resumeEntities, jobReqs.Experience.Areas)
	if len(missingAreas) > 0 {
		gaps = append(gaps, fmt.Sprintf("Need experience in: %s", strings.Join(missingAreas, ", ")))
	}

	return gaps
}

// Helper function to calculate weighted average
func calculateWeightedAverage(scores []float64) float64 {
	if len(scores) == 0 {
		return 0.0
	}

	sum := 0.0
	for _, score := range scores {
		sum += score
	}
	return sum / float64(len(scores))
}

// Extract experience statements from entities
func extractExperienceStatements(entities ExtractedEntities) []string {
	// This would normally parse work experience sections
	// For now, return empty slice
	return []string{}
}

// Calculate years match between resume experience and job requirements
func calculateYearsMatch(resumeExp []string, requiredYears int) float64 {
	// Create a dummy ExtractedEntities with the experience information
	entities := ExtractedEntities{
		Skills: resumeExp, // Use the experience strings as skills
	}

	years := extractYearsOfExperience(entities)
	if years >= float64(requiredYears) {
		return 1.0
	}
	return years / float64(requiredYears)
}

// Calculate area match between resume experience and required areas
func calculateAreaMatch(resumeExp []string, requiredAreas []string) float64 {
	if len(requiredAreas) == 0 {
		return 1.0
	}

	matches := 0
	for _, area := range requiredAreas {
		for _, exp := range resumeExp {
			if strings.Contains(strings.ToLower(exp), strings.ToLower(area)) {
				matches++
				break
			}
		}
	}
	return float64(matches) / float64(len(requiredAreas))
}

// Calculate level match between resume experience and required level
func calculateLevelMatch(resumeExp []string, requiredLevel string) float64 {
	levelScores := map[string]float64{
		"entry":  0.33,
		"mid":    0.66,
		"senior": 1.0,
	}

	resumeLevel := determineExperienceLevel(resumeExp)
	if score, ok := levelScores[resumeLevel]; ok {
		requiredScore := levelScores[requiredLevel]
		if score >= requiredScore {
			return 1.0
		}
		return score / requiredScore
	}
	return 0.0
}

// Calculate similarity between two text strings using BERT
func calculateBertSimilarity(text1, text2 string) float64 {
	// This would use BERT embeddings for similarity
	// For now, use simple string matching
	text1 = strings.ToLower(text1)
	text2 = strings.ToLower(text2)
	if strings.Contains(text1, text2) || strings.Contains(text2, text1) {
		return 1.0
	}
	return 0.0
}

// Calculate field match for education
func calculateFieldMatch(resumeField string, requiredFields []string) float64 {
	if len(requiredFields) == 0 {
		return 1.0
	}

	maxScore := 0.0
	for _, field := range requiredFields {
		score := calculateBertSimilarity(resumeField, field)
		if score > maxScore {
			maxScore = score
		}
	}
	return maxScore
}

// Calculate qualifications match
func calculateQualificationsMatch(edu Education, requiredQuals []string) float64 {
	if len(requiredQuals) == 0 {
		return 1.0
	}

	matches := 0
	for _, qual := range requiredQuals {
		if strings.Contains(strings.ToLower(edu.Degree), strings.ToLower(qual)) {
			matches++
		}
	}
	return float64(matches) / float64(len(requiredQuals))
}

// Generate education-related feedback
func generateEducationFeedback(resumeEdu []Education, jobEdu struct {
	Degree         string   `json:"degree"`
	Fields         []string `json:"fields"`
	Qualifications []string `json:"qualifications"`
}) []string {
	var feedback []string

	if len(resumeEdu) == 0 {
		feedback = append(feedback, "No education information found in resume")
		return feedback
	}

	// Check degree match
	degreeMatch := false
	for _, edu := range resumeEdu {
		if calculateBertSimilarity(edu.Degree, jobEdu.Degree) > 0.8 {
			degreeMatch = true
			break
		}
	}
	if !degreeMatch {
		feedback = append(feedback, fmt.Sprintf("Consider pursuing %s degree", jobEdu.Degree))
	}

	return feedback
}

// Load BERT model for embeddings
func loadBertModel() interface{} {
	// This would load the actual BERT model
	// For now, return a dummy implementation
	return struct{}{}
}

// Helper function to determine experience level
func determineExperienceLevel(experience []string) string {
	// Simple implementation - would be more sophisticated in practice
	for _, exp := range experience {
		if strings.Contains(strings.ToLower(exp), "senior") {
			return "senior"
		}
		if strings.Contains(strings.ToLower(exp), "lead") {
			return "senior"
		}
	}
	return "entry"
}

// Extract years of experience from resume
func extractYearsOfExperience(entities ExtractedEntities) float64 {
	var totalYears float64

	// Extract years from work experience if available
	// This is a simplified implementation
	for _, skill := range entities.Skills {
		if strings.Contains(strings.ToLower(skill), "year") {
			// Try to parse number before "year"
			parts := strings.Fields(skill)
			for i, part := range parts {
				if strings.Contains(part, "year") && i > 0 {
					if years, err := strconv.ParseFloat(parts[i-1], 64); err == nil {
						totalYears += years
					}
				}
			}
		}
	}

	return totalYears
}

// Add findMissingExperienceAreas function
func findMissingExperienceAreas(entities ExtractedEntities, required []string) []string {
	var missing []string
	for _, req := range required {
		found := false
		for _, skill := range entities.Skills {
			if calculateBertSimilarity(strings.ToLower(skill), strings.ToLower(req)) > 0.8 {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, req)
		}
	}
	return missing
}

// Helper function to find years mentioned in text
func findYearsInText(text string) float64 {
	// Simple implementation - would use regex in practice
	if strings.Contains(text, "year") {
		return 1.0
	}
	return 0.0
}

// Helper function for square root
func sqrt(x float32) float32 {
	return float32(math.Sqrt(float64(x)))
}

// Helper function for word similarity
func wordSimilarity(word1, word2 string) float64 {
	// Implement word2vec or WordNet similarity here
	// For now, using simple Levenshtein distance
	distance := levenshteinDistance(word1, word2)
	maxLen := float64(max(len(word1), len(word2)))
	return 1.0 - float64(distance)/maxLen
}

// Helper function for Levenshtein distance
func levenshteinDistance(s1, s2 string) int {
	// ...implement Levenshtein distance algorithm...
	return 0 // Placeholder
}

// Helper function for entity comparison
func compareEntities(e1, e2 prose.Entity) bool {
	return e1.Label == e2.Label && wordSimilarity(e1.Text, e2.Text) > wordSimThreshold
}

// Helper function for extracting entities
func extractEntities(skills []string, doc *prose.Document) []prose.Entity {
	var entities []prose.Entity
	for _, skill := range skills {
		doc, _ = prose.NewDocument(skill)
		entities = append(entities, doc.Entities()...)
	}
	return entities
}

// Helper math function
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Move these functions before they are used in calculateSkillsMatch
// Enhanced keyword matching with word similarity
func calculateKeywordMatch(resumeSkills, jobSkills []string) float64 {
	if len(jobSkills) == 0 {
		return 1.0
	}
	matches := 0.0
	for _, jobSkill := range jobSkills {
		maxSim := 0.0
		for _, resumeSkill := range resumeSkills {
			sim := wordSimilarity(jobSkill, resumeSkill)
			if sim > maxSim {
				maxSim = sim
			}
		}
		if maxSim > wordSimThreshold {
			matches++
		}
	}
	return matches / float64(len(jobSkills))
}

// Simplified entity matching focusing on meaningful comparison
func calculateEntityMatch(resumeSkills, jobSkills []string, doc *prose.Document) float64 {
	if len(jobSkills) == 0 {
		return 1.0
	}

	// Extract meaningful terms from skills
	resumeTerms := extractMeaningfulTerms(resumeSkills)
	jobTerms := extractMeaningfulTerms(jobSkills)

	// Calculate direct matches
	directMatches := calculateDirectMatches(resumeTerms, jobTerms)

	// Calculate fuzzy matches for non-exact matches
	fuzzyMatches := calculateFuzzyMatches(resumeTerms, jobTerms)

	// Combine scores with weights
	score := (directMatches * 0.7) + (fuzzyMatches * 0.3)
	return safeFloat64(score)
}

func extractMeaningfulTerms(skills []string) []string {
	var terms []string
	for _, skill := range skills {
		// Split compound terms
		words := strings.Fields(strings.ToLower(skill))
		// Filter out common stop words
		for _, word := range words {
			if !isStopWord(word) && len(word) > 2 {
				terms = append(terms, word)
			}
		}
	}
	return terms
}

func calculateDirectMatches(terms1, terms2 []string) float64 {
	if len(terms2) == 0 {
		return 1.0
	}

	matches := 0.0
	for _, term2 := range terms2 {
		for _, term1 := range terms1 {
			if term1 == term2 {
				matches++
				break
			}
		}
	}
	return matches / float64(len(terms2))
}

func calculateFuzzyMatches(terms1, terms2 []string) float64 {
	if len(terms2) == 0 {
		return 1.0
	}

	matches := 0.0
	for _, term2 := range terms2 {
		maxSimilarity := 0.0
		for _, term1 := range terms1 {
			similarity := wordSimilarity(term1, term2)
			if similarity > maxSimilarity {
				maxSimilarity = similarity
			}
		}
		if maxSimilarity > wordSimThreshold {
			matches += maxSimilarity
		}
	}
	return matches / float64(len(terms2))
}

func isStopWord(word string) bool {
	stopWords := map[string]bool{
		"the": true, "and": true, "or": true, "in": true, "on": true,
		"at": true, "to": true, "for": true, "with": true, "by": true,
	}
	return stopWords[word]
}

// Add new helper function for analyzing skill matches
func analyzeSkillMatches(resumeSkills, jobSkills []string) ([]string, []PartialMatch, []string) {
	var exactMatches []string
	var partialMatches []PartialMatch
	var missingSkills []string

	// Convert all skills to lowercase for comparison
	resumeSkillsLower := make(map[string]string)
	for _, skill := range resumeSkills {
		resumeSkillsLower[strings.ToLower(skill)] = skill
	}

	// Check each job skill
	for _, jobSkill := range jobSkills {
		jobSkillLower := strings.ToLower(jobSkill)

		// Check for exact match
		if _, exists := resumeSkillsLower[jobSkillLower]; exists {
			exactMatches = append(exactMatches, jobSkill)
			continue
		}

		// Check for partial matches
		bestMatch := ""
		bestSimilarity := 0.0
		for resumeSkillLower, originalResumeSkill := range resumeSkillsLower {
			similarity := wordSimilarity(jobSkillLower, resumeSkillLower)
			if similarity > wordSimThreshold && similarity > bestSimilarity {
				bestMatch = originalResumeSkill
				bestSimilarity = similarity
			}
		}

		if bestMatch != "" {
			partialMatches = append(partialMatches, PartialMatch{
				JobSkill:    jobSkill,
				ResumeSkill: bestMatch,
				Similarity:  bestSimilarity,
			})
		} else {
			missingSkills = append(missingSkills, jobSkill)
		}
	}

	return exactMatches, partialMatches, missingSkills
}

// Update the filterTechnicalSkills function to be exported (capitalize first letter)
func FilterTechnicalSkills(skills []string) []string {
	technicalKeywords := map[string]bool{
		"programming": true, "software": true, "development": true,
		"java": true, "python": true, "go": true, "golang": true,
		"javascript": true, "react": true, "node": true, "aws": true,
		"cloud": true, "docker": true, "kubernetes": true, "git": true,
		"database": true, "sql": true, "nosql": true, "api": true,
		// Add more technical keywords as needed
	}

	var technical []string
	for _, skill := range skills {
		skill = strings.ToLower(skill)
		for keyword := range technicalKeywords {
			if strings.Contains(skill, keyword) {
				technical = append(technical, skill)
				break
			}
		}
	}
	return technical
}

// Update function calls to use the exported version
