package authorization

import (
	"fmt"
	"net/http"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
)

const (
	// This should be read from a secret/secret key and be given by a provider
	jwtKey    = "Zq4t7w!z%C*F-J@NcRfUjXn2r5u8x/A?D(G+KbPdSgVkYp3s6v9y$B&E)H@McQfThWmZq4t7w!z%C*F-JaNdRgUkXn2r5u8x/A?D(G+KbPeShVmYq3s6v9y$B&E)H@Mc"
	userIDKey = "user-id"
	authToken = "auth-token"
)

var blackListedTokens = map[string]struct{}{}

// CleanBlacklist starts a go routine that periodically clears the black list
func CleanBlacklist(cleanInterval time.Duration) {
	go func() {
		timer := time.NewTimer(cleanInterval)
		for range timer.C {
			clearExpiredFromBlacklist()
		}
	}()
}

// Claim uses the standard JWT Claim to create a custom claim
type Claim struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

// GenerateJWTToken generates a JWT token for the provided username
func GenerateJWTToken(username string) (string, time.Time, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claim{
		Username: username,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtKey))
	return tokenString, expirationTime, err
}

// ValidateToken validates whether the value of the receoved key is a valid token
func ValidateToken(tknStr string) (bool, *Claim, error) {
	claim := &Claim{}
	tkn, err := jwt.ParseWithClaims(tknStr, claim, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtKey), nil
	})
	if err != nil {
		return false, nil, err
	}
	return tkn.Valid, claim, nil
}

// AddUserIDHeader adds a header to the request that represents the user ID
func AddUserIDHeader(r *http.Request, claim *Claim) {
	r.Header.Add(userIDKey, fmt.Sprint(claim.Username))
}

// RemoveUserIDHeader removes the header that contains the user ID
func RemoveUserIDHeader(r *http.Request) {
	r.Header.Del(userIDKey)
}

// GetUserIDFromRequest returns the user ID from the request
func GetUserIDFromRequest(r *http.Request) (string, error) {
	id := r.Header.Get(userIDKey)
	if id == "" {
		return "", fmt.Errorf("request is missing user ID")
	}
	return id, nil
}

// SetAuthCookie add an auth cookie to the response writer
func SetAuthCookie(w http.ResponseWriter, tokenString string, expires time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:    authToken,
		Value:   tokenString,
		Expires: expires,
	})
}

// GetAuthToken extracts the auth token from a request
func GetAuthToken(r *http.Request) (string, error) {
	c, err := r.Cookie(authToken)
	if err != nil {
		return "", err
	}
	return c.Value, nil
}

// BlacklistToken blacklists a token so that it cannot be used anymore
func BlacklistToken(token string) {
	blackListedTokens[token] = struct{}{}
}

// IsBlacklisted checks if a token is blacklisted
func IsBlacklisted(token string) bool {
	_, ok := blackListedTokens[token]
	return ok
}

// clearExpiredFromBlacklist clears any invalid token from the blacklist
func clearExpiredFromBlacklist() {
	validTokens := map[string]struct{}{}
	for token := range blackListedTokens {
		valid, _, err := ValidateToken(token)
		if valid && err == nil {
			validTokens[token] = struct{}{}
		}
	}
	blackListedTokens = validTokens
}
