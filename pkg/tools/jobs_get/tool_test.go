package jobs_get_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/collibra/chip/pkg/clients"
	tools "github.com/collibra/chip/pkg/tools/jobs_get"
	"github.com/collibra/chip/pkg/tools/testutil"
)

func TestJobsGet(t *testing.T) {
	handler := http.NewServeMux()
	handler.Handle("/rest/jobs/v1/jobs/00000000-0000-0000-0000-000000000001", testutil.JsonHandlerOut(func(r *http.Request) (int, clients.JobV1) {
		return http.StatusOK, clients.JobV1{Id: "00000000-0000-0000-0000-000000000001", Name: "my-job", State: "COMPLETED"}
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	output, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{
		JobId: "00000000-0000-0000-0000-000000000001",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !output.Found {
		t.Fatalf("expected found, got error: %s", output.Error)
	}
	if output.Job == nil || output.Job.Id != "00000000-0000-0000-0000-000000000001" {
		t.Fatal("expected job id 00000000-0000-0000-0000-000000000001")
	}
}

func TestJobsGetAPIError(t *testing.T) {
	handler := http.NewServeMux()
	handler.Handle("/rest/jobs/v1/jobs/00000000-0000-0000-0000-000000000001", testutil.JsonHandlerOut(func(r *http.Request) (int, string) {
		return http.StatusNotFound, "not found"
	}))

	server := httptest.NewServer(handler)
	defer server.Close()

	output, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{
		JobId: "00000000-0000-0000-0000-000000000001",
	})
	if err != nil {
		t.Fatalf("unexpected go error: %v", err)
	}
	if output.Found {
		t.Fatal("expected not found")
	}
}

func TestJobsGetInvalidId(t *testing.T) {
	server := httptest.NewServer(http.NewServeMux())
	defer server.Close()

	_, err := tools.NewTool(testutil.NewClient(server)).Handler(t.Context(), tools.Input{JobId: "bad"})
	if err == nil {
		t.Fatal("expected validation error")
	}
}
