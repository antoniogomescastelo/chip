package edge_get_job_status

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
	Id string `json:"id" jsonschema:"the UUID of the Edge job to get the status for"`
}

type Output struct {
	Status *clients.EdgeJobStatusLog `json:"status,omitempty" jsonschema:"the latest job status log entry"`
	Found  bool                      `json:"found" jsonschema:"whether the job was found"`
	Error  string                    `json:"error,omitempty" jsonschema:"error message if retrieval failed"`
}

func NewTool(collibraClient *http.Client) *chip.Tool[Input, Output] {
	return &chip.Tool[Input, Output]{
		Name:        "edge_get_job_status",
		Description: "Gets the current latest status of an Edge job. Returns the most recent status log entry including status, message, and timestamp.",
		Handler:     handler(collibraClient),
		Permissions: []string{},
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}
}

func handler(collibraClient *http.Client) chip.ToolHandlerFunc[Input, Output] {
	return func(ctx context.Context, input Input) (Output, error) {
		if err := validation.UUID("id", input.Id); err != nil {
			return Output{}, err
		}
		status, err := clients.GetEdgeJobStatus(ctx, collibraClient, input.Id)
		if err != nil {
			return Output{Found: false, Error: fmt.Sprintf("failed to get job status: %s", err.Error())}, nil
		}
		return Output{Status: status, Found: true}, nil
	}
}
