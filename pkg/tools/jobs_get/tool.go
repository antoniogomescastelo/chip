package jobs_get

import (
	"context"
	"fmt"
	"net/http"

	"github.com/collibra/chip/pkg/chip"
	"github.com/collibra/chip/pkg/clients"
	"github.com/collibra/chip/pkg/tools/validation"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Input struct {
	JobId string `json:"jobId" jsonschema:"the UUID of the job to retrieve"`
}

type Output struct {
	Job   *clients.JobV1 `json:"job,omitempty" jsonschema:"the job details"`
	Found bool           `json:"found" jsonschema:"whether the job was found"`
	Error string         `json:"error,omitempty" jsonschema:"error message if retrieval failed"`
}

func NewTool(collibraClient *http.Client) *chip.Tool[Input, Output] {
	return &chip.Tool[Input, Output]{
		Name:        "jobs_get",
		Description: "Retrieves details for a specific job by its ID from the Collibra Jobs API.",
		Handler:     handler(collibraClient),
		Permissions: []string{},
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}
}

func handler(collibraClient *http.Client) chip.ToolHandlerFunc[Input, Output] {
	return func(ctx context.Context, input Input) (Output, error) {
		if err := validation.UUID("jobId", input.JobId); err != nil {
			return Output{}, err
		}
		job, err := clients.GetJobV1(ctx, collibraClient, input.JobId)
		if err != nil {
			return Output{Found: false, Error: fmt.Sprintf("failed to get job: %s", err.Error())}, nil
		}
		return Output{Job: job, Found: true}, nil
	}
}
