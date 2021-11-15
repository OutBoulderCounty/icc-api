package main

import (
	"fmt"
	"log"
	"net/http"

	"api/database"
	"api/forms"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stytchauth/stytch-go/v3/stytch"
	"github.com/stytchauth/stytch-go/v3/stytch/stytchapi"
)

func main() {
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8000", "https://*.inclusivecareco.org, http://localhost:3000"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	db, err := database.Connect("dev")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// TODO: put secrets in env
	stytchClient := stytchapi.NewAPIClient(stytch.EnvTest, "project-test-c1af026b-43ef-4aee-830c-ace28e8822ac", "secret-test-ChVyWBGfThTd7Q59zqfuBuM7QuDGgdFakLg=")

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})

	r.POST("/login", func(c *gin.Context) {
		var json struct {
			Email string `json:"email"`
		}
		err := c.BindJSON(&json)
		if err != nil {
			fmt.Println("Failed to bind JSON: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		body := stytch.MagicLinksEmailLoginOrCreateParams{
			Email:              json.Email,
			LoginMagicLinkURL:  "http://localhost:8080/authenticate",
			SignupMagicLinkURL: "http://localhost:8080/authenticate",
		}
		resp, err := stytchClient.MagicLinks.Email.LoginOrCreate(&body)
		if err != nil {
			fmt.Println("Failed to create magic link: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
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
	})

	r.GET("/authenticate", func(c *gin.Context) {
		var user struct {
			Token string `form:"token"`
		}
		err := c.ShouldBindQuery(&user)
		if err != nil {
			fmt.Println("Failed to bind query: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		body := stytch.MagicLinksAuthenticateParams{
			Token:                  user.Token,
			SessionDurationMinutes: 10080,
		}
		resp, err := stytchClient.MagicLinks.Authenticate(&body)
		if err != nil {
			fmt.Println("Failed to authenticate: " + err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}
		fmt.Println("Authenticated!", resp.UserID, resp.SessionToken)
		c.JSON(http.StatusOK, gin.H{
			"user_id":       resp.UserID,
			"session_token": resp.SessionToken,
		})
	})

	authorized := r.Group("/forms", authRequired(stytchClient))

	authorized.GET("", func(c *gin.Context) {
		foundForms, err := forms.GetForms(db.DB)
		if err != nil {
			panic(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"forms": foundForms,
		})
	})

	r.Run()
}

func authRequired(stytchClient *stytchapi.API) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.Request.Header.Get("Authorization")
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Unauthorized",
			})
			c.Abort()
			return
		}
		body := stytch.SessionsAuthenticateParams{
			SessionToken:           token,
			SessionDurationMinutes: 10080,
		}
		resp, err := stytchClient.Sessions.Authenticate(&body)
		if err != nil {
			fmt.Println("Failed to authorize: " + err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}
		fmt.Println("Authorized!", resp.Session.UserID)
		c.Next()
	}
}
