package dapr

import (
	"context"
	"fmt"

	"github.com/dapr/go-sdk/client"
)

// Bindings wraps Dapr bindings operations for external services
type Bindings struct {
	client          *Client
	blobStorageName string
}

// BindingRequest represents a request to an external binding
type BindingRequest struct {
	Name      string
	Operation string
	Data      []byte
	Metadata  map[string]string
}

// BindingResponse represents a response from an external binding
type BindingResponse struct {
	Data     []byte
	Metadata map[string]string
}

// NewBindings creates a new bindings instance
func NewBindings(client *Client) *Bindings {
	environment := client.GetEnvironment()
	var blobStorageName string
	
	switch environment {
	case "production", "staging":
		blobStorageName = getEnv("DAPR_BLOB_BINDING_NAME", "blob-storage")
	default:
		blobStorageName = getEnv("DAPR_BLOB_BINDING_NAME", "blob-storage-local")
	}

	return &Bindings{
		client:          client,
		blobStorageName: blobStorageName,
	}
}

// InvokeBinding invokes an external binding with the given request
func (b *Bindings) InvokeBinding(ctx context.Context, req *BindingRequest) (*BindingResponse, error) {
	bindingReq := &client.InvokeBindingRequest{
		Name:      req.Name,
		Operation: req.Operation,
		Data:      req.Data,
		Metadata:  req.Metadata,
	}

	resp, err := b.client.GetClient().InvokeBinding(ctx, bindingReq)
	if err != nil {
		return nil, fmt.Errorf("failed to invoke binding %s: %w", req.Name, err)
	}

	return &BindingResponse{
		Data:     resp.Data,
		Metadata: resp.Metadata,
	}, nil
}

// UploadBlob uploads data to blob storage
func (b *Bindings) UploadBlob(ctx context.Context, blobName string, data []byte, contentType string) error {
	metadata := map[string]string{
		"blobName": blobName,
	}
	
	if contentType != "" {
		metadata["contentType"] = contentType
	}

	req := &BindingRequest{
		Name:      b.blobStorageName,
		Operation: "create",
		Data:      data,
		Metadata:  metadata,
	}

	_, err := b.InvokeBinding(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to upload blob %s: %w", blobName, err)
	}

	return nil
}

// DownloadBlob downloads data from blob storage
func (b *Bindings) DownloadBlob(ctx context.Context, blobName string) ([]byte, error) {
	req := &BindingRequest{
		Name:      b.blobStorageName,
		Operation: "get",
		Metadata: map[string]string{
			"blobName": blobName,
		},
	}

	resp, err := b.InvokeBinding(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to download blob %s: %w", blobName, err)
	}

	return resp.Data, nil
}

// DeleteBlob deletes a blob from storage
func (b *Bindings) DeleteBlob(ctx context.Context, blobName string) error {
	req := &BindingRequest{
		Name:      b.blobStorageName,
		Operation: "delete",
		Metadata: map[string]string{
			"blobName": blobName,
		},
	}

	_, err := b.InvokeBinding(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to delete blob %s: %w", blobName, err)
	}

	return nil
}

// ListBlobs lists blobs in storage with optional prefix
func (b *Bindings) ListBlobs(ctx context.Context, prefix string) ([]string, error) {
	metadata := map[string]string{}
	if prefix != "" {
		metadata["prefix"] = prefix
	}

	req := &BindingRequest{
		Name:      b.blobStorageName,
		Operation: "list",
		Metadata:  metadata,
	}

	_, err := b.InvokeBinding(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to list blobs: %w", err)
	}

	// Parse the response to extract blob names
	// This would depend on the specific binding response format
	var blobNames []string
	// Implementation would parse resp.Data or resp.Metadata for blob names
	
	return blobNames, nil
}

// GetBlobMetadata retrieves metadata for a blob
func (b *Bindings) GetBlobMetadata(ctx context.Context, blobName string) (map[string]string, error) {
	req := &BindingRequest{
		Name:      b.blobStorageName,
		Operation: "get",
		Metadata: map[string]string{
			"blobName":     blobName,
			"includeBody":  "false",
			"metadataOnly": "true",
		},
	}

	resp, err := b.InvokeBinding(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get blob metadata for %s: %w", blobName, err)
	}

	return resp.Metadata, nil
}

// CreateBlobURL creates a URL for blob access (if supported by binding)
func (b *Bindings) CreateBlobURL(ctx context.Context, blobName string, expiryMinutes int) (string, error) {
	metadata := map[string]string{
		"blobName": blobName,
	}
	
	if expiryMinutes > 0 {
		metadata["expiryMinutes"] = fmt.Sprintf("%d", expiryMinutes)
	}

	req := &BindingRequest{
		Name:      b.blobStorageName,
		Operation: "createURL",
		Metadata:  metadata,
	}

	resp, err := b.InvokeBinding(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to create blob URL for %s: %w", blobName, err)
	}

	if url, exists := resp.Metadata["url"]; exists {
		return url, nil
	}

	return string(resp.Data), nil
}

// CreateStoragePath creates a standardized storage path for content
func (b *Bindings) CreateStoragePath(domain, year, month, contentID, hash, extension string) string {
	environment := b.client.GetEnvironment()
	return fmt.Sprintf("%s/%s/%s/%s/%s/%s.%s", 
		environment, domain, year, month, contentID, hash, extension)
}

// QueryLoki queries Grafana Loki for log data
func (b *Bindings) QueryLoki(ctx context.Context, bindingName, query string) ([]byte, error) {
	req := &BindingRequest{
		Name:      bindingName,
		Operation: "query",
		Data:      []byte(query),
		Metadata: map[string]string{
			"query": query,
		},
	}

	resp, err := b.InvokeBinding(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to query Loki: %w", err)
	}

	return resp.Data, nil
}

