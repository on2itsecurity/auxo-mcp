package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/on2itsecurity/auxo-mcp/server/client"
	"github.com/on2itsecurity/auxo-mcp/server/types"
	"github.com/on2itsecurity/go-auxo/v2/zerotrust"
)

// ReadinessTools contains all Zero Trust readiness assessment related tools
type ReadinessTools struct {
	clientManager *client.Manager
}

// NewReadinessTools creates a new readiness tools instance
func NewReadinessTools(clientManager *client.Manager) *ReadinessTools {
	return &ReadinessTools{clientManager: clientManager}
}

// GetQuestions returns the readiness assessment questionnaire
func (t *ReadinessTools) GetQuestions(ctx context.Context, req *mcp.CallToolRequest, args types.EmptyParams) (*mcp.CallToolResult, any, error) {
	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	questions, err := auxoClient.ZeroTrust.GetReadinessQuestions(ctx)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	jsonData, err := json.Marshal(questions)
	if err != nil {
		return nil, nil, err
	}

	content := []mcp.Content{
		&mcp.TextContent{
			Text: string(jsonData),
		}}

	return &mcp.CallToolResult{
		Content: content,
	}, nil, nil
}

// Start opens the interactive readiness assessment. In hosts that support MCP
// Apps the linked ui:// resource is rendered and reads the questionnaire from
// StructuredContent; everywhere else the text content instructs the model to
// run the assessment conversationally.
func (t *ReadinessTools) Start(ctx context.Context, req *mcp.CallToolRequest, args types.EmptyParams) (*mcp.CallToolResult, any, error) {
	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	questions, err := auxoClient.ZeroTrust.GetReadinessQuestions(ctx)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	total := len(questions.Strategical) + len(questions.Tactical) + len(questions.Operational)
	fallback := fmt.Sprintf(
		"Readiness assessment started (%d questions: %d strategical, %d tactical, %d operational, plus scoping; question set version %d). "+
			"If this client rendered an interactive questionnaire panel, the user completes and submits the assessment there - wait for the result and do not interview them in parallel. "+
			"If no panel appeared, run the assessment conversationally: interview the user using the questions in structuredContent (each answered 1-5 CMMI with an actual and a goal level) and submit with createReadinessAssessment.",
		total, len(questions.Strategical), len(questions.Tactical), len(questions.Operational), questions.Version)

	content := []mcp.Content{
		&mcp.TextContent{
			Text: fallback,
		}}

	return &mcp.CallToolResult{
		Content:           content,
		StructuredContent: questions,
	}, nil, nil
}

// List returns a lightweight summary of all readiness assessments
func (t *ReadinessTools) List(ctx context.Context, req *mcp.CallToolRequest, args types.EmptyParams) (*mcp.CallToolResult, any, error) {
	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	summaries, err := auxoClient.ZeroTrust.GetReadinessAssessmentsSummary(ctx)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	// Enrich the unix timestamps with a readable date
	enriched := make([]map[string]interface{}, 0, len(summaries))
	for _, s := range summaries {
		enriched = append(enriched, map[string]interface{}{
			"id":                   s.ID,
			"assessment_timestamp": s.AssessmentTimestamp,
			"assessment_date":      time.Unix(s.AssessmentTimestamp, 0).UTC().Format(time.RFC3339),
		})
	}

	jsonData, err := json.Marshal(enriched)
	if err != nil {
		return nil, nil, err
	}

	content := []mcp.Content{
		&mcp.TextContent{
			Text: string(jsonData),
		}}

	return &mcp.CallToolResult{
		Content: content,
	}, nil, nil
}

// Get returns the full details of a readiness assessment by its ID
func (t *ReadinessTools) Get(ctx context.Context, req *mcp.CallToolRequest, args types.ReadinessAssessmentIDParams) (*mcp.CallToolResult, any, error) {
	if args.ID == "" {
		return nil, nil, fmt.Errorf("id is required to get a readiness assessment")
	}

	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	assessment, err := auxoClient.ZeroTrust.GetReadinessAssessmentByID(ctx, args.ID)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	jsonData, err := json.Marshal(assessment)
	if err != nil {
		return nil, nil, err
	}

	content := []mcp.Content{
		&mcp.TextContent{
			Text: string(jsonData),
		}}

	return &mcp.CallToolResult{
		Content: content,
	}, nil, nil
}

// Create validates and submits a completed readiness assessment
func (t *ReadinessTools) Create(ctx context.Context, req *mcp.CallToolRequest, args types.CreateReadinessAssessmentParams) (*mcp.CallToolResult, any, error) {
	// Validate required fields for creation
	if args.AnsweredBy == "" {
		return nil, nil, fmt.Errorf("answered_by is required for creating a readiness assessment (the e-mail of the user who answered the questions)")
	}
	if args.ScopeGoal == nil {
		return nil, nil, fmt.Errorf("scope_goal is required for creating a readiness assessment (the desired number of protect surfaces)")
	}
	if *args.ScopeGoal < 1 {
		return nil, nil, fmt.Errorf("scope_goal must be at least 1")
	}

	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	// Fetch the current question set to validate the answers against
	questions, err := auxoClient.ZeroTrust.GetReadinessQuestions(ctx)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	if args.Version != nil && *args.Version != questions.Version {
		return nil, nil, fmt.Errorf("version %d does not match the current question set version %d; re-run getReadinessQuestions and answer the current questions", *args.Version, questions.Version)
	}

	now := time.Now()
	answerTimestamp := strconv.FormatInt(now.Unix(), 10)

	strategical, err := buildReadinessAnswers("strategical", args.Strategical, questions.Strategical, args.AnsweredBy, answerTimestamp)
	if err != nil {
		return nil, nil, err
	}
	tactical, err := buildReadinessAnswers("tactical", args.Tactical, questions.Tactical, args.AnsweredBy, answerTimestamp)
	if err != nil {
		return nil, nil, err
	}
	operational, err := buildReadinessAnswers("operational", args.Operational, questions.Operational, args.AnsweredBy, answerTimestamp)
	if err != nil {
		return nil, nil, err
	}

	answers := zerotrust.ReadinessAnswers{
		Timestamp:   now.Unix(),
		Version:     questions.Version,
		Strategical: strategical,
		Tactical:    tactical,
		Operational: operational,
		Scoping:     zerotrust.Scope{Goal: *args.ScopeGoal},
	}

	result, err := auxoClient.ZeroTrust.PostReadinessAnswers(ctx, answers)
	if err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	response := map[string]interface{}{
		"status":               "success",
		"id":                   result.ID,
		"assessment_timestamp": result.Timestamp,
		"version":              result.Version,
		"summary": map[string]interface{}{
			"strategical": dimensionSummary(strategical),
			"tactical":    dimensionSummary(tactical),
			"operational": dimensionSummary(operational),
			"scope_goal":  *args.ScopeGoal,
		},
	}

	jsonData, err := json.Marshal(response)
	if err != nil {
		return nil, nil, err
	}

	content := []mcp.Content{
		&mcp.TextContent{
			Text: string(jsonData),
		}}

	return &mcp.CallToolResult{
		Content: content,
	}, nil, nil
}

// Delete removes a readiness assessment by its ID
func (t *ReadinessTools) Delete(ctx context.Context, req *mcp.CallToolRequest, args types.ReadinessAssessmentIDParams) (*mcp.CallToolResult, any, error) {
	if args.ID == "" {
		return nil, nil, fmt.Errorf("id is required to delete a readiness assessment")
	}

	auxoClient, err := t.clientManager.CreateClient(ctx)
	if err != nil {
		return nil, nil, err
	}

	if err := auxoClient.ZeroTrust.DeleteReadinessAssessmentByID(ctx, args.ID); err != nil {
		return nil, nil, client.FriendlyError(err)
	}

	successMsg := map[string]string{
		"status":  "success",
		"message": fmt.Sprintf("Readiness assessment %s deleted successfully", args.ID),
	}

	jsonData, err := json.Marshal(successMsg)
	if err != nil {
		return nil, nil, err
	}

	content := []mcp.Content{
		&mcp.TextContent{
			Text: string(jsonData),
		}}

	return &mcp.CallToolResult{
		Content: content,
	}, nil, nil
}

// buildReadinessAnswers validates the provided answers against the question set of a
// dimension (strategical, tactical, operational) and converts them to go-auxo answers.
// Every question must be answered exactly once with actual and goal on the 1-5 CMMI scale.
func buildReadinessAnswers(dimension string, provided []types.ReadinessAnswerParams, questions []zerotrust.Questions, defaultAnsweredBy string, answerTimestamp string) ([]zerotrust.Answer, error) {
	valid := make(map[string]bool, len(questions))
	for _, q := range questions {
		valid[q.QuestionID] = true
	}

	seen := make(map[string]bool, len(provided))
	answers := make([]zerotrust.Answer, 0, len(provided))
	for _, a := range provided {
		if a.QuestionID == "" {
			return nil, fmt.Errorf("every %s answer requires a question_id (see getReadinessQuestions)", dimension)
		}
		if !valid[a.QuestionID] {
			return nil, fmt.Errorf("unknown %s question_id '%s'; use getReadinessQuestions to get the current question set", dimension, a.QuestionID)
		}
		if seen[a.QuestionID] {
			return nil, fmt.Errorf("duplicate answer for %s question '%s'", dimension, a.QuestionID)
		}
		seen[a.QuestionID] = true

		if a.Actual == nil || *a.Actual < 1 || *a.Actual > 5 {
			return nil, fmt.Errorf("%s question '%s': actual is required and must be between 1 and 5 (CMMI scale)", dimension, a.QuestionID)
		}
		if a.Goal == nil || *a.Goal < 1 || *a.Goal > 5 {
			return nil, fmt.Errorf("%s question '%s': goal is required and must be between 1 and 5 (CMMI scale)", dimension, a.QuestionID)
		}

		answeredBy := a.AnsweredBy
		if answeredBy == "" {
			answeredBy = defaultAnsweredBy
		}

		answers = append(answers, zerotrust.Answer{
			QuestionID: a.QuestionID,
			Actual:     *a.Actual,
			Goal:       *a.Goal,
			Timestamp:  answerTimestamp,
			AnsweredBy: answeredBy,
			Comment:    a.Comment,
		})
	}

	var missing []string
	for _, q := range questions {
		if !seen[q.QuestionID] {
			missing = append(missing, q.QuestionID)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing answers for %s question(s): %s (all questions must be answered)", dimension, strings.Join(missing, ", "))
	}

	return answers, nil
}

// dimensionSummary computes the average actual and goal maturity for a dimension
func dimensionSummary(answers []zerotrust.Answer) map[string]interface{} {
	if len(answers) == 0 {
		return map[string]interface{}{"questions": 0}
	}

	var actualSum, goalSum int
	for _, a := range answers {
		actualSum += a.Actual
		goalSum += a.Goal
	}

	n := float64(len(answers))
	return map[string]interface{}{
		"questions":      len(answers),
		"average_actual": round1(float64(actualSum) / n),
		"average_goal":   round1(float64(goalSum) / n),
	}
}

// round1 rounds to one decimal place
func round1(v float64) float64 {
	return float64(int(v*10+0.5)) / 10
}
