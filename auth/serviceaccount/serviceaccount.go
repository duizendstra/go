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
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/duizendstra/go/logging/cloudrun"
	"golang.org/x/oauth2"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

// IAMServiceClient defines the interface for the IAM service operations.
type IAMServiceClient interface {
	SignJwt(ctx context.Context, name string, payload string) (*iam.SignJwtResponse, error)
}

// DefaultIAMServiceClient is an implementation of IAMServiceClient that talks to the real IAM service.
type DefaultIAMServiceClient struct{}

func (c *DefaultIAMServiceClient) SignJwt(ctx context.Context, name string, payload string) (*iam.SignJwtResponse, error) {
	iamService, err := iam.NewService(ctx, option.WithScopes(iam.CloudPlatformScope))
	if err != nil {
		return nil, err
	}
	return iamService.Projects.ServiceAccounts.SignJwt(name, &iam.SignJwtRequest{Payload: payload}).Context(ctx).Do()
}

// GenerateHTTPClient creates an authenticated HTTP client for GCP services.
func GenerateHTTPClient(ctx context.Context, logger *structured.StructuredLogger, iamClient IAMServiceClient, targetServiceAccount, userEmail, scopes string, tokenURL ...string) (*http.Client, error) {
	jwtAssertion, err := createJWTAssertion(targetServiceAccount, userEmail, scopes)
	if err != nil {
		logger.LogError(fmt.Sprintf("Error creating JWT assertion: %v", err))
		return nil, err
	}

	name := "projects/-/serviceAccounts/" + targetServiceAccount
	signJwtResponse, err := iamClient.SignJwt(ctx, name, jwtAssertion)
	if err != nil {
		logger.LogError(fmt.Sprintf("Error signing JWT: %v", err))
		return nil, err
	}

	tokenUrl := "https://oauth2.googleapis.com/token"
	if len(tokenURL) > 0 {
		tokenUrl = tokenURL[0]
	}
	
	data := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {signJwtResponse.SignedJwt},
	}
	resp, err := http.PostForm(tokenUrl, data)
	if err != nil {
		logger.LogError(fmt.Sprintf("Error posting to token endpoint: %v", err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, resp.Status)
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.LogError(fmt.Sprintf("Error reading response body: %v", err))
		return nil, err
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(responseBody, &tokenResponse); err != nil {
		logger.LogError(fmt.Sprintf("Error unmarshaling token response: %v", err))
		return nil, err
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: tokenResponse.AccessToken})
	return oauth2.NewClient(ctx, tokenSource), nil
}

// createJWTAssertion generates the JWT assertion string for the HTTP client.
func createJWTAssertion(targetServiceAccount, userEmail, scopes string) (string, error) {
	if targetServiceAccount == "" || userEmail == "" || scopes == "" {
		return "", fmt.Errorf("service account, user email, and scopes must all be provided")
	}

	return fmt.Sprintf(`{
        "iss": "%s",
        "sub": "%s",
        "scope": "%s",
        "aud": "https://oauth2.googleapis.com/token",
        "iat": %d
    }`, targetServiceAccount, userEmail, scopes, time.Now().Unix()), nil
}
