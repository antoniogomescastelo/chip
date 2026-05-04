package edge_create_capability

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
	Name        string         `json:"name" jsonschema:"the capability name"`
	TypeId      string         `json:"typeId" jsonschema:"the type of the capability to be created"`
	EdgeSiteId  string         `json:"edgeSiteId" jsonschema:"the UUID of the edge site where this capability will run"`
	Description string         `json:"description,omitempty" jsonschema:"optional capability description"`
	Parameters  map[string]any `json:"parameters,omitempty" jsonschema:"optional capability parameters; exact keys are defined in the capability type manifest"`
}

type Output struct {
	Capability *clients.Capability `json:"capability,omitempty" jsonschema:"the created capability"`
	Success    bool                `json:"success" jsonschema:"whether the capability was created successfully"`
	Error      string              `json:"error,omitempty" jsonschema:"error message if creation failed"`
}

func NewTool(collibraClient *http.Client) *chip.Tool[Input, Output] {
	return &chip.Tool[Input, Output]{
		Name:        "edge_create_capability",
		Description: "Creates a new Edge capability. Requires the capability name, type ID, and the edge site UUID where it will run.",
		Handler:     handler(collibraClient),
		Permissions: []string{},
		Annotations: &mcp.ToolAnnotations{DestructiveHint: chip.Ptr(true)},
	}
}

func handler(collibraClient *http.Client) chip.ToolHandlerFunc[Input, Output] {
	return func(ctx context.Context, input Input) (Output, error) {
		if err := validation.UUID("edgeSiteId", input.EdgeSiteId); err != nil {
			return Output{}, err
		}
		cap, err := clients.CreateCapability(ctx, collibraClient, clients.CapabilityCreateRequest{
			Name:        input.Name,
			TypeId:      input.TypeId,
			EdgeSiteId:  input.EdgeSiteId,
			Description: input.Description,
			Parameters:  input.Parameters,
		})
		if err != nil {
			return Output{Success: false, Error: fmt.Sprintf("failed to create capability: %s", err.Error())}, nil
		}
		return Output{Capability: cap, Success: true}, nil
	}
}