package users

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stytchauth/stytch-go/v3/stytch"
	"github.com/stytchauth/stytch-go/v3/stytch/stytchapi"
)

type Auth struct {
	Token string `json:"token"`
}

type SessionAuth struct {
	SessionToken string `json:"session_token"`
}

// AuthenticateUser is a gin handler function that authenticates a user
func AuthenticateUser(c *gin.Context, stytchClient *stytchapi.API) bool {
	var auth Auth
	if err := c.ShouldBindJSON(&auth); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return false
	}
	sessionToken, err := Authenticate(auth.Token, stytchClient)
	if err != nil {
		fmt.Println("Failed to authenticate:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return false
	}
	c.JSON(http.StatusOK, gin.H{"session_token": sessionToken})

	return true
}

// Authenticates a token
func Authenticate(token string, stytchClient *stytchapi.API) (sessionToken string, err error) {
	resp, err := stytchClient.MagicLinks.Authenticate(&stytch.MagicLinksAuthenticateParams{
		Token:                  token,
		SessionDurationMinutes: 10080,
	})
	if err != nil {
		return "", err
	}
	return resp.SessionToken, nil
}
