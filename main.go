package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"api/env"
	"api/forms"
	"api/users"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stytchauth/stytch-go/v3/stytch"
)

func setup() *env.Env {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	appEnv := os.Getenv("APP_ENV")
	var environment *env.Env
	switch appEnv {
	case "prod":
		environment, err = env.Connect(env.EnvProd)
	case "test":
		environment, err = env.Connect(env.EnvTest)
	case "dev":
		environment, err = env.Connect(env.EnvDev)
	default:
		log.Fatal("Invalid APP_ENV")
	}
	if err != nil {
		log.Fatal("Failed to connect services: " + err.Error())
	}

	config := cors.DefaultConfig()
	config.AllowWildcard = true
	config.AllowOrigins = []string{"http://localhost:8000", "https://*.inclusivecareco.org", "http://localhost:3000", "http://localhost:3002"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	environment.Router.Use(cors.New(config))

	environment.Router.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})

	environment.Router.POST("/login", func(c *gin.Context) {
		users.LoginHandler(c, environment)
	})

	environment.Router.POST("/authenticate", func(c *gin.Context) {
		users.AuthenticateUser(c, environment)
	})

	// for testing locally without a UI
	environment.Router.GET("/localauth", func(c *gin.Context) {
		var login struct {
			Token string `form:"token"`
		}
		err := c.ShouldBindQuery(&login)
		if err != nil {
			fmt.Println("Failed to bind query: " + err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		sessionToken, err := users.Authenticate(login.Token, environment)
		if err != nil {
			fmt.Println("Failed to authenticate: " + err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}
		fmt.Println("Authenticated!")
		c.JSON(http.StatusOK, gin.H{
			"session_token": sessionToken,
		})
	})

	authorizedUser := environment.Router.Group("/user", authRequired(environment))

	authorizedUser.PUT("", func(c *gin.Context) {
		users.UpdateUserHandler(c, environment)
	})

	authorizedForms := environment.Router.Group("/forms", authRequired(environment))

	authorizedForms.GET("", func(c *gin.Context) {
		foundForms, err := forms.GetForms(environment.DB)
		if err != nil {
			panic(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"forms": foundForms,
		})
	})

	return environment
}

func main() {
	env := setup()
	defer env.DB.Close()
	env.Router.Run()
}

func authRequired(environment *env.Env) gin.HandlerFunc {
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
		resp, err := environment.Stytch.Sessions.Authenticate(&body)
		if err != nil {
			fmt.Println("Failed to authorize: " + err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}
		fmt.Println("Authorized!", resp.Session.UserID)
		c.Set("stytch_user_id", resp.Session.UserID)
		c.Next()
	}
}
