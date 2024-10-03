package utils

import (
	"fmt"
	"net/http"
)

func RespondWithMessage(rw http.ResponseWriter, statusCode int, message string) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)

	fmt.Fprintf(rw, `{"message": "%s"}`, message)
}

func RespondWithJSON(rw http.ResponseWriter, statusCode int, data interface{}) {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(statusCode)

	ToJSON(data, rw)
}