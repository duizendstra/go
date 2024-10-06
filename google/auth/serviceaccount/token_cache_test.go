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

package serviceaccount_test

import (
	"testing"
	"time"

	"github.com/duizendstra/go/google/auth/serviceaccount"
	"golang.org/x/oauth2"
	"github.com/stretchr/testify/assert"
)

func TestTokenCache_SetGetToken(t *testing.T) {
	tokenCache := serviceaccount.NewTokenCache()
	userEmail := "test@example.com"
	scopes := []string{"scope1", "scope2"}
	token := &oauth2.Token{
		AccessToken: "test-access-token",
		Expiry:      time.Now().Add(time.Hour),
	}

	// Set token in cache
	tokenCache.SetToken(userEmail, scopes, token)

	// Get token from cache
	retrievedToken, exists := tokenCache.GetToken(userEmail, scopes)

	// Assert that the token was successfully retrieved and is equal to the original
	assert.True(t, exists, "Expected token to exist in the cache")
	assert.Equal(t, token, retrievedToken, "Expected retrieved token to match the original token")
}

// func TestTokenCache_ExpiredToken(t *testing.T) {
// 	tokenCache := serviceaccount.NewTokenCache()
// 	userEmail := "test@example.com"
// 	scopes := []string{"scope1", "scope2"}
// 	token := &oauth2.Token{
// 		AccessToken: "expired-access-token",
// 		Expiry:      time.Now().Add(-time.Hour),
// 	}

// 	// Set expired token in cache
// 	tokenCache.SetToken(userEmail, scopes, token)

// 	// Try to get the expired token from cache
// 	retrievedToken, exists := tokenCache.GetToken(userEmail, scopes)

// 	// Assert that the token was not retrieved as it is expired
// 	assert.False(t, exists, "Expected token to not exist in the cache since it is expired")
// 	assert.Nil(t, retrievedToken, "Expected retrieved token to be nil since it is expired")
// }

// func TestTokenCache_ConcurrentAccess(t *testing.T) {
// 	tokenCache := serviceaccount.NewTokenCache()
// 	userEmail := "test@example.com"
// 	scopes := []string{"scope1", "scope2"}
// 	token := &oauth2.Token{
// 		AccessToken: "test-access-token",
// 		Expiry:      time.Now().Add(time.Hour),
// 	}

// 	// Set token in cache
// 	tokenCache.SetToken(userEmail, scopes, token)

// 	done := make(chan bool)

// 	// Simulate concurrent access
// 	for i := 0; i < 10; i++ {
// 		go func() {
// 			retrievedToken, exists := tokenCache.GetToken(userEmail, scopes)
// 			assert.True(t, exists, "Expected token to exist in the cache during concurrent access")
// 			assert.Equal(t, token, retrievedToken, "Expected retrieved token to match the original token during concurrent access")
// 			done <- true
// 		}()
// 	}

// 	// Wait for all goroutines to finish
// 	for i := 0; i < 10; i++ {
// 		<-done
// 	}
// }