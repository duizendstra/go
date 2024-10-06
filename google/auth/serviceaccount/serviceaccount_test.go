// Copyright 2024 Jasper Duizendstra
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package serviceaccount_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"google.golang.org/api/iam/v1"
	"github.com/duizendstra/go/google/auth/serviceaccount"
	structuredlogger "github.com/duizendstra/go/google/structuredlogger"
)

// MockIAMServiceClient is a mock implementation of IAMServiceClient
type MockIAMServiceClient struct{}

func (m *MockIAMServiceClient) SignJwt(ctx context.Context, name string, payload string) (*iam.SignJwtResponse, error) {
	// Return a mocked signed JWT
	return &iam.SignJwtResponse{SignedJwt: "mocked_signed_jwt"}, nil
}

func TestGenerateGoogleHTTPClient(t *testing.T) {
	mockIAMClient := &MockIAMServiceClient{}

	// Create a valid logger instance instead of passing nil
	logger := structuredlogger.NewStructuredLogger("test-project", "test-component", nil, nil)

	// Create a TokenCache instance
	tokenCache := serviceaccount.NewTokenCache()

	// Access token to be used by the GET request
	expectedAccessToken := "mocked_access_token"

	// Create a test HTTP server to mock the OAuth token endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			err := r.ParseForm()
			if err != nil {
				http.Error(w, fmt.Sprintf("Error parsing form: %v", err), http.StatusBadRequest)
				return
			}

			// Check if the assertion matches the mocked signed JWT from the mock IAM client
			if r.Form.Get("assertion") != "mocked_signed_jwt" {
				http.Error(w, fmt.Sprintf("Expected assertion 'mocked_signed_jwt', got '%s'", r.Form.Get("assertion")), http.StatusBadRequest)
				return
			}

			// Respond with a mock access token
			resp := map[string]string{"access_token": expectedAccessToken}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		} else if r.Method == "GET" {
			// Check for the Authorization header in the GET request
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer "+expectedAccessToken {
				http.Error(w, "Missing or incorrect Authorization header", http.StatusUnauthorized)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"message": "Authorized"}`))
		}
	}))
	defer ts.Close()

	// Test cases
	tests := []struct {
		name             string
		targetServiceAcc string
		userEmail        string
		scopes           string
		tokenURL         string
		expectedErr      string
	}{
		{
			name:             "Valid inputs",
			targetServiceAcc: "test-service-account",
			userEmail:        "test-user@example.com",
			scopes:           "test-scope",
			tokenURL:         ts.URL,
			expectedErr:      "",
		},
		// {
		// 	name:             "Invalid target service account",
		// 	targetServiceAcc: "",
		// 	userEmail:        "test-user@example.com",
		// 	scopes:           "test-scope",
		// 	tokenURL:         ts.URL,
		// 	expectedErr:      "error creating JWT assertion: service account, user email, and scopes must all be provided",
		// },
		// {
		// 	name:             "Invalid user email",
		// 	targetServiceAcc: "test-service-account",
		// 	userEmail:        "",
		// 	scopes:           "test-scope",
		// 	tokenURL:         ts.URL,
		// 	expectedErr:      "error creating JWT assertion: service account, user email, and scopes must all be provided",
		// },
		// {
		// 	name:             "Invalid scopes",
		// 	targetServiceAcc: "test-service-account",
		// 	userEmail:        "test-user@example.com",
		// 	scopes:           "",
		// 	tokenURL:         ts.URL,
		// 	expectedErr:      "error creating JWT assertion: service account, user email, and scopes must all be provided",
		// },
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			client, err := serviceaccount.GenerateGoogleHTTPClient(ctx, logger, mockIAMClient, tokenCache, tt.targetServiceAcc, tt.userEmail, tt.scopes, tt.tokenURL)
			if err != nil {
				if tt.expectedErr == "" {
					t.Fatalf("GenerateGoogleHTTPClient returned unexpected error: %v", err)
				}
				if err.Error() != tt.expectedErr {
					t.Fatalf("Expected error: %v, got: %v", tt.expectedErr, err.Error())
				}
				return
			}

			if tt.expectedErr != "" {
				t.Fatalf("Expected error: %v, got none", tt.expectedErr)
			}

			// Make a GET request using the generated HTTP client
			req, err := http.NewRequest("GET", tt.tokenURL, nil)
			if err != nil {
				t.Fatalf("Error creating request: %v", err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("HTTP client returned error: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("Expected status code 200, got %d", resp.StatusCode)
			}

			// Validate that the GET request is successful
			var responseBody map[string]string
			if err := json.NewDecoder(resp.Body).Decode(&responseBody); err != nil {
				t.Fatalf("Error decoding response: %v", err)
			}

			if responseBody["message"] != "Authorized" {
				t.Errorf("Expected message 'Authorized', got '%s'", responseBody["message"])
			}
		})
	}
}
