package edge_create_connection

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
	Name        string         `json:"name" jsonschema:"the connection name"`
	TypeId      string         `json:"typeId" jsonschema:"the ID of the connection type"`
	EdgeSiteId  string         `json:"edgeSiteId" jsonschema:"the UUID of the edge site where this connection is valid"`
	Description string         `json:"description,omitempty" jsonschema:"optional connection description"`
	Parameters  map[string]any `json:"parameters,omitempty" jsonschema:"optional connection parameters; exact keys are defined in the connection type manifest"`
	VaultId     string         `json:"vaultId,omitempty" jsonschema:"optional UUID of the vault to retrieve vault parameters from"`
}

type Output struct {
	Connection *clients.Connection `json:"connection,omitempty" jsonschema:"the created connection"`
	Success    bool                `json:"success" jsonschema:"whether the connection was created successfully"`
	Error      string              `json:"error,omitempty" jsonschema:"error message if creation failed"`
}

func NewTool(collibraClient *http.Client) *chip.Tool[Input, Output] {
	return &chip.Tool[Input, Output]{
		Name:        "edge_create_connection",
		Description: "Creates a new Edge connection. Requires the connection name, type ID, and edge site UUID.",
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
		if err := validation.UUIDOptional("vaultId", input.VaultId); err != nil {
			return Output{}, err
		}
		conn, err := clients.CreateConnection(ctx, collibraClient, clients.ConnectionCreateRequest{
			Name:        input.Name,
			TypeId:      input.TypeId,
			EdgeSiteId:  input.EdgeSiteId,
			Description: input.Description,
			Parameters:  input.Parameters,
			VaultId:     input.VaultId,
		})
		if err != nil {
			return Output{Success: false, Error: fmt.Sprintf("failed to create connection: %s", err.Error())}, nil
		}
		return Output{Connection: conn, Success: true}, nil
	}
}
