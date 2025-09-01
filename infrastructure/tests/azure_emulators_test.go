package tests

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// addAzureStorageAuth adds Azure Storage Shared Key authentication to HTTP requests
func addAzureStorageAuth(req *http.Request, accountName, accountKey string) error {
	// Azurite development account key
	keyBytes, err := base64.StdEncoding.DecodeString(accountKey)
	if err != nil {
		return err
	}

	// Add required date header
	req.Header.Set("x-ms-date", time.Now().UTC().Format(time.RFC1123))
	
	// Create canonical string for signature
	canonicalString := createCanonicalString(req, accountName)
	
	// Sign the canonical string
	h := hmac.New(sha256.New, keyBytes)
	h.Write([]byte(canonicalString))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	
	// Add Authorization header
	authHeader := fmt.Sprintf("SharedKey %s:%s", accountName, signature)
	req.Header.Set("Authorization", authHeader)
	
	return nil
}

// createCanonicalString creates the canonical string for Azure Storage authentication
func createCanonicalString(req *http.Request, accountName string) string {
	// Simplified canonical string for Azurite compatibility
	verb := req.Method
	contentEncoding := ""
	contentLanguage := ""
	contentLength := ""
	if req.ContentLength > 0 {
		contentLength = fmt.Sprintf("%d", req.ContentLength)
	}
	contentMD5 := ""
	contentType := req.Header.Get("Content-Type")
	date := ""
	ifModifiedSince := ""
	ifMatch := ""
	ifNoneMatch := ""
	ifUnmodifiedSince := ""
	range_ := ""
	
	// Get canonicalized headers (x-ms-* headers)
	canonicalizedHeaders := ""
	var xmsHeaders []string
	for k := range req.Header {
		if strings.HasPrefix(strings.ToLower(k), "x-ms-") {
			xmsHeaders = append(xmsHeaders, strings.ToLower(k))
		}
	}
	sort.Strings(xmsHeaders)
	for _, header := range xmsHeaders {
		canonicalizedHeaders += fmt.Sprintf("%s:%s\n", header, req.Header.Get(header))
	}
	
	// Get canonicalized resource
	canonicalizedResource := fmt.Sprintf("/%s%s", accountName, req.URL.Path)
	if req.URL.RawQuery != "" {
		params, _ := url.ParseQuery(req.URL.RawQuery)
		var keys []string
		for k := range params {
			keys = append(keys, strings.ToLower(k))
		}
		sort.Strings(keys)
		for _, key := range keys {
			canonicalizedResource += fmt.Sprintf("\n%s:%s", key, params[key][0])
		}
	}
	
	return fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s\n%s%s",
		verb, contentEncoding, contentLanguage, contentLength, contentMD5, contentType,
		date, ifModifiedSince, ifMatch, ifNoneMatch, ifUnmodifiedSince, range_,
		canonicalizedHeaders, canonicalizedResource)
}

func TestAzuriteStorageEmulator(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Skip if not in integration test environment
	if os.Getenv("INTEGRATION_TESTS") != "true" {
		t.Skip("Skipping integration test: INTEGRATION_TESTS not set to true")
	}

	t.Run("azurite blob storage container creation", func(t *testing.T) {
		// Test: Azurite can create blob containers
		azuritePort := getEnvWithDefault("AZURITE_BLOB_PORT", "10000")
		containerName := "test-container"
		
		// Create container using Azure Blob Storage REST API
		containerURL := fmt.Sprintf("http://localhost:%s/devstoreaccount1/%s?restype=container", 
			azuritePort, containerName)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "PUT", containerURL, nil)
		require.NoError(t, err)
		
		// Add required Azure Storage headers
		req.Header.Set("x-ms-version", "2020-04-08")
		req.Header.Set("x-ms-client-request-id", "test-client-request")
		
		// Add Azure Storage authentication
		accountKey := "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="
		err = addAzureStorageAuth(req, "devstoreaccount1", accountKey)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Azurite blob storage not accessible at port %s: %v", azuritePort, err)
		}
		require.NoError(t, err, "Should be able to connect to Azurite")
		defer resp.Body.Close()
		
		// Container creation should succeed or already exist
		assert.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusConflict,
			"Container creation should succeed or already exist, got status: %d", resp.StatusCode)
	})

	t.Run("blob storage read/write operations", func(t *testing.T) {
		// Test: Can perform basic blob operations
		azuritePort := getEnvWithDefault("AZURITE_BLOB_PORT", "10000")
		containerName := "test-container"
		blobName := "test-blob.txt"
		testContent := "test content for integration testing"
		
		// First ensure container exists
		containerURL := fmt.Sprintf("http://localhost:%s/devstoreaccount1/%s?restype=container", 
			azuritePort, containerName)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "PUT", containerURL, nil)
		require.NoError(t, err)
		req.Header.Set("x-ms-version", "2020-04-08")
		
		// Add authentication for container creation
		accountKey := "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="
		err = addAzureStorageAuth(req, "devstoreaccount1", accountKey)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		require.NoError(t, err)
		resp.Body.Close()
		
		// Upload blob
		blobURL := fmt.Sprintf("http://localhost:%s/devstoreaccount1/%s/%s", 
			azuritePort, containerName, blobName)
		
		uploadReq, err := http.NewRequestWithContext(ctx, "PUT", blobURL, 
			strings.NewReader(testContent))
		require.NoError(t, err)
		
		uploadReq.Header.Set("x-ms-version", "2020-04-08")
		uploadReq.Header.Set("x-ms-blob-type", "BlockBlob")
		uploadReq.Header.Set("Content-Type", "text/plain")
		uploadReq.Header.Set("Content-Length", fmt.Sprintf("%d", len(testContent)))
		
		// Add authentication for blob upload
		err = addAzureStorageAuth(uploadReq, "devstoreaccount1", accountKey)
		require.NoError(t, err)
		
		uploadResp, err := client.Do(uploadReq)
		if err == nil {
			defer uploadResp.Body.Close()
			assert.Equal(t, http.StatusCreated, uploadResp.StatusCode,
				"Blob upload should succeed")
			
			// Try to read the blob back
			downloadReq, err := http.NewRequestWithContext(ctx, "GET", blobURL, nil)
			require.NoError(t, err)
			downloadReq.Header.Set("x-ms-version", "2020-04-08")
			
			// Add authentication for blob download
			err = addAzureStorageAuth(downloadReq, "devstoreaccount1", accountKey)
			require.NoError(t, err)
			
			downloadResp, err := client.Do(downloadReq)
			if err == nil {
				defer downloadResp.Body.Close()
				assert.Equal(t, http.StatusOK, downloadResp.StatusCode,
					"Blob download should succeed")
				
				body, err := io.ReadAll(downloadResp.Body)
				if err == nil {
					assert.Equal(t, testContent, string(body),
						"Downloaded content should match uploaded content")
				}
			}
		}
	})

	t.Run("storage backend health monitoring", func(t *testing.T) {
		// Test: Storage backend reports healthy status
		azuritePort := getEnvWithDefault("AZURITE_BLOB_PORT", "10000")
		
		// Test Azurite health by checking service availability
		healthURL := fmt.Sprintf("http://localhost:%s/devstoreaccount1?comp=properties&restype=service", 
			azuritePort)
		
		client := &http.Client{Timeout: 5 * time.Second}
		req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
		require.NoError(t, err)
		req.Header.Set("x-ms-version", "2020-04-08")
		
		// Add authentication for health check
		accountKey := "Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw=="
		err = addAzureStorageAuth(req, "devstoreaccount1", accountKey)
		require.NoError(t, err)
		
		resp, err := client.Do(req)
		if err != nil {
			t.Logf("Azurite health check failed: %v", err)
		}
		require.NoError(t, err, "Azurite should be healthy")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode,
			"Azurite health check should return 200 OK")
	})
}

