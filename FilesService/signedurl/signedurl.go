package signedurl

import "time"

type SignedUrlConfig struct {
	SecretKey string
	ExpireIn  time.Duration
}

// TODO: CONSIDER DISSOLVING CONFIG
type SignedURL struct {
	config SignedUrlConfig
}

func NewSignedURL(secretKey string, expireIn time.Duration) *SignedURL {
	return &SignedURL{
		config: SignedUrlConfig{
			SecretKey: secretKey,
			ExpireIn: expireIn,
		},
	}
}

func (s *SignedURL) GenerateSignedURL(filepath string, expires time.Duration) {
	
}