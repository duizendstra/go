Here's a basic `README.md` for your `serviceaccount` package that explains the purpose, how to install, use, and run tests.

```markdown
# Google Service Account Authentication

This Go package simplifies the process of generating authenticated HTTP clients using Google Service Accounts. It handles signing JWTs, exchanging them for OAuth2 access tokens, and generating authenticated HTTP clients that can interact with Google Cloud services.

## Features
- Sign JWTs using Google IAM Service
- Generate OAuth2 tokens from signed JWTs
- Create authenticated HTTP clients for accessing Google services

## Installation

To use this package in your Go project, you can install it with `go get`:

```bash
go get github.com/duizendstra/go/google/auth/serviceaccount
```

Ensure that you have the required dependencies installed in your project:

```bash
go mod tidy
```

## Usage

### Generate an Authenticated HTTP Client

You can use this package to generate an authenticated HTTP client using a Google Service Account. Here's an example of how to use it:

```go
package main

import (
    "context"
    "fmt"
    "github.com/duizendstra/go/google/auth/serviceaccount"
    "github.com/duizendstra/go/google/logging"
)

func main() {
    ctx := context.Background()
    logger := logging.NewStructuredLogger("test-project", "test-component", nil, nil)
    iamClient := &serviceaccount.GoogleIAMServiceClient{}
    
    // Replace with your service account details
    serviceAccount := "your-service-account@your-project.iam.gserviceaccount.com"
    userEmail := "user@example.com"
    scopes := "https://www.googleapis.com/auth/cloud-platform"
    
    client, err := serviceaccount.GenerateGoogleHTTPClient(ctx, logger, iamClient, serviceAccount, userEmail, scopes)
    if err != nil {
        logger.LogError(ctx, "Failed to generate HTTP client", "error", err)
        return
    }

    fmt.Println("Successfully created authenticated HTTP client!")
}
```

### JWT Claims

When generating the signed JWT, the following claims are used:

- `iss`: The service account email address
- `sub`: The user email address
- `scope`: The scopes required for accessing Google Cloud resources
- `aud`: The audience, which is the Google token endpoint
- `iat`: The issued-at time (current time in seconds)
- `exp`: The expiration time (set to 1 hour after `iat`)

### Logging

The package uses structured logging with a logger interface, allowing you to log errors and events during the process of creating and signing JWTs, and exchanging tokens.

### Error Handling

The package includes error handling for different stages:
- Creating the JWT assertion
- Signing the JWT using Google IAM Service
- Exchanging the signed JWT for an access token
- Handling non-OK HTTP responses from the token endpoint

## Running Tests

To run the tests for this package, simply use the Go testing tool:

```bash
go test ./...
```

The tests include:
- Mocked IAM service client for JWT signing
- Mock HTTP server to simulate Google token endpoints
- Various test cases to cover valid and invalid input scenarios

## License

This project is licensed under the MIT License. See the LICENSE file for details.

---

For any issues or questions, feel free to contact [Jasper Duizendstra](mailto:jasper@duizendstra.com).
