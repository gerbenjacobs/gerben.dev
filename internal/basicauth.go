package internal

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
)

func BasicAuth(next http.HandlerFunc, expectedPasswordHash string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if expectedPasswordHash == "" {
			http.Error(w, "No secret token provided", http.StatusUnauthorized)
			return
		}

		_, password, ok := r.BasicAuth()
		if ok {
			passwordHash := sha256.Sum256([]byte(password))
			passwordHashHex := base64.URLEncoding.EncodeToString(passwordHash[:])
			passwordMatch := (subtle.ConstantTimeCompare([]byte(expectedPasswordHash), []byte(passwordHashHex)) == 1)

			if passwordMatch {
				next.ServeHTTP(w, r)
				return
			}
		}

		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
