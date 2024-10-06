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
	"fmt"
	"hash/fnv"
	"strings"
	"sync"

	"golang.org/x/oauth2"
)

type TokenCache struct {
	mu     sync.RWMutex
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
    key := fmt.Sprintf("%s|%s", userEmail, strings.Join(scopes, ","))
    hasher := fnv.New64a()
    hasher.Write([]byte(key))
    return fmt.Sprintf("%x", hasher.Sum64())
}

// GetToken retrieves the cached token for the given userEmail and scopes.
// It also checks whether the token is expired before returning.
func (c *TokenCache) GetToken(userEmail string, scopes []string) (*oauth2.Token, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	key := getKey(userEmail, scopes)
	token, exists := c.tokens[key]
	if !exists {
		// Token not found in cache
		return nil, false
	}

	if !token.Valid() {
		// Token is expired
		c.mu.Lock()
		delete(c.tokens, key) // Remove expired token from cache
		c.mu.Unlock()
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