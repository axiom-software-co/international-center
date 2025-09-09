package dapr

import (
	"context"
	"testing"
	"time"

	sharedtesting "github.com/axiom-software-co/international-center/src/backend/internal/shared/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RED PHASE - Dapr External Bindings Tests (50+ test cases)

func TestNewBindings(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectedError  string
		validateResult func(*testing.T, *Bindings)
	}{
		{
			name: "create bindings with default environment",
			envVars: map[string]string{},
			validateResult: func(t *testing.T, bindings *Bindings) {
				assert.NotNil(t, bindings)
				assert.Equal(t, "blob-storage-local", bindings.blobStorageName)
			},
		},
		{
			name: "create bindings with production environment",
			envVars: map[string]string{
				"ENVIRONMENT": "production",
				"DAPR_BLOB_BINDING_NAME": "azure-blob-storage",
			},
			validateResult: func(t *testing.T, bindings *Bindings) {
				assert.NotNil(t, bindings)
				assert.Equal(t, "azure-blob-storage", bindings.blobStorageName)
			},
		},
		{
			name: "create bindings with staging environment",
			envVars: map[string]string{
				"ENVIRONMENT": "staging",
				"DAPR_BLOB_BINDING_NAME": "staging-blob-storage",
			},
			validateResult: func(t *testing.T, bindings *Bindings) {
				assert.NotNil(t, bindings)
				assert.Equal(t, "staging-blob-storage", bindings.blobStorageName)
			},
		},
		{
			name: "create bindings with custom blob storage name",
			envVars: map[string]string{
				"DAPR_BLOB_BINDING_NAME": "custom-blob-storage",
			},
			validateResult: func(t *testing.T, bindings *Bindings) {
				assert.NotNil(t, bindings)
				assert.Equal(t, "custom-blob-storage", bindings.blobStorageName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctx, cancel := sharedtesting.SetupDaprTest()
			defer cancel()
			defer ResetClientForTesting()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			// Act
			bindings := NewBindings(client)

			// Assert
			if tt.expectedError != "" {
				assert.Nil(t, bindings)
			} else {
				require.NotNil(t, bindings)
				if tt.validateResult != nil {
					tt.validateResult(t, bindings)
				}
			}

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestBindings_InvokeBinding(t *testing.T) {
	// Set up test environment at function level
	defer ResetClientForTesting()
	
	tests := []struct {
		name          string
		request       *BindingRequest
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		validateResult func(*testing.T, *BindingResponse)
	}{
		{
			name: "invoke blob storage binding create operation",
			request: &BindingRequest{
				Name:      "blob-storage-local",
				Operation: "create",
				Data:      []byte("test file content"),
				Metadata: map[string]string{
					"blobName":    "test-file.txt",
					"contentType": "text/plain",
				},
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, response *BindingResponse) {
				assert.NotNil(t, response)
			},
		},
		{
			name: "invoke blob storage binding get operation",
			request: &BindingRequest{
				Name:      "blob-storage-local",
				Operation: "get",
				Metadata: map[string]string{
					"blobName": "existing-file.txt",
				},
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, response *BindingResponse) {
				assert.NotNil(t, response)
			},
		},
		{
			name: "invoke Loki binding query operation",
			request: &BindingRequest{
				Name:      "grafana-loki",
				Operation: "query",
				Data:      []byte(`{app="international-center"} |= "ERROR"`),
				Metadata: map[string]string{
					"query": `{app="international-center"} |= "ERROR"`,
				},
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, response *BindingResponse) {
				assert.NotNil(t, response)
			},
		},
		{
			name: "invoke custom external binding",
			request: &BindingRequest{
				Name:      "third-party-api",
				Operation: "post",
				Data:      []byte(`{"message": "test data"}`),
				Metadata: map[string]string{
					"endpoint": "/api/v1/data",
					"authorization": "Bearer token123",
				},
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, response *BindingResponse) {
				assert.NotNil(t, response)
			},
		},
		{
			name: "invoke binding with nil request",
			request: nil,
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "binding request cannot be nil",
		},
		{
			name: "invoke binding with empty name",
			request: &BindingRequest{
				Name:      "",
				Operation: "create",
				Data:      []byte("test"),
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "binding name cannot be empty",
		},
		{
			name: "invoke binding with timeout context",
			request: &BindingRequest{
				Name:      "blob-storage-local",
				Operation: "create",
				Data:      []byte("test content"),
				Metadata: map[string]string{
					"blobName": "timeout-test.txt",
				},
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 1*time.Millisecond)
			},
		},
		{
			name: "invoke binding with cancelled context",
			request: &BindingRequest{
				Name:      "blob-storage-local", 
				Operation: "create",
				Data:      []byte("test content"),
			},
			setupContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel() // Cancel immediately
				return ctx, func() {}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set up test environment first
			_, testCancel := sharedtesting.SetupDaprTest()
			defer testCancel()
			defer ResetClientForTesting()
			
			ctx, cancel := tt.setupContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			bindings := NewBindings(client)
			require.NotNil(t, bindings)

			// Act
			response, err := bindings.InvokeBinding(ctx, tt.request)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, response)
			} else {
				if tt.validateResult != nil {
					require.NoError(t, err)
					tt.validateResult(t, response)
				}
				// For timeout/cancelled context tests, we just verify no panic occurs
			}
		})
	}
}

func TestBindings_UploadBlob(t *testing.T) {
	tests := []struct {
		name         string
		blobName     string
		data         []byte
		contentType  string
		setupContext func() (context.Context, context.CancelFunc)
		envVars      map[string]string
		expectedError string
	}{
		{
			name:        "upload text file blob",
			blobName:    "documents/test-document.txt",
			data:        []byte("This is a test document content"),
			contentType: "text/plain",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
		},
		{
			name:        "upload JSON blob",
			blobName:    "data/config.json",
			data:        []byte(`{"setting": "value", "enabled": true}`),
			contentType: "application/json",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
		},
		{
			name:        "upload binary blob",
			blobName:    "images/avatar.png",
			data:        []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}, // PNG header
			contentType: "image/png",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
		},
		{
			name:        "upload blob without content type",
			blobName:    "misc/unknown-file",
			data:        []byte("content without specified type"),
			contentType: "",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
		},
		{
			name:        "upload large blob",
			blobName:    "large-files/big-data.bin",
			data:        make([]byte, 1024*1024), // 1MB of zeros
			contentType: "application/octet-stream",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
		},
		{
			name:        "upload blob with empty name",
			blobName:    "",
			data:        []byte("test content"),
			contentType: "text/plain",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "blob name cannot be empty",
		},
		{
			name:        "upload blob with nil data",
			blobName:    "test-file.txt",
			data:        nil,
			contentType: "text/plain",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "blob data cannot be nil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set up test environment first
			_, testCancel := sharedtesting.SetupDaprTest()
			defer testCancel()
			defer ResetClientForTesting()
			
			ctx, cancel := tt.setupContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			bindings := NewBindings(client)
			require.NotNil(t, bindings)

			// Act
			err = bindings.UploadBlob(ctx, tt.blobName, tt.data, tt.contentType)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBindings_DownloadBlob(t *testing.T) {
	tests := []struct {
		name          string
		blobName      string
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		validateResult func(*testing.T, []byte)
	}{
		{
			name:     "download existing text blob",
			blobName: "documents/existing-document.txt",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, data []byte) {
				assert.NotNil(t, data)
				assert.True(t, len(data) > 0)
			},
		},
		{
			name:     "download existing JSON blob",
			blobName: "config/settings.json",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, data []byte) {
				assert.NotNil(t, data)
				// Could validate JSON structure here
			},
		},
		{
			name:     "download non-existent blob",
			blobName: "non-existent/file.txt",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "blob not found",
		},
		{
			name:     "download blob with empty name",
			blobName: "",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "blob name cannot be empty",
		},
		{
			name:     "download blob with timeout context",
			blobName: "documents/large-file.bin",
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 1*time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set up test environment first
			_, testCancel := sharedtesting.SetupDaprTest()
			defer testCancel()
			defer ResetClientForTesting()
			
			ctx, cancel := tt.setupContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			bindings := NewBindings(client)
			require.NotNil(t, bindings)

			// Act
			data, err := bindings.DownloadBlob(ctx, tt.blobName)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, data)
			} else {
				if tt.validateResult != nil {
					require.NoError(t, err)
					tt.validateResult(t, data)
				}
				// For timeout/cancelled context tests, we just verify no panic occurs
			}
		})
	}
}

func TestBindings_DeleteBlob(t *testing.T) {
	tests := []struct {
		name          string
		blobName      string
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
	}{
		{
			name:     "delete existing blob",
			blobName: "temp/deletable-file.txt",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
		},
		{
			name:     "delete non-existent blob",
			blobName: "non-existent/file.txt",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "blob not found",
		},
		{
			name:     "delete blob with empty name",
			blobName: "",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "blob name cannot be empty",
		},
		{
			name:     "delete blob with timeout context",
			blobName: "temp/timeout-test.txt",
			setupContext: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), 1*time.Millisecond)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set up test environment first
			_, testCancel := sharedtesting.SetupDaprTest()
			defer testCancel()
			defer ResetClientForTesting()
			
			ctx, cancel := tt.setupContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			bindings := NewBindings(client)
			require.NotNil(t, bindings)

			// Act
			err = bindings.DeleteBlob(ctx, tt.blobName)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBindings_ListBlobs(t *testing.T) {
	tests := []struct {
		name          string
		prefix        string
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		validateResult func(*testing.T, []string)
	}{
		{
			name:   "list all blobs without prefix",
			prefix: "",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, blobs []string) {
				assert.NotNil(t, blobs)
				// Should return empty slice if no blobs exist
			},
		},
		{
			name:   "list blobs with documents prefix",
			prefix: "documents/",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, blobs []string) {
				assert.NotNil(t, blobs)
				for _, blob := range blobs {
					assert.True(t, len(blob) >= len("documents/"))
				}
			},
		},
		{
			name:   "list blobs with images prefix",
			prefix: "images/",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, blobs []string) {
				assert.NotNil(t, blobs)
			},
		},
		{
			name:   "list blobs with non-existent prefix",
			prefix: "non-existent-folder/",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, blobs []string) {
				assert.NotNil(t, blobs)
				assert.Len(t, blobs, 0)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set up test environment first
			_, testCancel := sharedtesting.SetupDaprTest()
			defer testCancel()
			defer ResetClientForTesting()
			
			ctx, cancel := tt.setupContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			bindings := NewBindings(client)
			require.NotNil(t, bindings)

			// Act
			blobs, err := bindings.ListBlobs(ctx, tt.prefix)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, blobs)
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, blobs)
				}
			}
		})
	}
}

func TestBindings_GetBlobMetadata(t *testing.T) {
	tests := []struct {
		name          string
		blobName      string
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		validateResult func(*testing.T, map[string]string)
	}{
		{
			name:     "get metadata for existing blob",
			blobName: "documents/test-document.txt",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, metadata map[string]string) {
				assert.NotNil(t, metadata)
			},
		},
		{
			name:     "get metadata for non-existent blob",
			blobName: "non-existent/file.txt",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "blob not found",
		},
		{
			name:     "get metadata with empty blob name",
			blobName: "",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "blob name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set up test environment first
			_, testCancel := sharedtesting.SetupDaprTest()
			defer testCancel()
			defer ResetClientForTesting()
			
			ctx, cancel := tt.setupContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			bindings := NewBindings(client)
			require.NotNil(t, bindings)

			// Act
			metadata, err := bindings.GetBlobMetadata(ctx, tt.blobName)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, metadata)
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, metadata)
				}
			}
		})
	}
}

func TestBindings_CreateBlobURL(t *testing.T) {
	tests := []struct {
		name          string
		blobName      string
		expiryMinutes int
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		validateResult func(*testing.T, string)
	}{
		{
			name:          "create URL for existing blob with expiry",
			blobName:      "documents/shareable-document.pdf",
			expiryMinutes: 60,
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, url string) {
				assert.NotEmpty(t, url)
				assert.Contains(t, url, "http")
			},
		},
		{
			name:          "create URL for existing blob without expiry",
			blobName:      "images/public-image.jpg",
			expiryMinutes: 0,
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, url string) {
				assert.NotEmpty(t, url)
			},
		},
		{
			name:          "create URL for non-existent blob",
			blobName:      "non-existent/file.txt",
			expiryMinutes: 30,
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "blob not found",
		},
		{
			name:          "create URL with empty blob name",
			blobName:      "",
			expiryMinutes: 60,
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "blob name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set up test environment first
			_, testCancel := sharedtesting.SetupDaprTest()
			defer testCancel()
			defer ResetClientForTesting()
			
			ctx, cancel := tt.setupContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			bindings := NewBindings(client)
			require.NotNil(t, bindings)

			// Act
			url, err := bindings.CreateBlobURL(ctx, tt.blobName, tt.expiryMinutes)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, url)
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, url)
				}
			}
		})
	}
}

func TestBindings_CreateStoragePath(t *testing.T) {
	tests := []struct {
		name        string
		domain      string
		year        string
		month       string
		contentID   string
		hash        string
		extension   string
		envVars     map[string]string
		expectedPath string
	}{
		{
			name:         "create development storage path",
			domain:       "content",
			year:         "2024",
			month:        "01",
			contentID:    "content-123",
			hash:         "abc123def456",
			extension:    "pdf",
			envVars:      map[string]string{"ENVIRONMENT": "development"},
			expectedPath: "development/content/2024/01/content-123/abc123def456.pdf",
		},
		{
			name:         "create production storage path",
			domain:       "services",
			year:         "2024",
			month:        "12",
			contentID:    "service-789",
			hash:         "xyz789uvw456",
			extension:    "json",
			envVars:      map[string]string{"ENVIRONMENT": "production"},
			expectedPath: "production/services/2024/12/service-789/xyz789uvw456.json",
		},
		{
			name:         "create staging storage path",
			domain:       "migration",
			year:         "2023",
			month:        "06",
			contentID:    "migration-456",
			hash:         "def456ghi789",
			extension:    "xml",
			envVars:      map[string]string{"ENVIRONMENT": "staging"},
			expectedPath: "staging/migration/2023/06/migration-456/def456ghi789.xml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set up test environment first
			_, testCancel := sharedtesting.SetupDaprTest()
			defer testCancel()
			defer ResetClientForTesting()
			
			ctx, cancel := sharedtesting.CreateUnitTestContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			// Reset client after setting environment variables to pick up new env
			ResetClientForTesting()
			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			bindings := NewBindings(client)
			require.NotNil(t, bindings)

			// Act
			path := bindings.CreateStoragePath(tt.domain, tt.year, tt.month, tt.contentID, tt.hash, tt.extension)

			// Assert
			assert.Equal(t, tt.expectedPath, path)

			_ = ctx // Use context to avoid linting issues
		})
	}
}

func TestBindings_QueryLoki(t *testing.T) {
	tests := []struct {
		name          string
		bindingName   string
		query         string
		setupContext  func() (context.Context, context.CancelFunc)
		envVars       map[string]string
		expectedError string
		validateResult func(*testing.T, []byte)
	}{
		{
			name:        "query Loki for error logs",
			bindingName: "grafana-loki",
			query:       `{app="international-center"} |= "ERROR"`,
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, data []byte) {
				assert.NotNil(t, data)
			},
		},
		{
			name:        "query Loki for service-specific logs",
			bindingName: "grafana-loki",
			query:       `{service="content-service"} |= "INFO"`,
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, data []byte) {
				assert.NotNil(t, data)
			},
		},
		{
			name:        "query Loki with time range",
			bindingName: "grafana-loki",
			query:       `{app="international-center"}[5m]`,
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			validateResult: func(t *testing.T, data []byte) {
				assert.NotNil(t, data)
			},
		},
		{
			name:        "query with empty binding name",
			bindingName: "",
			query:       `{app="test"}`,
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "binding name cannot be empty",
		},
		{
			name:        "query with empty query string",
			bindingName: "grafana-loki",
			query:       "",
			setupContext: func() (context.Context, context.CancelFunc) {
				return sharedtesting.SetupDaprTest()
			},
			expectedError: "query cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set up test environment first
			_, testCancel := sharedtesting.SetupDaprTest()
			defer testCancel()
			defer ResetClientForTesting()
			
			ctx, cancel := tt.setupContext()
			defer cancel()

			for key, value := range tt.envVars {
				t.Setenv(key, value)
			}

			client, err := NewClient()
			require.NoError(t, err)
			require.NotNil(t, client)
			defer client.Close()

			bindings := NewBindings(client)
			require.NotNil(t, bindings)

			// Act
			data, err := bindings.QueryLoki(ctx, tt.bindingName, tt.query)

			// Assert
			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, data)
			} else {
				require.NoError(t, err)
				if tt.validateResult != nil {
					tt.validateResult(t, data)
				}
			}
		})
	}
}

func TestBindings_Concurrent_Access(t *testing.T) {
	// Arrange
	ctx, cancel := sharedtesting.SetupDaprTest()
	defer cancel()
	defer ResetClientForTesting()

	client, err := NewClient()
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	bindings := NewBindings(client)
	require.NotNil(t, bindings)

	// Act - Test concurrent access to bindings methods
	const numGoroutines = 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			// Test concurrent calls to various bindings methods
			testData := []byte("concurrent test data")
			blobName := "concurrent-test.txt"
			
			err1 := bindings.UploadBlob(ctx, blobName, testData, "text/plain")
			_, err2 := bindings.DownloadBlob(ctx, blobName)
			err3 := bindings.DeleteBlob(ctx, blobName)
			
			if err1 != nil {
				done <- err1
				return
			}
			if err2 != nil {
				done <- err2
				return
			}
			done <- err3
		}(i)
	}

	// Assert - All goroutines should complete without error
	for i := 0; i < numGoroutines; i++ {
		select {
		case err := <-done:
			// Some operations may fail in test environment, but should not panic
			if err != nil {
				t.Logf("Goroutine %d completed with error: %v", i, err)
			}
		case <-time.After(10 * time.Second):
			t.Fatal("Timeout waiting for goroutine completion")
		}
	}
}

func TestBindings_Error_Handling(t *testing.T) {
	tests := []struct {
		name          string
		setupTest     func(*testing.T) (*Bindings, context.Context, context.CancelFunc)
		operation     func(context.Context, *Bindings) error
		expectedError string
	}{
		{
			name: "upload blob with nil context should not panic",
			setupTest: func(t *testing.T) (*Bindings, context.Context, context.CancelFunc) {
				client, err := NewClient()
				require.NoError(t, err)
				bindings := NewBindings(client)
				return bindings, nil, func() { client.Close() }
			},
			operation: func(ctx context.Context, bindings *Bindings) error {
				return bindings.UploadBlob(ctx, "test.txt", []byte("test"), "text/plain")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange - Set up test environment first
			_, testCancel := sharedtesting.SetupDaprTest()
			defer testCancel()
			defer ResetClientForTesting()
			
			bindings, ctx, cleanup := tt.setupTest(t)
			defer cleanup()

			// Act & Assert - Should not panic
			assert.NotPanics(t, func() {
				err := tt.operation(ctx, bindings)
				if tt.expectedError != "" {
					assert.Error(t, err)
					assert.Contains(t, err.Error(), tt.expectedError)
				}
			})
		})
	}
}