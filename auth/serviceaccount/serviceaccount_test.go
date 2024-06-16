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
