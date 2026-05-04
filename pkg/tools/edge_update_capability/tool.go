package edge_update_capability

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
	Id          string         `json:"id" jsonschema:"the UUID of the capability to update"`
	Name        string         `json:"name,omitempty" jsonschema:"updated capability name"`
	TypeId      string         `json:"typeId,omitempty" jsonschema:"updated capability type ID"`
	EdgeSiteId  string         `json:"edgeSiteId,omitempty" jsonschema:"updated UUID of the edge site"`
	Description string         `json:"description,omitempty" jsonschema:"updated capability description"`
	Parameters  map[string]any `json:"parameters,omitempty" jsonschema:"updated capability parameters"`
}

type Output struct {
	Capability *clients.Capability `json:"capability,omitempty" jsonschema:"the updated capability"`
	Success    bool                `json:"success" jsonschema:"whether the update succeeded"`
	Error      string              `json:"error,omitempty" jsonschema:"error message if the update failed"`
}

func NewTool(collibraClient *http.Client) *chip.Tool[Input, Output] {
	return &chip.Tool[Input, Output]{
		Name:        "edge_update_capability",
		Description: "Updates an existing Edge capability or creates a new one if it does not exist (upsert by UUID). Provide only the fields to change.",
		Handler:     handler(collibraClient),
		Permissions: []string{},
		Annotations: &mcp.ToolAnnotations{DestructiveHint: chip.Ptr(true)},
	}
}

func handler(collibraClient *http.Client) chip.ToolHandlerFunc[Input, Output] {
	return func(ctx context.Context, input Input) (Output, error) {
		if err := validation.UUID("id", input.Id); err != nil {
			return Output{}, err
		}
		if err := validation.UUIDOptional("edgeSiteId", input.EdgeSiteId); err != nil {
			return Output{}, err
		}
		cap, err := clients.UpdateCapability(ctx, collibraClient, input.Id, clients.CapabilityUpdateRequest{
			Name:        input.Name,
			TypeId:      input.TypeId,
			EdgeSiteId:  input.EdgeSiteId,
			Description: input.Description,
			Parameters:  input.Parameters,
		})
		if err != nil {
			return Output{Success: false, Error: fmt.Sprintf("failed to update capability: %s", err.Error())}, nil
		}
		return Output{Capability: cap, Success: true}, nil
	}
}
