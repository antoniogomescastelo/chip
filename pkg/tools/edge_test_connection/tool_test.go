package edge_test_connection_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/collibra/chip/pkg/clients"
	tools "github.com/collibra/chip/pkg/tools/edge_test_connection"
	"github.com/collibra/chip/pkg/tools/testutil"
)

func TestEdgeTestConnection(t *testing.T) {
	handler := http.NewServeMux()
	handler.Handle("/edge/api/rest/v2/connections/00000000-0000-0000-0000-000000000001/test", testutil.JsonHandlerOut(func(r *http.Request) (int, clients.TestConnectionResponse) {
		return http.StatusOK, clients.TestConnectionResponse{Success: true, Message: "Connection OK"}
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	output, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{
		ConnectionId: "00000000-0000-0000-0000-000000000001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Error != "" {
		t.Fatalf("unexpected output error: %s", output.Error)
	}
	if !output.Success {
		t.Fatal("expected test success")
	}
	if output.Message != "Connection OK" {
		t.Fatalf("expected message Connection OK, got %s", output.Message)
	}
}

func TestEdgeTestConnectionAPIError(t *testing.T) {
	handler := http.NewServeMux()
	handler.Handle("/edge/api/rest/v2/connections/00000000-0000-0000-0000-000000000001/test", testutil.JsonHandlerOut(func(r *http.Request) (int, string) {
		return http.StatusNotFound, "not found"
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	output, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{
		ConnectionId: "00000000-0000-0000-0000-000000000001",
	})
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if output.Error == "" {
		t.Fatal("expected output error")
	}
}

func TestEdgeTestConnectionInvalidId(t *testing.T) {
	server := httptest.NewServer(http.NewServeMux())
	defer server.Close()

	_, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{ConnectionId: "bad"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
