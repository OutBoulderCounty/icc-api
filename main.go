package main

import (
	"context"
	"log"
	"net/http"

	"api/auth"
	"api/database"
	"api/forms"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"http://localhost:8000", "https://*.inclusivecareco.org"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.Use(cors.New(config))

	db, err := database.Connect("dev") // TODO: use a variable
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r.GET("/forms", func(c *gin.Context) {
		// authorize
		token := c.Request.Header.Get("Authorization")
		ok, err := auth.Authorize(context.TODO(), []byte(token), []string{"Admin"})
		if err != nil {
			log.Println("Authorization failure: " + err.Error())
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if !ok {
			log.Println("User not authorized")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		foundForms, err := forms.GetForms(db.DB)
		if err != nil {
			panic(err)
		}
		c.JSON(http.StatusOK, gin.H{
			"forms": foundForms,
		})
	})

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hello World!")
	})
	r.Run()
}
