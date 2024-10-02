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
	"github.com/duizendstra/go/google/errors"
	"github.com/duizendstra/go/google/logging"
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
    logger       *structured.StructuredLogger
}

// NewGoogleBaseServiceClient creates a new instance of GoogleBaseServiceClient
func NewGoogleBaseServiceClient(ctx context.Context, logger *structured.StructuredLogger, targetServiceAccount, userEmail, scopes, baseEndpoint string) (*GoogleBaseServiceClient, error) {
    httpClient, err := serviceaccount.GenerateHTTPClient(ctx, logger, &serviceaccount.DefaultIAMServiceClient{}, targetServiceAccount, userEmail, scopes)
    if err != nil {
        if strings.Contains(err.Error(), "Gaia id not found for email") {
            apiErr := &errors.APIError{
                StatusCode:   http.StatusNotFound,
                Body:         fmt.Sprintf("Gaia ID not found for email %s: %v", targetServiceAccount, err),
                ErrorCode:    "1000",
                ErrorMessage: fmt.Sprintf("Gaia ID not found for email %s", targetServiceAccount),
            }
            logger.LogError(apiErr.Error())
            return nil, apiErr
        }
        apiErr := &errors.APIError{
            StatusCode: http.StatusInternalServerError,
            Body:       fmt.Sprintf("Error generating HTTP client: %v", err),
        }
        logger.LogError(apiErr.Error())
        return nil, apiErr
    }
    return &GoogleBaseServiceClient{
        httpClient:   httpClient,
        baseEndpoint: baseEndpoint,
        logger:       logger,
    }, nil
}

// makeRequest executes an HTTP GET request to the specified endpoint with given parameters
func (c *GoogleBaseServiceClient) makeRequest(ctx context.Context, endpoint string, params url.Values) ([]byte, error) {
    reqURL := fmt.Sprintf("%s/%s?%s", c.baseEndpoint, endpoint, params.Encode())

    req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
    if err != nil {
        return nil, fmt.Errorf("error creating request: %w", err)
    }

    // Extract the token from the HTTP client's transport
    tokenSource := c.httpClient.Transport.(*oauth2.Transport).Source
    token, err := tokenSource.Token()
    if err != nil {
        return nil, fmt.Errorf("error getting token: %w", err)
    }

    // Set the Authorization header with the Bearer token
    req.Header.Set("Authorization", "Bearer "+token.AccessToken)
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error making API call: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
    }

    return io.ReadAll(resp.Body)
}

// makePostRequest executes an HTTP POST request to the specified endpoint with given headers and body.
func (c *GoogleBaseServiceClient) makePostRequest(ctx context.Context, endpoint string, headers map[string]string, body []byte) ([]byte, error) {
    reqURL := fmt.Sprintf("%s/%s", c.baseEndpoint, endpoint)

    req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewReader(body))
    if err != nil {
        return nil, fmt.Errorf("error creating POST request: %w", err)
    }

    // Set headers
    for key, value := range headers {
        req.Header.Set(key, value)
    }

    // Execute the request
    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("error making API call: %w", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode < 200 || resp.StatusCode >= 300 {
        bodyBytes, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
    }

    return io.ReadAll(resp.Body)
}
