package signedurl

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

type SignedUrl struct {
	secretKey	string
	baseUrl		string
	expiresIn	time.Duration
}

func NewSignedUrl(secretKey string, baseUrl string, defaultExpire time.Duration) *SignedUrl {
	return &SignedUrl{
		secretKey:	secretKey,
		baseUrl:	baseUrl,
		expiresIn:	defaultExpire,
	}
}

// Returns a new signed url with default expiration time
func (s *SignedUrl) GenerateSignedUrl(tempID string) string {
	signedURL := s.GenerateSignedUrlCustom(
		tempID,
		s.expiresIn,
	)

	return signedURL
}

// Returns a new signed url with custom expiration time
func (s *SignedUrl) GenerateSignedUrlCustom(tempID string, expiresIn time.Duration) string {
	expirationTime := time.Now().Add(expiresIn).Unix()
	signature := s.createHMACSignature(tempID, expirationTime)

	signedUrl := fmt.Sprintf(
		"%s/?id=%s&expires=%d&signature=%s",
		s.baseUrl,
		url.QueryEscape(tempID),
		expirationTime,
		signature,
	)

	return signedUrl
}

// Checks if signed url is valid. Verifies id, signature and checks expiration time. Returns an error on any discrepancy
func (s *SignedUrl) ValidateSignedUrl(id string, expires string, signature string) error {
	expirationTime, err := strconv.ParseInt(expires, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid expiration timestamp")
	}
	
	if time.Now().Unix() > expirationTime {
		return fmt.Errorf("URL has expired")
	}

	expectedSignature := s.createHMACSignature(id, expirationTime)
	if !hmac.Equal([]byte(expectedSignature), []byte(signature)) {
		return fmt.Errorf("invalid signature")
	}

	return nil
}

// Generates HMAC signature based on secret key
func (s *SignedUrl) createHMACSignature(tempID string, expirationTime int64) string {
	mac := hmac.New(sha256.New, []byte(s.secretKey))
	data := fmt.Sprintf("%s:%d", tempID, expirationTime)
	mac.Write([]byte(data))

	return hex.EncodeToString(mac.Sum(nil))
}