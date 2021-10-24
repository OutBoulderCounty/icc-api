package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
)

func Authorize(ctx context.Context, token []byte, roles []string) (bool, error) {
	// strip the Bearer prefix
	if strings.HasPrefix(string(token), "Bearer ") {
		token = token[7:]
	}
	// get the public key
	jwksUrl := "https://icc.frontegg.com/.well-known/jwks.json"
	keyset, err := jwk.Fetch(ctx, jwksUrl)
	if err != nil {
		return false, err
	}
	// Parse the token
	parsedToken, err := jwt.Parse(token, jwt.WithKeySet(keyset))
	if err != nil {
		return false, err
	}
	// Get the claims
	claims := parsedToken.PrivateClaims()
	fmt.Println(claims)
	return true, nil
}
