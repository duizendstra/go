package serviceaccount

import (
	"fmt"
	"strings"
	"sync"

	"golang.org/x/oauth2"
)

type TokenCache struct {
	mu     sync.Mutex
	tokens map[string]*oauth2.Token
}

// NewTokenCache creates a new TokenCache.
func NewTokenCache() *TokenCache {
	return &TokenCache{
		tokens: make(map[string]*oauth2.Token),
	}
}

// getKey generates a unique cache key using userEmail and scopes.
func getKey(userEmail string, scopes []string) string {
	return fmt.Sprintf("%s|%s", userEmail, strings.Join(scopes, ","))
}

// GetToken retrieves the cached token for the given userEmail and scopes.
// It also checks whether the token is expired before returning.
func (c *TokenCache) GetToken(userEmail string, scopes []string) (*oauth2.Token, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := getKey(userEmail, scopes)
	token, exists := c.tokens[key]
	if !exists || token == nil {
		// Token not found in cache
		return nil, false
	}

	if !token.Valid() {
		// Token is expired
		delete(c.tokens, key) // Remove expired token from cache
		return nil, false
	}

	// Return valid token
	return token, true
}

// SetToken caches the token for the given userEmail and scopes.
func (c *TokenCache) SetToken(userEmail string, scopes []string, token *oauth2.Token) {
	c.mu.Lock()
	defer c.mu.Unlock()

	key := getKey(userEmail, scopes)
	c.tokens[key] = token
}
