package edge_create_capability_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/collibra/chip/pkg/clients"
	tools "github.com/collibra/chip/pkg/tools/edge_create_capability"
	"github.com/collibra/chip/pkg/tools/testutil"
)

func TestEdgeCreateCapability(t *testing.T) {
	handler := http.NewServeMux()
	handler.Handle("/edge/api/rest/v2/capabilities", testutil.JsonHandlerInOut(func(r *http.Request, in clients.CapabilityCreateRequest) (int, clients.Capability) {
		return http.StatusOK, clients.Capability{Id: "cap-new", Name: in.Name, EdgeSiteId: in.EdgeSiteId}
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	output, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{
		Name:       "My Cap",
		TypeId:     "databricks-uc",
		EdgeSiteId: "00000000-0000-0000-0000-000000000001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Success {
		t.Fatalf("expected success, got error: %s", output.Error)
	}
	if output.Capability == nil || output.Capability.Id != "cap-new" {
		t.Fatalf("expected capability id cap-new")
	}
}

func TestEdgeCreateCapabilityAPIError(t *testing.T) {
	handler := http.NewServeMux()
	handler.Handle("/edge/api/rest/v2/capabilities", testutil.JsonHandlerOut(func(r *http.Request) (int, string) {
		return http.StatusBadRequest, "invalid request"
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	output, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{
		Name:       "Bad Cap",
		TypeId:     "type-x",
		EdgeSiteId: "00000000-0000-0000-0000-000000000001",
	})
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if output.Success {
		t.Fatal("expected failure")
	}
	if output.Error == "" {
		t.Fatal("expected error message")
	}
}

func TestEdgeCreateCapabilityInvalidSiteId(t *testing.T) {
	server := httptest.NewServer(http.NewServeMux())
	defer server.Close()

	_, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{
		Name:       "Cap",
		TypeId:     "type-x",
		EdgeSiteId: "not-a-uuid",
	})
	if err == nil {
		t.Fatal("expected validation error for invalid edgeSiteId")
	}
}