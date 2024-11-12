package signedurl_test

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/KrzysztofSieczkiewicz/go--model-viewer-backend/FilesService/signedurl"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	secretKey     = "testsecret"
	baseUrl       = "https://example.com/resource"
	defaultExpire = 10 * time.Minute
	tempID        = "testID"
)

// Helper to initialize SignedUrl instance
func setupSignedUrl() *signedurl.SignedUrl {
	return signedurl.NewSignedUrl(secretKey, baseUrl, defaultExpire)
}

// Helper to parse and extract URL parameters
func parseSignedUrl(t *testing.T, generatedURL string) url.Values {
	parsedURL, err := url.Parse(generatedURL)
	require.NoError(t, err, "Failed to parse URL")
	return parsedURL.Query()
}

// Helper to validate presence of parameters in URL
func assertUrlParams(t *testing.T, queryParams url.Values, expectedID string) (string, string, string) {
	id := queryParams.Get("id")
	expires := queryParams.Get("expires")
	signature := queryParams.Get("signature")

	assert.Equal(t, expectedID, id, "Expected 'id' parameter to match tempID")
	assert.NotEmpty(t, expires, "Expected 'expires' parameter to be present")
	assert.NotEmpty(t, signature, "Expected 'signature' parameter to be present")

	return id, expires, signature
}

func TestNewSignedUrlInitialization(t *testing.T) {
	signedUrlGen := setupSignedUrl()

	// Generate and parse the signed URL
	generatedURL := signedUrlGen.GenerateSignedUrl(tempID)
	_, err := url.Parse(generatedURL)
	require.NoError(t, err, fmt.Sprintf("Failed to parse URL: %v", err))

	// Verify URL begins with base URL and contains expected parameters
	assert.True(t, strings.HasPrefix(generatedURL, baseUrl), fmt.Sprintf("Expected URL to start with base URL %s", baseUrl))
}

func TestGenerateSignedUrl(t *testing.T) {
	signedUrlGen := setupSignedUrl()
	generatedURL := signedUrlGen.GenerateSignedUrl(tempID)
	queryParams := parseSignedUrl(t, generatedURL)

	_, expires, signature := assertUrlParams(t, queryParams, tempID)

	assert.NotEmpty(t, expires, "Expected non-empty 'expires' value")
	assert.NotEmpty(t, signature, "Expected non-empty 'signature' value")
}

func TestGenerateSignedUrlCustom(t *testing.T) {
	signedUrlGen := setupSignedUrl()
	customExpire := 5 * time.Minute

	generatedURL := signedUrlGen.GenerateSignedUrlCustom(tempID, customExpire)
	queryParams := parseSignedUrl(t, generatedURL)

	_, expiresStr, _ := assertUrlParams(t, queryParams, tempID)

	// Verify expiration time is within expected range
	expirationTime, err := strconv.ParseInt(expiresStr, 10, 64)
	require.NoError(t, err, "Expiration time should be a valid integer")
	expectedExpiration := time.Now().Add(customExpire).Unix()
	assert.InDelta(t, expectedExpiration, expirationTime, 5, "Expiration time is not within expected range")
}

func TestGenerateAndValidateSignedUrlCustom(t *testing.T) {
	signedUrlGen := setupSignedUrl()
	customExpire := 5 * time.Minute

	generatedURL := signedUrlGen.GenerateSignedUrlCustom(tempID, customExpire)
	queryParams := parseSignedUrl(t, generatedURL)

	id, expires, signature := assertUrlParams(t, queryParams, tempID)

	// Validate signed URL
	err := signedUrlGen.ValidateSignedUrl(id, expires, signature)
	assert.NoError(t, err, "Expected URL to be valid")
}

func TestValidateSignedUrl(t *testing.T) {
	signedUrlGen := setupSignedUrl()
	customExpire := 5 * time.Minute

	t.Run("Valid URL", func(t *testing.T) {
		generatedURL := signedUrlGen.GenerateSignedUrlCustom(tempID, customExpire)
		queryParams := parseSignedUrl(t, generatedURL)

		id, expires, signature := assertUrlParams(t, queryParams, tempID)

		err := signedUrlGen.ValidateSignedUrl(id, expires, signature)
		assert.NoError(t, err, "Expected valid signed URL")
	})

	t.Run("Expired URL", func(t *testing.T) {
		expiredURL := signedUrlGen.GenerateSignedUrlCustom(tempID, -1*time.Minute)
		queryParams := parseSignedUrl(t, expiredURL)

		id, expires, signature := assertUrlParams(t, queryParams, tempID)

		err := signedUrlGen.ValidateSignedUrl(id, expires, signature)
		assert.Equal(t, signedurl.ErrUrlExpired, err, "Expected ErrUrlExpired for an expired signed URL")
	})

	t.Run("Invalid Signature", func(t *testing.T) {
		generatedURL := signedUrlGen.GenerateSignedUrlCustom(tempID, customExpire)
		queryParams := parseSignedUrl(t, generatedURL)

		id, expires := queryParams.Get("id"), queryParams.Get("expires")
		invalidSignature := "invalidsignature"

		err := signedUrlGen.ValidateSignedUrl(id, expires, invalidSignature)
		assert.Equal(t, signedurl.ErrInvalidSignature, err, "Expected ErrInvalidSignature for a URL with an invalid signature")
	})

	t.Run("Invalid Timestamp", func(t *testing.T) {
		generatedURL := signedUrlGen.GenerateSignedUrl(tempID)
		queryParams := parseSignedUrl(t, generatedURL)

		id := queryParams.Get("id")
		signature := queryParams.Get("signature")
		invalidTimestamp := "notanumber"

		err := signedUrlGen.ValidateSignedUrl(id, invalidTimestamp, signature)
		assert.ErrorContains(t, err, "invalid expiration timestamp", "Expected an error for invalid expiration timestamp")
	})
}
