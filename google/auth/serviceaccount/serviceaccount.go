
package googleserviceaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/duizendstra/go/google/errors"
	"github.com/duizendstra/go/google/logging"
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

func (c *GoogleIAMServiceClient) SignJwt(ctx context.Context, name string, payload string) (*iam.SignJwtResponse, error) {
	iamService, err := iam.NewService(ctx, option.WithScopes(iam.CloudPlatformScope))
	if err != nil {
		return nil, err
	}
	return iamService.Projects.ServiceAccounts.SignJwt(name, &iam.SignJwtRequest{Payload: payload}).Context(ctx).Do()
}

// GenerateGoogleHTTPClient creates an authenticated HTTP client for GCP services.
func GenerateGoogleHTTPClient(ctx context.Context, logger *structured.StructuredLogger, iamClient IAMServiceClient, targetServiceAccount, userEmail, scopes string, tokenURL ...string) (*http.Client, error) {
	jwtAssertion, err := createJWTAssertion(targetServiceAccount, userEmail, scopes)
	if err != nil {
		return nil, fmt.Errorf("error creating JWT assertion: %w", err)
	}

	name := "projects/-/serviceAccounts/" + targetServiceAccount
	signJwtResponse, err := iamClient.SignJwt(ctx, name, jwtAssertion)
	if err != nil {
		return nil, fmt.Errorf("error signing JWT: %w", err)
	}

	tokenUrl := "https://oauth2.googleapis.com/token"
	if len(tokenURL) > 0 {
		tokenUrl = tokenURL[0]
	}

	accessToken, err := getAccessToken(tokenUrl, signJwtResponse.SignedJwt)
	if err != nil {
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
	return fmt.Sprintf(`{
		"iss": "%s",
		"sub": "%s",
		"scope": "%s",
		"aud": "https://oauth2.googleapis.com/token",
		"iat": %d,
		"exp": %d
	}`, targetServiceAccount, userEmail, scopes, now, now+3600), nil
}

// getAccessToken exchanges the signed JWT for an access token.
func getAccessToken(tokenUrl, signedJwt string) (string, error) {
	data := url.Values{
		"grant_type": {"urn:ietf:params:oauth:grant-type:jwt-bearer"},
		"assertion":  {signedJwt},
	}
	resp, err := http.PostForm(tokenUrl, data)
	if err != nil {
		return "", fmt.Errorf("error posting to token endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", &errors.GoogleAPIError{
			StatusCode: resp.StatusCode,
			Body:       string(body),
		}
	}

	var tokenResponse struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tokenResponse); err != nil {
		return "", fmt.Errorf("error unmarshaling token response: %w", err)
	}

	return tokenResponse.AccessToken, nil
}
