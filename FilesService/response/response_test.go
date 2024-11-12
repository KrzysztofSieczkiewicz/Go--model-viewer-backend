package response_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/response"
	"github.com/stretchr/testify/assert"
)

func TestRespondWithNoContent(t *testing.T) {
	rw := httptest.NewRecorder()
	response.RespondWithNoContent(rw)

	assert.Equal(t, http.StatusNoContent, rw.Code)
	assert.Equal(t, "application/json", rw.Header().Get("Content-Type"))
}

// TestRespondWithMessage tests the RespondWithMessage function
func TestRespondWithMessage(t *testing.T) {
	message := "Success"
	statusCode := http.StatusOK

	// Call the function we are testing
	rw := httptest.NewRecorder()
	response.RespondWithMessage(rw, statusCode, message)

	assert.Equal(t, statusCode, rw.Code)
	assert.Equal(t, "application/json", rw.Header().Get("Content-Type"))

	// Check that the response body contains the correct message
	expectedBody := fmt.Sprintf(`{"message": "%s"}`, message)
	body := rw.Body.String()
	assert.Equal(t, expectedBody, strings.TrimSpace(body) )
}

// TestRespondWithJSON tests the RespondWithJSON function using actual JSON serialization
func TestRespondWithJSON(t *testing.T) {
	statusCode := http.StatusOK
	data := map[string]string{"key": "value"}

	rw := httptest.NewRecorder()
	response.RespondWithJSON(rw, statusCode, data)

	assert.Equal(t, statusCode, rw.Code)
	assert.Equal(t, "application/json", rw.Header().Get("Content-Type"))
	
	// Check that the response body contains the correct JSON output
	expectedBody, err := json.Marshal(data)
	assert.NoError(t, err)

	// Compare the expected body with the actual response body (ignoring extra spaces)
	actualBody := rw.Body.String()
	assert.Equal(t, string(expectedBody), strings.TrimSpace(actualBody) )
}

