package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const jobsV1BasePath = "/rest/jobs/v1"

// --- Types ---

type JobV1 struct {
	Id                 string  `json:"id,omitempty"`
	Name               string  `json:"name,omitempty"`
	Type               string  `json:"type,omitempty"`
	State              string  `json:"state,omitempty"`
	Result             string  `json:"result,omitempty"`
	Message            string  `json:"message,omitempty"`
	ProgressPercentage int     `json:"progressPercentage,omitempty"`
	StartDate          string  `json:"startDate,omitempty"`
	EndDate            string  `json:"endDate,omitempty"`
	CreatedBy          string  `json:"createdBy,omitempty"`
	CreatedOn          string  `json:"createdOn,omitempty"`
	LastModifiedBy     string  `json:"lastModifiedBy,omitempty"`
	LastModifiedOn     string  `json:"lastModifiedOn,omitempty"`
	User               string  `json:"user,omitempty"`
	SelfManaged        bool    `json:"selfManaged,omitempty"`
}

type JobV1PagedResponse struct {
	Results    []JobV1 `json:"results"`
	NextCursor string  `json:"nextCursor,omitempty"`
}

type jobsV1FindParams struct {
	Name          string   `url:"name,omitempty"`
	NameMatchMode string   `url:"nameMatchMode,omitempty"`
	Result        []string `url:"result,omitempty"`
	State         []string `url:"state,omitempty"`
	Type          []string `url:"type,omitempty"`
	User          string   `url:"user,omitempty"`
	SortField     string   `url:"sortField,omitempty"`
	SortOrder     string   `url:"sortOrder,omitempty"`
	Cursor        string   `url:"cursor,omitempty"`
	PageSize      int      `url:"pageSize,omitempty"`
}

// --- Functions ---

func FindJobsV1(
	ctx context.Context,
	client *http.Client,
	name string,
	nameMatchMode string,
	result []string,
	state []string,
	jobType []string,
	user string,
	sortField string,
	sortOrder string,
	cursor string,
	pageSize int,
) (*JobV1PagedResponse, error) {
	params := jobsV1FindParams{
		Name:          name,
		NameMatchMode: nameMatchMode,
		Result:        result,
		State:         state,
		Type:          jobType,
		User:          user,
		SortField:     sortField,
		SortOrder:     sortOrder,
		Cursor:        cursor,
		PageSize:      pageSize,
	}

	endpoint, err := buildUrl(jobsV1BasePath+"/jobs", params)
	if err != nil {
		return nil, fmt.Errorf("failed to build endpoint: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	body, err := executeRequest(client, req)
	if err != nil {
		return nil, err
	}

	var resp JobV1PagedResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse jobs response: %w", err)
	}
	return &resp, nil
}

func GetJobV1(ctx context.Context, client *http.Client, jobId string) (*JobV1, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", jobsV1BasePath+"/jobs/"+jobId, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	body, err := executeRequest(client, req)
	if err != nil {
		return nil, err
	}
	var job JobV1
	if err := json.Unmarshal(body, &job); err != nil {
		return nil, fmt.Errorf("failed to parse job response: %w", err)
	}
	return &job, nil
}