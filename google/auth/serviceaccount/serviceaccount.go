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

	"github.com/duizendstra/go/google/errors"
	"github.com/duizendstra/go/google/structuredlogger"
	"golang.org/x/oauth2"
	"google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

// IAMServiceClient defines the interface for the IAM service operations.
type IAMServiceClient interface {
	SignJwt(ctx context.Context, name string, payload string) (*iam.SignJwtResponse, error)
}

// GoogleIAMServiceClient is an implementation of IAMServiceClient that talks to the real IAM service.
type GoogleIAMServiceClient struct{}

// SignJwt creates a signed JWT by calling Google's IAM service.
func (c *GoogleIAMServiceClient) SignJwt(ctx context.Context, name string, payload string) (*iam.SignJwtResponse, error) {
	iamService, err := iam.NewService(ctx, option.WithScopes(iam.CloudPlatformScope))
	if err != nil {
		return nil, fmt.Errorf("failed to initialize IAM service: %w", err)
	}
	return iamService.Projects.ServiceAccounts.SignJwt(name, &iam.SignJwtRequest{Payload: payload}).Context(ctx).Do()
}

// JWTClaims represents the claims needed for creating a JWT assertion.
type JWTClaims struct {
	Iss   string `json:"iss"`
	Sub   string `json:"sub"`
	Scope string `json:"scope"`
	Aud   string `json:"aud"`
	Iat   int64  `json:"iat"`
	Exp   int64  `json:"exp"`
}

// GenerateGoogleHTTPClient creates an authenticated HTTP client for GCP services.
func GenerateGoogleHTTPClient(ctx context.Context, logger *structuredlogger.StructuredLogger, iamClient IAMServiceClient, targetServiceAccount, userEmail, scopes string, tokenURL ...string) (*http.Client, error) {
	jwtAssertion, err := createJWTAssertion(targetServiceAccount, userEmail, scopes)
	if err != nil {
		logger.LogError(ctx, "Error creating JWT assertion", "error", err)
		return nil, fmt.Errorf("error creating JWT assertion: %w", err)
	}

	name := "projects/-/serviceAccounts/" + targetServiceAccount
	signJwtResponse, err := iamClient.SignJwt(ctx, name, jwtAssertion)
	if err != nil {
		logger.LogError(ctx, "Error signing JWT", "error", err)
		return nil, fmt.Errorf("error signing JWT: %w", err)
	}
	
	tokenUrl := "https://oauth2.googleapis.com/token"
	if len(tokenURL) > 0 {
		tokenUrl = tokenURL[0]
	}
	
	accessToken, err := getAccessToken(logger, tokenUrl, signJwtResponse.SignedJwt)
	if err != nil {
		logger.LogError(ctx, "Error getting access token", "error", err)
		return nil, err
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	return oauth2.NewClient(ctx, tokenSource), nil
}

// createJWTAssertion generates the JWT assertion string for the HTTP client.
func createJWTAssertion(targetServiceAccount, userEmail, scopes string) (string, error) {
	if targetServiceAccount == "" || userEmail == "" || scopes == "" {
		return "", fmt.Errorf("service account, user email, and scopes must all be provided")
	}

	now := time.Now().Unix()
	jwtPayload := JWTClaims{
		Iss:   targetServiceAccount,
		Sub:   userEmail,
		Scope: scopes,
		Aud:   "https://oauth2.googleapis.com/token",
		Iat:   now,
		Exp:   now + 3600, // Token expiration time (1 hour)
	}

	payloadBytes, err := json.Marshal(jwtPayload)
	if err != nil {
		return "", fmt.Errorf("error marshaling JWT payload: %w", err)
	}

	return string(payloadBytes), nil
}

// getAccessToken exchanges the signed JWT for an access token.
func getAccessToken(logger *structuredlogger.StructuredLogger, tokenUrl, signedJwt string) (string, error) {
	data := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {signedJwt},  // Ensure the signed JWT is being passed here
	}
	
	resp, err := http.PostForm(tokenUrl, data)
	if err != nil {
		logger.LogError(context.Background(), "Error posting to token URL", "url", tokenUrl, "error", err)
		return "", fmt.Errorf("error posting to token endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logger.LogError(context.Background(), "Non-OK response from token URL", "status", resp.StatusCode, "body", string(body))
		return "", &errors.GoogleAPIError{
			StatusCode: resp.StatusCode,
			Body:       string(body),
		}
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		logger.LogError(context.Background(), "Error decoding access token response", "error", err)
		return "", fmt.Errorf("error unmarshaling token response: %w", err)
	}

	return tokenResponse.AccessToken, nil
}
