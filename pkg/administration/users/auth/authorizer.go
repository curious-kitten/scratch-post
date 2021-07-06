package auth

import (
	"net/http"

	jwt "github.com/dgrijalva/jwt-go"
)

// Authorization verifies a request has a valid token associated
func Authorization(authorizer Authorizer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			token, err := GetToken(r)
			if err != nil {
				if err == http.ErrNoCookie {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			valid, username, err := authorizer.Validate(token)
			if err != nil {
				if err == jwt.ErrSignatureInvalid {
					w.WriteHeader(http.StatusUnauthorized)
					return
				}
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			if !valid {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			AddUserIDHeader(r, username)
			defer RemoveUserIDHeader(r)
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
