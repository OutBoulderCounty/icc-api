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

func Authenticate(c *gin.Context, stytchClient *stytchapi.API) bool {
	var auth Auth
	if err := c.ShouldBindJSON(&auth); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return false
	}
	resp, err := stytchClient.MagicLinks.Authenticate(&stytch.MagicLinksAuthenticateParams{
		Token:                  auth.Token,
		SessionDurationMinutes: 10080,
	})
	if err != nil {
		fmt.Println("Failed to authenticate:", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return false
	}
	c.JSON(http.StatusOK, gin.H{"session_token": resp.SessionToken})

	return true
}
