package edge_update_capability_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/collibra/chip/pkg/clients"
	tools "github.com/collibra/chip/pkg/tools/edge_update_capability"
	"github.com/collibra/chip/pkg/tools/testutil"
)

func TestEdgeUpdateCapability(t *testing.T) {
	handler := http.NewServeMux()
	handler.Handle("/edge/api/rest/v2/capabilities/00000000-0000-0000-0000-000000000001", testutil.JsonHandlerInOut(func(r *http.Request, in clients.CapabilityUpdateRequest) (int, clients.Capability) {
		return http.StatusOK, clients.Capability{Id: "00000000-0000-0000-0000-000000000001", Name: in.Name}
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	output, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{
		Id:   "00000000-0000-0000-0000-000000000001",
		Name: "Updated Cap",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Success {
		t.Fatalf("expected success, got: %s", output.Error)
	}
	if output.Capability == nil || output.Capability.Name != "Updated Cap" {
		t.Fatal("expected name Updated Cap")
	}
}

func TestEdgeUpdateCapabilityAPIError(t *testing.T) {
	handler := http.NewServeMux()
	handler.Handle("/edge/api/rest/v2/capabilities/00000000-0000-0000-0000-000000000001", testutil.JsonHandlerOut(func(r *http.Request) (int, string) {
		return http.StatusNotFound, "not found"
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	output, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{
		Id: "00000000-0000-0000-0000-000000000001",
	})
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if output.Success {
		t.Fatal("expected failure")
	}
}

func TestEdgeUpdateCapabilityInvalidId(t *testing.T) {
	server := httptest.NewServer(http.NewServeMux())
	defer server.Close()

	_, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{Id: "bad"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
