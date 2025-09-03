package auth

import (
	"context"
	"testing"
	"time"
)

func TestTokenValidator_ValidateToken(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name          string
		token         string
		expectedValid bool
		expectedError bool
		expectedUser  *UserInfo
	}{
		{
			name:          "valid admin JWT token",
			token:         "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJhZG1pbi11c2VyLTEyMyIsInJvbGVzIjpbImFkbWluIl0sImV4cCI6OTk5OTk5OTk5OX0",
			expectedValid: true,
			expectedError: false,
			expectedUser: &UserInfo{
				UserID: "admin-user-123",
				Roles:  []string{"admin"},
			},
		},
		{
			name:          "valid user JWT token", 
			token:         "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ1c2VyLTQ1NiIsInJvbGVzIjpbInVzZXIiXSwiZXhwIjo5OTk5OTk5OTk5fQ",
			expectedValid: true,
			expectedError: false,
			expectedUser: &UserInfo{
				UserID: "user-456",
				Roles:  []string{"user"},
			},
		},
		{
			name:          "expired token",
			token:         "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ1c2VyLTQ1NiIsInJvbGVzIjpbInVzZXIiXSwiZXhwIjoxfQ",
			expectedValid: false,
			expectedError: true,
			expectedUser:  nil,
		},
		{
			name:          "invalid token format",
			token:         "invalid.token.format",
			expectedValid: false,
			expectedError: true,
			expectedUser:  nil,
		},
		{
			name:          "empty token",
			token:         "",
			expectedValid: false,
			expectedError: true,
			expectedUser:  nil,
		},
	}

	validator := NewTestTokenValidator()
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateToken(ctx, tt.token)
			
			if tt.expectedError && err == nil {
				t.Errorf("ValidateToken() expected error but got none")
			}
			
			if !tt.expectedError && err != nil {
				t.Errorf("ValidateToken() unexpected error = %v", err)
			}
			
			if tt.expectedValid && !result.Valid {
				t.Errorf("ValidateToken() expected valid=true but got valid=%v", result.Valid)
			}
			
			if !tt.expectedValid && result.Valid {
				t.Errorf("ValidateToken() expected valid=false but got valid=%v", result.Valid)
			}
			
			if tt.expectedUser != nil {
				if result.UserInfo == nil {
					t.Errorf("ValidateToken() expected UserInfo but got nil")
				} else {
					if result.UserInfo.UserID != tt.expectedUser.UserID {
						t.Errorf("ValidateToken() UserID = %v, want %v", result.UserInfo.UserID, tt.expectedUser.UserID)
					}
					
					if len(result.UserInfo.Roles) != len(tt.expectedUser.Roles) {
						t.Errorf("ValidateToken() Roles length = %v, want %v", len(result.UserInfo.Roles), len(tt.expectedUser.Roles))
					}
				}
			}
		})
	}
}

func TestUserInfoProvider_GetUserInfo(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name         string
		userID       string
		expectedUser *UserInfo
		expectedErr  bool
	}{
		{
			name:   "get admin user info",
			userID: "admin-user-123",
			expectedUser: &UserInfo{
				UserID:      "admin-user-123",
				Username:    "admin",
				Email:       "admin@example.com",
				Roles:       []string{"admin"},
				Permissions: []string{"admin:*"},
			},
			expectedErr: false,
		},
		{
			name:   "get regular user info", 
			userID: "user-456",
			expectedUser: &UserInfo{
				UserID:      "user-456",
				Username:    "testuser",
				Email:       "testuser@example.com",
				Roles:       []string{"user"},
				Permissions: []string{"user:read"},
			},
			expectedErr: false,
		},
		{
			name:         "user not found",
			userID:       "nonexistent-user",
			expectedUser: nil,
			expectedErr:  true,
		},
	}

	provider := NewTestUserInfoProvider()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userInfo, err := provider.GetUserInfo(ctx, tt.userID)
			
			if tt.expectedErr && err == nil {
				t.Errorf("GetUserInfo() expected error but got none")
			}
			
			if !tt.expectedErr && err != nil {
				t.Errorf("GetUserInfo() unexpected error = %v", err)
			}
			
			if tt.expectedUser != nil && userInfo == nil {
				t.Errorf("GetUserInfo() expected user info but got nil")
			}
			
			if tt.expectedUser != nil && userInfo != nil {
				if userInfo.UserID != tt.expectedUser.UserID {
					t.Errorf("GetUserInfo() UserID = %v, want %v", userInfo.UserID, tt.expectedUser.UserID)
				}
				if userInfo.Username != tt.expectedUser.Username {
					t.Errorf("GetUserInfo() Username = %v, want %v", userInfo.Username, tt.expectedUser.Username)
				}
			}
		})
	}
}

func TestAuthenticationMiddleware_ProcessRequest(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tests := []struct {
		name            string
		authHeader      string
		expectedAuth    bool
		expectedUserID  string
		expectedRoles   []string
	}{
		{
			name:           "valid bearer token",
			authHeader:     "Bearer valid-jwt-token-admin",
			expectedAuth:   true,
			expectedUserID: "admin-user-123",
			expectedRoles:  []string{"admin"},
		},
		{
			name:           "invalid bearer token",
			authHeader:     "Bearer invalid-token",
			expectedAuth:   false,
			expectedUserID: "",
			expectedRoles:  nil,
		},
		{
			name:           "missing authorization header",
			authHeader:     "",
			expectedAuth:   false,
			expectedUserID: "",
			expectedRoles:  nil,
		},
		{
			name:           "wrong auth type",
			authHeader:     "Basic dXNlcjpwYXNz",
			expectedAuth:   false,
			expectedUserID: "",
			expectedRoles:  nil,
		},
	}

	middleware := NewTestAuthenticationMiddleware()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &AuthRequest{
				AuthorizationHeader: tt.authHeader,
				RequestPath:         "/api/v1/test",
				Method:             "GET",
			}

			result, err := middleware.ProcessRequest(ctx, request)
			if err != nil {
				t.Fatalf("ProcessRequest() error = %v", err)
			}

			if result.Authenticated != tt.expectedAuth {
				t.Errorf("ProcessRequest() Authenticated = %v, want %v", result.Authenticated, tt.expectedAuth)
			}

			if tt.expectedAuth {
				if result.UserInfo == nil {
					t.Errorf("ProcessRequest() expected UserInfo but got nil")
				} else {
					if result.UserInfo.UserID != tt.expectedUserID {
						t.Errorf("ProcessRequest() UserID = %v, want %v", result.UserInfo.UserID, tt.expectedUserID)
					}
				}
			}
		})
	}
}