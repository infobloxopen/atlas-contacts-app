package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc/metadata"
)

const (
	// TestSecret dummy secret used for signing test JWTs
	TestSecret = "some-secret-123"
)

var (
	// DefaultClaims is the standard payload inside a test JWT
	DefaultClaims = jwt.MapClaims{"AccountID": "TestAccount"}
)

// DefaultContext returns a context that has a jwt for basic testing purposes
func DefaultContext(t *testing.T) context.Context {
	token, err := MakeToken(DefaultClaims)
	if err != nil {
		t.Fatalf("unable to create default context: %v", err)
	}
	ctx := context.Background()
	return ContextWithToken(ctx, token)
}

// ContextWithToken creates a context with a jwt
func ContextWithToken(ctx context.Context, token string) context.Context {
	c := metadata.AppendToOutgoingContext(
		ctx,
		"Authorization", fmt.Sprintf("token %s", token),
	)
	return c
}

// MakeToken generates a token string based on the given jwt claims
func MakeToken(claims jwt.Claims) (string, error) {
	method := jwt.SigningMethodHS256
	token := jwt.NewWithClaims(method, claims)
	signingString, err := token.SigningString()
	if err != nil {
		return "", err
	}
	signature, err := method.Sign(signingString, []byte(TestSecret))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s.%s", signingString, signature), nil
}

// AddDefaultTokenToRequest adds the default authorization token to the http
// header of a given http request
func AddDefaultTokenToRequest(req *http.Request) {
	token, err := MakeToken(DefaultClaims)
	if err != nil {
		log.Fatalf("unable to create token")
	}
	AddTokenToRequest("token", token, req)
}

// AddTokenToRequest adds an authorization token to the http request header
func AddTokenToRequest(key, token string, req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("%s %s", key, token))
}

// MakeRequestWithDefaults issues request that contains the necessary parameters
// request to reach the contacts application (e.g. an authorization token)
func MakeRequestWithDefaults(method, url string, payload interface{}) (*http.Response, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	AddDefaultTokenToRequest(req)
	if err != nil {
		return nil, err
	}
	client := http.Client{}
	return client.Do(req)
}
