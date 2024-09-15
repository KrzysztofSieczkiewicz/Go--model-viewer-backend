package middleware

import "net/http"

func Cors(next http.Handler) http.Handler {
	return http.HandlerFunc( func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Add("Access-Control-Allow-Origin", "http://localhost:3000")
		rw.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		rw.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight request
		if r.Method == http.MethodOptions {
            rw.WriteHeader(http.StatusOK)
            return
        }

		next.ServeHTTP(rw, r)
	})
}