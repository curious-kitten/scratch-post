package auth

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v4"
)

const (
	userIDKey = "user-id"
	authToken = "auth-token"
)

// Cleanup starts a go routine that periodically clears the black list
func (j *JWT) Cleanup(cleanInterval time.Duration) {
	go func() {
		timer := time.NewTimer(cleanInterval)
		for range timer.C {
			j.clearExpiredFromBlacklist()
		}
	}()
}

// Claim uses the standard JWT Claim to create a custom claim
type Claim struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// Keys is the interface that needs to be implemented to retrieve the key needed to sign a token
type Keys interface {
	GetOne() ([]byte, error)
}

// NewJWTHandler is used to setup a JWT authorizer
func NewJWTHandler(keys Keys) *JWT {
	return &JWT{
		keys:              keys,
		blackListedTokens: map[string]struct{}{},
	}
}

// JWT is used to generate and verify JST tokens
type JWT struct {
	keys              Keys
	blackListedTokens map[string]struct{}
}

// GenerateSecurityString generates a JWT token for the provided username
func (j *JWT) GenerateSecurityString(username string) (string, time.Time, error) {
	expirationTime := time.Now().Add(5 * time.Minute)
	claims := &Claim{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	key, err := j.keys.GetOne()
	if err != nil {
		return "", time.Time{}, err
	}
	tokenString, err := token.SignedString(key)
	return tokenString, expirationTime, err
}

// Validate validates whether the value of the receoved key is a valid token
func (j *JWT) Validate(tknStr string) (bool, string, error) {
	claim := &Claim{}
	isBlacklisted := j.isBlacklisted(tknStr)
	if isBlacklisted {
		return false, "", nil
	}
	tkn, err := jwt.ParseWithClaims(tknStr, claim, func(token *jwt.Token) (interface{}, error) {
		return j.keys.GetOne()
	})
	if err != nil {
		return false, "", err
	}
	return tkn.Valid, claim.Username, nil
}

// Invalidate blacklists a token so that it cannot be used anymore
func (j *JWT) Invalidate(token string) error {
	j.blackListedTokens[token] = struct{}{}
	return nil
}

// IsBlacklisted checks if a token is blacklisted
func (j *JWT) isBlacklisted(token string) bool {
	_, ok := j.blackListedTokens[token]
	return ok
}

// clearExpiredFromBlacklist clears any invalid token from the blacklist
func (j *JWT) clearExpiredFromBlacklist() {
	stillValid := map[string]struct{}{}
	for token := range j.blackListedTokens {
		valid, _, err := j.Validate(token)
		if valid && err == nil {
			stillValid[token] = struct{}{}
		}
	}
	j.blackListedTokens = stillValid
}
