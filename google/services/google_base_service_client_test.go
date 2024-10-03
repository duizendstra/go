package googleclient

import (

	"context"
	"encoding/json"

	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/duizendstra/go/google/structuredlogger"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2"
)

type MockTokenSource struct{}

func (mts *MockTokenSource) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: "mocked_access_token",
	}, nil
}

func TestMakeRequest(t *testing.T) {
	logger := structuredlogger.NewStructuredLogger("test-project", "test-component", nil, nil)
	client := &GoogleBaseServiceClient{
		httpClient: &http.Client{
			Transport: &oauth2.Transport{
				Source: &MockTokenSource{},
			},
		},
		baseEndpoint: "http://example.com",
		logger:       logger,
	}

	// Create a test server that will mock the API response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer mocked_access_token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer ts.Close()

	client.baseEndpoint = ts.URL

	// Create request parameters
	params := url.Values{}
	params.Add("key", "value")

	// Execute the GET request
	response, err := client.MakeRequest(context.Background(), "test-endpoint", params)
	assert.NoError(t, err)

	// Validate the response
	var jsonResponse map[string]string
	err = json.Unmarshal(response, &jsonResponse)
	assert.NoError(t, err)
	assert.Equal(t, "success", jsonResponse["message"])
}

func TestMakePostRequest(t *testing.T) {
	logger := structuredlogger.NewStructuredLogger("test-project", "test-component", nil, nil)
	client := &GoogleBaseServiceClient{
		httpClient: &http.Client{
			Transport: &oauth2.Transport{
				Source: &MockTokenSource{},
			},
		},
		baseEndpoint: "http://example.com",
		logger:       logger,
	}

	// Create a test server that will mock the API response
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer mocked_access_token", r.Header.Get("Authorization"))
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var requestBody map[string]string
		err := json.NewDecoder(r.Body).Decode(&requestBody)
		assert.NoError(t, err)
		assert.Equal(t, "test_value", requestBody["test_key"])

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "post success"}`))
	}))
	defer ts.Close()

	client.baseEndpoint = ts.URL

	// Create request body
	body := map[string]string{"test_key": "test_value"}
	bodyBytes, _ := json.Marshal(body)

	// Create headers
	headers := map[string]string{
		"Content-Type": "application/json",
	}

	// Execute the POST request
	response, err := client.MakePostRequest(context.Background(), "test-post-endpoint", headers, bodyBytes)
	assert.NoError(t, err)

	// Validate the response
	var jsonResponse map[string]string
	err = json.Unmarshal(response, &jsonResponse)
	assert.NoError(t, err)
	assert.Equal(t, "post success", jsonResponse["message"])
}
