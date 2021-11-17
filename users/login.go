package users

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/stytchauth/stytch-go/v3/stytch"
	"github.com/stytchauth/stytch-go/v3/stytch/stytchapi"
)

type User struct {
	Email       string `json:"email"`
	RedirectURL string `json:"redirect_url"` // must be defined in Stytch as a redirect URL
}

func Login(c *gin.Context, stytchClient *stytchapi.API) (*User, error) {
	var user User
	err := c.BindJSON(&user)
	if err != nil {
		fmt.Println("Failed to bind JSON: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return nil, err
	}
	body := stytch.MagicLinksEmailLoginOrCreateParams{
		Email:              user.Email,
		LoginMagicLinkURL:  user.RedirectURL,
		SignupMagicLinkURL: user.RedirectURL,
	}
	resp, err := stytchClient.MagicLinks.Email.LoginOrCreate(&body)
	if err != nil {
		fmt.Println("Failed to create magic link: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return nil, err
	}
	var status int
	if resp.UserCreated {
		// TODO: create user in DB
		fmt.Println("User created")
		status = http.StatusCreated
	} else {
		status = http.StatusOK
	}
	c.JSON(status, gin.H{
		"user_id": resp.UserID,
	})
	return &user, nil
}
