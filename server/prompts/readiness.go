package prompts

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/on2itsecurity/auxo-mcp/server/client"
)

// ReadinessPrompts handles readiness assessment related prompts
type ReadinessPrompts struct {
	clientManager *client.Manager
}

// NewReadinessPrompts creates a new instance of ReadinessPrompts
func NewReadinessPrompts(clientManager *client.Manager) *ReadinessPrompts {
	return &ReadinessPrompts{
		clientManager: clientManager,
	}
}

// RunAssessment generates a prompt that turns the assistant into a readiness assessment interviewer
func (p *ReadinessPrompts) RunAssessment(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	// Extract arguments from the request
	args := make(map[string]string)
	if req.Params.Arguments != nil {
		args = req.Params.Arguments
	}

	answeredBy := getStringArgFromStringMap(args, "answered_by", "")
	answeredByLine := "First ask for the e-mail address of the person answering (needed as answered_by when submitting)."
	if answeredBy != "" {
		answeredByLine = fmt.Sprintf("The answers are given by: %s (use this as answered_by when submitting).", answeredBy)
	}

	promptContent := fmt.Sprintf(`Please conduct a Zero Trust readiness assessment with me. Act as an experienced ON2IT Zero Trust consultant: professional, to the point, and pleasant to talk to.

%s

## How to run the assessment

### Step 1: Start the assessment
Call **startReadinessAssessment**. It returns questions in three dimensions (strategical, tactical, operational) plus a scoping question. Every question is answered on a 1-5 CMMI maturity scale with BOTH an "actual" (where we are today) and a "goal" (where we want to be).

In clients that support interactive MCP Apps this also opens an interactive questionnaire panel. If that panel appeared, let me fill it in there - don't interview me in parallel; stay available for questions and once the panel reports the submitted result, continue with Step 6. If no panel appeared, run the conversational interview below.

### Step 2: Let me choose the pace
Offer me two modes and respect my choice:
- **Guided interview** - walk through the dimensions one at a time. Present each question's caption with a short version of its explanation and the answer options, then ask for my actual and goal level. Group a handful of questions per message so it flows quickly; never dump all questions at once. Accept natural-language answers ("we do this ad hoc, but we want it managed") and translate them to the CMMI scale, confirming your interpretation.
- **Quick draft** - I describe my organization's situation in my own words, you draft actual and goal scores for ALL questions based on that, and we review and adjust the draft together.

### Step 3: Scoping
Ask for the ambition in number of protect surfaces (scope_goal). If I'm unsure, briefly explain what a protect surface is and suggest a realistic number based on the conversation.

### Step 4: Review before submitting
Show a compact summary table per dimension (question, actual, goal) with the average actual and goal per dimension, and highlight the biggest gaps between actual and goal. Ask me to confirm or adjust. Assessments cannot be edited after submission, so only submit after my explicit confirmation.

### Step 5: Submit
Call **createReadinessAssessment** with all answers, scope_goal and answered_by. Every question in every dimension must be answered (actual and goal, 1-5). Add my remarks as per-answer comments where they give useful context.

### Step 6: Present the result
After submission, present the outcome: the maturity per dimension, the biggest ambition gaps, and 2-3 concrete recommendations on where to focus first based on my answers.

## Ground rules
- One dimension at a time; keep the momentum, this should feel fast and effortless.
- Never invent scores in guided mode; in quick-draft mode clearly mark scores as a draft until I confirm them.
- Goal should normally be equal to or higher than actual; if I answer otherwise, double-check with me.
- If an existing assessment is relevant for comparison, you can use listReadinessAssessments and getReadinessAssessment.

Let's begin.`, answeredByLine)

	result := &mcp.GetPromptResult{
		Description: "An interactive Zero Trust readiness assessment interview, resulting in a submitted assessment",
		Messages: []*mcp.PromptMessage{
			{
				Role: mcp.Role("user"),
				Content: &mcp.TextContent{
					Text: promptContent,
				},
			},
		},
	}

	return result, nil
}
