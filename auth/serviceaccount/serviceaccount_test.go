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

package serviceaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/duizendstra/go/logging/cloudrun"
	"google.golang.org/api/iam/v1"
)

// MockIAMServiceClient is a mock implementation of IAMServiceClient
type MockIAMServiceClient struct{}

func (m *MockIAMServiceClient) SignJwt(ctx context.Context, name string, payload string) (*iam.SignJwtResponse, error) {
	return &iam.SignJwtResponse{SignedJwt: "mocked_signed_jwt"}, nil
}

func TestGenerateHTTPClient(t *testing.T) {
	logger := structured.NewStructuredLogger("test-project", "test-component", nil)
	mockIAMClient := &MockIAMServiceClient{}

	// Create a test HTTP server to mock the OAuth token endpoint
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("Mock server received request")
		err := r.ParseForm()
		if err != nil {
			log.Printf("Error parsing form: %v", err)
			http.Error(w, fmt.Sprintf("Error parsing form: %v", err), http.StatusBadRequest)
			return
		}
		log.Printf("Form data: %v", r.Form)
		if r.Form.Get("assertion") != "mocked_signed_jwt" {
			log.Printf("Expected assertion 'mocked_signed_jwt', got '%s'", r.Form.Get("assertion"))
			http.Error(w, fmt.Sprintf("Expected assertion 'mocked_signed_jwt', got '%s'", r.Form.Get("assertion")), http.StatusBadRequest)
			return
		}
		resp := map[string]string{"access_token": "mocked_access_token"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	ctx := context.Background()
	client, err := GenerateHTTPClient(ctx, logger, mockIAMClient, "test-service-account", "test-user@example.com", "test-scope", ts.URL)
	if err != nil {
		t.Fatalf("GenerateHTTPClient returned error: %v", err)
	}

	// Make a request using the generated HTTP client
	req, err := http.NewRequest("GET", ts.URL, nil)
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

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		t.Fatalf("Error decoding response: %v", err)
	}

	expectedAccessToken := "mocked_access_token"
	if tokenResponse.AccessToken != expectedAccessToken {
		t.Errorf("Expected access token '%s', got '%s'", expectedAccessToken, tokenResponse.AccessToken)
	}
}
