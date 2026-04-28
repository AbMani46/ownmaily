package mailer

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"
)

const sesService = "ses"

// signRequest adds AWS Signature Version 4 auth to req.
// Signs content-type, host, and x-amz-date headers.
// body must be the raw request body bytes.
func signRequest(req *http.Request, accessKeyID, secretAccessKey, region string, body []byte) error {
	now := time.Now().UTC()
	timestamp := now.Format("20060102T150405Z")
	date := now.Format("20060102")

	host := req.URL.Host
	contentType := req.Header.Get("Content-Type")

	req.Header.Set("x-amz-date", timestamp)

	bodyHash := hexSHA256(body)

	canonicalHeaders := fmt.Sprintf("content-type:%s\nhost:%s\nx-amz-date:%s\n",
		contentType, host, timestamp)
	signedHeaders := "content-type;host;x-amz-date"

	canonicalRequest := fmt.Sprintf("%s\n%s\n%s\n%s\n%s\n%s",
		req.Method,
		req.URL.Path,
		req.URL.RawQuery,
		canonicalHeaders,
		signedHeaders,
		bodyHash,
	)

	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", date, region, sesService)

	stringToSign := fmt.Sprintf("AWS4-HMAC-SHA256\n%s\n%s\n%s",
		timestamp,
		credentialScope,
		hexSHA256([]byte(canonicalRequest)),
	)

	signingKey := hmacSHA256(
		hmacSHA256(
			hmacSHA256(
				hmacSHA256([]byte("AWS4"+secretAccessKey), date),
				region,
			),
			sesService,
		),
		"aws4_request",
	)

	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))

	req.Header.Set("Authorization", fmt.Sprintf(
		"AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKeyID, credentialScope, signedHeaders, signature,
	))

	return nil
}

func hexSHA256(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

func hmacSHA256(key []byte, data string) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write([]byte(data))
	return mac.Sum(nil)
}
