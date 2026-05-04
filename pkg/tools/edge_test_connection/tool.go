package edge_test_connection

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
	ConnectionId string `json:"connectionId" jsonschema:"the UUID of the connection to test"`
	TimeoutSec   int    `json:"timeoutSec,omitempty" jsonschema:"optional timeout in seconds; if not provided the test runs asynchronously"`
}

type Output struct {
	Success bool   `json:"success" jsonschema:"whether the connection test passed"`
	Message string `json:"message,omitempty" jsonschema:"result message from the connection test"`
	JobId   string `json:"jobId,omitempty" jsonschema:"job UUID if the test runs asynchronously"`
	Error   string `json:"error,omitempty" jsonschema:"error message if the test request failed"`
}

func NewTool(collibraClient *http.Client) *chip.Tool[Input, Output] {
	return &chip.Tool[Input, Output]{
		Name:        "edge_test_connection",
		Description: "Tests an Edge connection. Optionally waits up to timeoutSec seconds for a synchronous result; without a timeout the test is submitted asynchronously and a job ID is returned.",
		Handler:     handler(collibraClient),
		Permissions: []string{},
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}
}

func handler(collibraClient *http.Client) chip.ToolHandlerFunc[Input, Output] {
	return func(ctx context.Context, input Input) (Output, error) {
		if err := validation.UUID("connectionId", input.ConnectionId); err != nil {
			return Output{}, err
		}
		result, err := clients.TestEdgeConnection(ctx, collibraClient, input.ConnectionId, input.TimeoutSec)
		if err != nil {
			return Output{Success: false, Error: fmt.Sprintf("failed to test connection: %s", err.Error())}, nil
		}
		return Output{
			Success: result.Success,
			Message: result.Message,
			JobId:   result.JobId,
		}, nil
	}
}
