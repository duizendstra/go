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
package googleclient

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/duizendstra/go/google/auth/serviceaccount"
	apierrors "github.com/duizendstra/go/google/apierrors"
	"github.com/duizendstra/go/google/structuredlogger"
	"golang.org/x/oauth2"
)

type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API request failed with status %d: %s", e.StatusCode, e.Body)
}

type GoogleBaseServiceClient struct {
	httpClient   *http.Client
	baseEndpoint string
	logger       *structuredlogger.StructuredLogger
}

// NewGoogleBaseServiceClient creates a new instance of GoogleBaseServiceClient
func NewGoogleBaseServiceClient(ctx context.Context, logger *structuredlogger.StructuredLogger, tokenCache *serviceaccount.TokenCache, targetServiceAccount, userEmail, scopes, baseEndpoint string) (*GoogleBaseServiceClient, error) {
	// Assuming you have a TokenCache instance created as tokenCache
httpClient, err := serviceaccount.GenerateGoogleHTTPClient(ctx, logger, &serviceaccount.GoogleIAMServiceClient{}, tokenCache, targetServiceAccount, userEmail, scopes)

	if err != nil {
		if strings.Contains(err.Error(), "Gaia id not found for email") {
			apiErr := &apierrors.GoogleAPIError{
				StatusCode:   http.StatusNotFound,
				Body:         fmt.Sprintf("Gaia ID not found for email %s: %v", targetServiceAccount, err),
				ErrorCode:    "1000",
				ErrorMessage: fmt.Sprintf("Gaia ID not found for email %s", targetServiceAccount),
			}
			logger.LogError(ctx, apiErr.Error(), "email", targetServiceAccount)
			return nil, apiErr
		}

		apiErr := &apierrors.GoogleAPIError{
			StatusCode: http.StatusInternalServerError,
			Body:       fmt.Sprintf("Error generating HTTP client: %v", err),
		}
		logger.LogError(ctx, apiErr.Error(), "error", err)

		return nil, apiErr
	}
	return &GoogleBaseServiceClient{
		httpClient:   httpClient,
		baseEndpoint: baseEndpoint,
		logger:       logger,
	}, nil
}

// makeRequest executes an HTTP GET request to the specified endpoint with given parameters
func (c *GoogleBaseServiceClient) MakeRequest(ctx context.Context, endpoint string, params url.Values) ([]byte, error) {
	reqURL := fmt.Sprintf("%s/%s?%s", c.baseEndpoint, endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		c.logger.LogError(ctx, "error creating request", "error", err)
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Extract the token from the HTTP client's transport
	tokenSource := c.httpClient.Transport.(*oauth2.Transport).Source
	token, err := tokenSource.Token()
	if err != nil {
		c.logger.LogError(ctx, "error getting token", "error", err)
		return nil, fmt.Errorf("error getting token: %w", err)
	}

	// Set the Authorization header with the Bearer token
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.LogError(ctx, "error making API call", "error", err)
		return nil, fmt.Errorf("error making API call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		c.logger.LogError(ctx, "API request failed", "status_code", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	return io.ReadAll(resp.Body)
}

// makePostRequest executes an HTTP POST request to the specified endpoint with given headers and body.
func (c *GoogleBaseServiceClient) MakePostRequest(ctx context.Context, endpoint string, headers map[string]string, body []byte) ([]byte, error) {
	reqURL := fmt.Sprintf("%s/%s", c.baseEndpoint, endpoint)

	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewReader(body))
	if err != nil {
		c.logger.LogError(ctx, "error creating POST request", "error", err)
		return nil, fmt.Errorf("error creating POST request: %w", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Execute the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.LogError(ctx, "error making API call", "error", err)
		return nil, fmt.Errorf("error making API call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.logger.LogError(ctx, "API request failed", "status_code", resp.StatusCode, "body", string(bodyBytes))
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return io.ReadAll(resp.Body)
}
