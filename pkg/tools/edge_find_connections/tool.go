package edge_find_connections

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
	EdgeSiteId    string `json:"edgeSiteId,omitempty" jsonschema:"optional UUID of the edge site to filter connections by"`
	Name          string `json:"name,omitempty" jsonschema:"optional connection name to filter by"`
	NameMatchMode string `json:"nameMatchMode,omitempty" jsonschema:"name match mode: ANYWHERE (default) or EXACT"`
}

type Output struct {
	Connections []clients.Connection `json:"connections" jsonschema:"list of matching connections"`
	Count       int                  `json:"count" jsonschema:"number of matching connections"`
	Error       string               `json:"error,omitempty" jsonschema:"error message if the request failed"`
}

func NewTool(collibraClient *http.Client) *chip.Tool[Input, Output] {
	return &chip.Tool[Input, Output]{
		Name:        "edge_find_connections",
		Description: "Finds Edge connections based on search criteria. Filter by edge site UUID, connection name, or name match mode (ANYWHERE or EXACT).",
		Handler:     handler(collibraClient),
		Permissions: []string{},
		Annotations: &mcp.ToolAnnotations{ReadOnlyHint: true},
	}
}

func handler(collibraClient *http.Client) chip.ToolHandlerFunc[Input, Output] {
	return func(ctx context.Context, input Input) (Output, error) {
		if err := validation.UUIDOptional("edgeSiteId", input.EdgeSiteId); err != nil {
			return Output{}, err
		}
		conns, err := clients.FindConnections(ctx, collibraClient, clients.ConnectionFindRequest{
			EdgeSiteId:    input.EdgeSiteId,
			Name:          input.Name,
			NameMatchMode: input.NameMatchMode,
		})
		if err != nil {
			return Output{Connections: []clients.Connection{}, Error: fmt.Sprintf("failed to find connections: %s", err.Error())}, nil
		}
		if conns == nil {
			conns = []clients.Connection{}
		}
		return Output{Connections: conns, Count: len(conns)}, nil
	}
}
