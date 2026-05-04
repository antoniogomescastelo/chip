package edge_update_connection_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/collibra/chip/pkg/clients"
	tools "github.com/collibra/chip/pkg/tools/edge_update_connection"
	"github.com/collibra/chip/pkg/tools/testutil"
)

func TestEdgeUpdateConnection(t *testing.T) {
	handler := http.NewServeMux()
	handler.Handle("/edge/api/rest/v2/connections/00000000-0000-0000-0000-000000000001", testutil.JsonHandlerInOut(func(r *http.Request, in clients.ConnectionUpdateRequest) (int, clients.Connection) {
		return http.StatusOK, clients.Connection{Id: "00000000-0000-0000-0000-000000000001", Name: in.Name}
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	output, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{
		ConnectionId: "00000000-0000-0000-0000-000000000001",
		Name:         "Updated Conn",
		TypeId:       "type-x",
		EdgeSiteId:   "00000000-0000-0000-0000-000000000002",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Success {
		t.Fatalf("expected success, got: %s", output.Error)
	}
	if output.Connection.Name != "Updated Conn" {
		t.Fatalf("expected Updated Conn, got %s", output.Connection.Name)
	}
}

func TestEdgeUpdateConnectionAPIError(t *testing.T) {
	handler := http.NewServeMux()
	handler.Handle("/edge/api/rest/v2/connections/00000000-0000-0000-0000-000000000001", testutil.JsonHandlerOut(func(r *http.Request) (int, string) {
		return http.StatusInternalServerError, "error"
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	output, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{
		ConnectionId: "00000000-0000-0000-0000-000000000001",
		Name:         "Conn",
		TypeId:       "type",
		EdgeSiteId:   "00000000-0000-0000-0000-000000000002",
	})
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if output.Success {
		t.Fatal("expected failure")
	}
}

func TestEdgeUpdateConnectionInvalidId(t *testing.T) {
	server := httptest.NewServer(http.NewServeMux())
	defer server.Close()

	_, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{
		ConnectionId: "bad",
		EdgeSiteId:   "00000000-0000-0000-0000-000000000001",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
