package users

import (
	"api/env"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stytchauth/stytch-go/v3/stytch"
)

type Auth struct {
	Token string `json:"token"`
}

type SessionAuth struct {
	SessionToken string `json:"session_token"`
}

// AuthenticateUser is a gin handler function that authenticates a user
func AuthenticateUser(c *gin.Context, e *env.Env) bool {
	var auth Auth
	if err := c.ShouldBindJSON(&auth); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return false
	}
	sessionToken, err := Authenticate(auth.Token, e)
	if err != nil {
		fmt.Println("Failed to authenticate:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return false
	}
	c.JSON(http.StatusOK, gin.H{"session_token": sessionToken})

	return true
}

// Authenticates a token
func Authenticate(token string, e *env.Env) (sessionToken string, err error) {
	resp, err := e.Stytch.MagicLinks.Authenticate(&stytch.MagicLinksAuthenticateParams{
		Token:                  token,
		SessionDurationMinutes: 10080,
	})
	if err != nil {
		return "", err
	}
	user, err := GetUserByStytchID(&resp.UserID, e)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", errors.New("User not found. Stytch user ID " + resp.UserID)
	}
	return resp.SessionToken, nil
}
