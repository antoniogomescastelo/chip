package edge_get_job_status_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/collibra/chip/pkg/clients"
	tools "github.com/collibra/chip/pkg/tools/edge_get_job_status"
	"github.com/collibra/chip/pkg/tools/testutil"
)

func TestEdgeGetJobStatus(t *testing.T) {
	handler := http.NewServeMux()
	handler.Handle("/edge/api/rest/v2/jobs/00000000-0000-0000-0000-000000000001/statusLog", testutil.JsonHandlerOut(func(r *http.Request) (int, clients.EdgeJobStatusLog) {
		return http.StatusOK, clients.EdgeJobStatusLog{
			JobId:   "00000000-0000-0000-0000-000000000001",
			Status:  "SUCCEEDED",
			Message: "Job completed",
		}
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	output, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{
		Id: "00000000-0000-0000-0000-000000000001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Found {
		t.Fatalf("expected found, got error: %s", output.Error)
	}
	if output.Status == nil || output.Status.Status != "SUCCEEDED" {
		t.Fatal("expected status SUCCEEDED")
	}
}

func TestEdgeGetJobStatusNotFound(t *testing.T) {
	handler := http.NewServeMux()
	handler.Handle("/edge/api/rest/v2/jobs/00000000-0000-0000-0000-000000000002/statusLog", testutil.JsonHandlerOut(func(r *http.Request) (int, string) {
		return http.StatusNotFound, "not found"
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	output, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{
		Id: "00000000-0000-0000-0000-000000000002",
	})
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if output.Found {
		t.Fatal("expected not found")
	}
}

func TestEdgeGetJobStatusInvalidId(t *testing.T) {
	server := httptest.NewServer(http.NewServeMux())
	defer server.Close()

	_, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{Id: "bad"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
