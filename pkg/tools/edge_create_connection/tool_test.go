package edge_create_connection_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/collibra/chip/pkg/clients"
	tools "github.com/collibra/chip/pkg/tools/edge_create_connection"
	"github.com/collibra/chip/pkg/tools/testutil"
)

func TestEdgeCreateConnection(t *testing.T) {
	handler := http.NewServeMux()
	handler.Handle("/edge/api/rest/v2/connections", testutil.JsonHandlerInOut(func(r *http.Request, in clients.ConnectionCreateRequest) (int, clients.Connection) {
		return http.StatusOK, clients.Connection{Id: "conn-new", Name: in.Name}
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	output, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{
		Name:       "My Connection",
		TypeId:     "databricks-conn",
		EdgeSiteId: "00000000-0000-0000-0000-000000000001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Success {
		t.Fatalf("expected success, got: %s", output.Error)
	}
	if output.Connection == nil || output.Connection.Id != "conn-new" {
		t.Fatal("expected connection id conn-new")
	}
}

func TestEdgeCreateConnectionAPIError(t *testing.T) {
	handler := http.NewServeMux()
	handler.Handle("/edge/api/rest/v2/connections", testutil.JsonHandlerOut(func(r *http.Request) (int, string) {
		return http.StatusBadRequest, "bad request"
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	output, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{
		Name:       "Conn",
		TypeId:     "type",
		EdgeSiteId: "00000000-0000-0000-0000-000000000001",
	})
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if output.Success {
		t.Fatal("expected failure")
	}
}

func TestEdgeCreateConnectionInvalidSiteId(t *testing.T) {
	server := httptest.NewServer(http.NewServeMux())
	defer server.Close()

	_, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{
		Name:       "Conn",
		TypeId:     "type",
		EdgeSiteId: "bad",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
