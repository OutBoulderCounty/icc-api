package auth

import (
	"context"
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
	// Get the roles
	var roleInterfaces []interface{}
	for k, v := range claims {
		if k == "roles" {
			roleInterfaces = v.([]interface{})
		}
	}
	var userRoles []string
	for i := 0; i < len(roleInterfaces); i++ {
		userRoles = append(userRoles, roleInterfaces[i].(string))
	}

	// check if the user has the right roles
	for i := 0; i < len(roles); i++ {
		for j := 0; j < len(userRoles); j++ {
			if roles[i] == userRoles[j] {
				return true, nil
			}
		}
	}
	return false, nil
}
