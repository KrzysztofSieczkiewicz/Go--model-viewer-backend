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

type SignedUrlConfig struct {
	
}

type SignedSlug struct {
	secretKey string
	expiresIn  time.Duration
}

func NewSignedSlug(secretKey string, expireIn time.Duration) *SignedSlug {
	return &SignedSlug{
		secretKey: secretKey,
		expiresIn: expireIn,
	}
}

func (s *SignedSlug) GenerateSignedSlug(tempID string) string {
	expirationTime := time.Now().Add(s.expiresIn).Unix()
	signature := s.createHMACSignature(tempID, expirationTime)

	signedURL := fmt.Sprintf(
		"?id=%s&expires%d&signature%s",
		url.QueryEscape(tempID),
		expirationTime,
		signature,
	)

	return signedURL
}

func (s *SignedSlug) GenerateSignedSlugCustom(tempID string, expiresIn time.Duration) string {
	expirationTime := time.Now().Add(expiresIn).Unix()
	signature := s.createHMACSignature(tempID, expirationTime)

	signedUrl := fmt.Sprintf(
		"?id=%s&expires%d&signature%s",
		url.QueryEscape(tempID),
		expirationTime,
		signature,
	)

	return signedUrl
}

func (s *SignedSlug) ValidateSignedSlug(id string, expires string, signature string) error {
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

func (s *SignedSlug) createHMACSignature(tempID string, expirationTime int64) string {
	mac := hmac.New(sha256.New, []byte(s.secretKey))
	data := fmt.Sprintf("%s:%d", tempID, expirationTime)
	mac.Write([]byte(data))

	return hex.EncodeToString(mac.Sum(nil))
}