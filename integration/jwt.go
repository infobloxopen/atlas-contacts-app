package integration

import (
	"context"
	"fmt"
	"testing"

	jwt "github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc/metadata"
)

const (
	// TestSecret dummy secret used for signing test JWTs
	TestSecret = "some-secret-123"
)

// DefaultContext returns a context that has a jwt for basic testing purposes
func DefaultContext(t *testing.T) context.Context {
	token, err := MakeToken(jwt.MapClaims{
		"AccountID": "TestAccount",
	})
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
