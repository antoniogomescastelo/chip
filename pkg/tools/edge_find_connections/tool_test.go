package edge_find_connections_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/collibra/chip/pkg/clients"
	tools "github.com/collibra/chip/pkg/tools/edge_find_connections"
	"github.com/collibra/chip/pkg/tools/testutil"
)

func TestEdgeFindConnections(t *testing.T) {
	handler := http.NewServeMux()
	handler.Handle("/edge/api/rest/v2/connections/find", testutil.JsonHandlerInOut(func(r *http.Request, in clients.ConnectionFindRequest) (int, []clients.Connection) {
		return http.StatusOK, []clients.Connection{{Id: "conn-1", Name: "Databricks"}}
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	output, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{Name: "Databricks"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Error != "" {
		t.Fatalf("unexpected output error: %s", output.Error)
	}
	if len(output.Connections) != 1 || output.Connections[0].Id != "conn-1" {
		t.Fatalf("expected conn-1, got %+v", output.Connections)
	}
}

func TestEdgeFindConnectionsAPIError(t *testing.T) {
	handler := http.NewServeMux()
	handler.Handle("/edge/api/rest/v2/connections/find", testutil.JsonHandlerOut(func(r *http.Request) (int, string) {
		return http.StatusInternalServerError, "error"
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	output, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{})
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if output.Error == "" {
		t.Fatal("expected output error")
	}
}

func TestEdgeFindConnectionsInvalidSiteId(t *testing.T) {
	server := httptest.NewServer(http.NewServeMux())
	defer server.Close()

	_, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{EdgeSiteId: "bad"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
