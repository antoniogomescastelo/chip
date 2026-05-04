package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const edgeAPIBasePath = "/edge/api/rest/v2"

// --- Types ---

type Capability struct {
	Id           string         `json:"id,omitempty"`
	Name         string         `json:"name,omitempty"`
	Description  string         `json:"description,omitempty"`
	EdgeSiteId   string         `json:"edgeSiteId,omitempty"`
	EdgeSiteName string         `json:"edgeSiteName,omitempty"`
	Type         *CapabilityType `json:"type,omitempty"`
	Parameters   map[string]any `json:"parameters,omitempty"`
}

type CapabilityType struct {
	Id string `json:"id,omitempty"`
}

type CapabilityCreateRequest struct {
	Name        string         `json:"name"`
	TypeId      string         `json:"typeId"`
	EdgeSiteId  string         `json:"edgeSiteId"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type CapabilityUpdateRequest struct {
	Name        string         `json:"name,omitempty"`
	TypeId      string         `json:"typeId,omitempty"`
	EdgeSiteId  string         `json:"edgeSiteId,omitempty"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type CapabilityFindRequest struct {
	EdgeSiteId string            `json:"edgeSiteId,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Parameters map[string]string `json:"parameters,omitempty"`
}

type CapabilityRunRequest struct {
	JobId           string         `json:"jobId,omitempty"`
	InFastNamespace bool           `json:"inFastNamespace,omitempty"`
	Parameters      map[string]any `json:"parameters,omitempty"`
	WorkflowName    string         `json:"workflowName,omitempty"`
}

type Connection struct {
	Id             string          `json:"id,omitempty"`
	Name           string          `json:"name,omitempty"`
	Description    string          `json:"description,omitempty"`
	EdgeSiteId     string          `json:"edgeSiteId,omitempty"`
	VaultId        string          `json:"vaultId,omitempty"`
	ConnectionType *EdgeConnectionType `json:"connectionType,omitempty"`
	Parameters     map[string]any  `json:"parameters,omitempty"`
}

type EdgeConnectionType struct {
	Id string `json:"id,omitempty"`
}

type ConnectionCreateRequest struct {
	Name        string         `json:"name"`
	TypeId      string         `json:"typeId"`
	EdgeSiteId  string         `json:"edgeSiteId"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
	VaultId     string         `json:"vaultId,omitempty"`
}

type ConnectionUpdateRequest struct {
	Name        string         `json:"name"`
	TypeId      string         `json:"typeId"`
	EdgeSiteId  string         `json:"edgeSiteId"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
	VaultId     string         `json:"vaultId,omitempty"`
}

type ConnectionFindRequest struct {
	EdgeSiteId    string `json:"edgeSiteId,omitempty"`
	Name          string `json:"name,omitempty"`
	NameMatchMode string `json:"nameMatchMode,omitempty"`
}

type TestConnectionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	JobId   string `json:"jobId,omitempty"`
}

type EdgeJobStatusLog struct {
	JobId               string `json:"jobId,omitempty"`
	Status              string `json:"status,omitempty"`
	Message             string `json:"message,omitempty"`
	LastUpdatedDateTime string `json:"lastUpdatedDateTime,omitempty"`
}

// --- Capability functions ---

func ListCapabilities(ctx context.Context, client *http.Client) ([]Capability, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", edgeAPIBasePath+"/capabilities", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	body, err := executeRequest(client, req)
	if err != nil {
		return nil, err
	}
	var capabilities []Capability
	if err := json.Unmarshal(body, &capabilities); err != nil {
		return nil, fmt.Errorf("failed to parse capabilities response: %w", err)
	}
	return capabilities, nil
}

func CreateCapability(ctx context.Context, client *http.Client, reqBody CapabilityCreateRequest) (*Capability, error) {
	return doCapabilityRequest(ctx, client, "POST", edgeAPIBasePath+"/capabilities", reqBody)
}

func FindCapabilities(ctx context.Context, client *http.Client, reqBody CapabilityFindRequest) ([]Capability, error) {
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", edgeAPIBasePath+"/capabilities/find", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	body, err := executeRequest(client, req)
	if err != nil {
		return nil, err
	}
	var caps []Capability
	if err := json.Unmarshal(body, &caps); err != nil {
		return nil, fmt.Errorf("failed to parse capabilities response: %w", err)
	}
	return caps, nil
}

func GetCapability(ctx context.Context, client *http.Client, id string) (*Capability, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", edgeAPIBasePath+"/capabilities/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	body, err := executeRequest(client, req)
	if err != nil {
		return nil, err
	}
	var cap Capability
	if err := json.Unmarshal(body, &cap); err != nil {
		return nil, fmt.Errorf("failed to parse capability response: %w", err)
	}
	return &cap, nil
}

func UpdateCapability(ctx context.Context, client *http.Client, id string, reqBody CapabilityUpdateRequest) (*Capability, error) {
	return doCapabilityRequest(ctx, client, "PUT", edgeAPIBasePath+"/capabilities/"+id, reqBody)
}

func DeleteCapability(ctx context.Context, client *http.Client, id string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", edgeAPIBasePath+"/capabilities/"+id, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	_, err = executeRequest(client, req)
	return err
}

func RunCapability(ctx context.Context, client *http.Client, id string, reqBody CapabilityRunRequest) (string, error) {
	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", edgeAPIBasePath+"/capabilities/"+id+"/run", bytes.NewBuffer(data))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	body, err := executeRequest(client, req)
	if err != nil {
		return "", err
	}
	var jobId string
	if err := json.Unmarshal(body, &jobId); err != nil {
		return "", fmt.Errorf("failed to parse run response: %w", err)
	}
	return jobId, nil
}

func doCapabilityRequest(ctx context.Context, client *http.Client, method, url string, reqBody any) (*Capability, error) {
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	body, err := executeRequest(client, req)
	if err != nil {
		return nil, err
	}
	var cap Capability
	if err := json.Unmarshal(body, &cap); err != nil {
		return nil, fmt.Errorf("failed to parse capability response: %w", err)
	}
	return &cap, nil
}

// --- Connection functions ---

func ListConnections(ctx context.Context, client *http.Client) ([]Connection, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", edgeAPIBasePath+"/connections", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	body, err := executeRequest(client, req)
	if err != nil {
		return nil, err
	}
	var connections []Connection
	if err := json.Unmarshal(body, &connections); err != nil {
		return nil, fmt.Errorf("failed to parse connections response: %w", err)
	}
	return connections, nil
}

func CreateConnection(ctx context.Context, client *http.Client, reqBody ConnectionCreateRequest) (*Connection, error) {
	return doConnectionRequest(ctx, client, "POST", edgeAPIBasePath+"/connections", reqBody)
}

func FindConnections(ctx context.Context, client *http.Client, reqBody ConnectionFindRequest) ([]Connection, error) {
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", edgeAPIBasePath+"/connections/find", bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	body, err := executeRequest(client, req)
	if err != nil {
		return nil, err
	}
	var connections []Connection
	if err := json.Unmarshal(body, &connections); err != nil {
		return nil, fmt.Errorf("failed to parse connections response: %w", err)
	}
	return connections, nil
}

func GetConnection(ctx context.Context, client *http.Client, id string) (*Connection, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", edgeAPIBasePath+"/connections/"+id, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	body, err := executeRequest(client, req)
	if err != nil {
		return nil, err
	}
	var conn Connection
	if err := json.Unmarshal(body, &conn); err != nil {
		return nil, fmt.Errorf("failed to parse connection response: %w", err)
	}
	return &conn, nil
}

func UpdateConnection(ctx context.Context, client *http.Client, id string, reqBody ConnectionUpdateRequest) (*Connection, error) {
	return doConnectionRequest(ctx, client, "PUT", edgeAPIBasePath+"/connections/"+id, reqBody)
}

func DeleteConnection(ctx context.Context, client *http.Client, id string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", edgeAPIBasePath+"/connections/"+id, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	_, err = executeRequest(client, req)
	return err
}

type testConnectionParams struct {
	TimeoutSec int `url:"timeoutSec,omitempty"`
}

func TestEdgeConnection(ctx context.Context, client *http.Client, id string, timeoutSec int) (*TestConnectionResponse, error) {
	params := testConnectionParams{TimeoutSec: timeoutSec}
	endpoint, err := buildUrl(edgeAPIBasePath+"/connections/"+id+"/test", params)
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
	var result TestConnectionResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse test connection response: %w", err)
	}
	return &result, nil
}

func doConnectionRequest(ctx context.Context, client *http.Client, method, url string, reqBody any) (*Connection, error) {
	data, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	body, err := executeRequest(client, req)
	if err != nil {
		return nil, err
	}
	var conn Connection
	if err := json.Unmarshal(body, &conn); err != nil {
		return nil, fmt.Errorf("failed to parse connection response: %w", err)
	}
	return &conn, nil
}

// --- Edge Job functions ---

func CancelEdgeJob(ctx context.Context, client *http.Client, id string) error {
	req, err := http.NewRequestWithContext(ctx, "POST", edgeAPIBasePath+"/jobs/"+id+"/cancel", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	_, err = executeRequest(client, req)
	return err
}

func GetEdgeJobStatus(ctx context.Context, client *http.Client, id string) (*EdgeJobStatusLog, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", edgeAPIBasePath+"/jobs/"+id+"/statusLog", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	body, err := executeRequest(client, req)
	if err != nil {
		return nil, err
	}
	var log EdgeJobStatusLog
	if err := json.Unmarshal(body, &log); err != nil {
		return nil, fmt.Errorf("failed to parse job status response: %w", err)
	}
	return &log, nil
}

func GetEdgeJobStatusHistory(ctx context.Context, client *http.Client, id string) ([]EdgeJobStatusLog, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", edgeAPIBasePath+"/jobs/"+id+"/statusLogHistory", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	body, err := executeRequest(client, req)
	if err != nil {
		return nil, err
	}
	var logs []EdgeJobStatusLog
	if err := json.Unmarshal(body, &logs); err != nil {
		return nil, fmt.Errorf("failed to parse job status history response: %w", err)
	}
	return logs, nil
}