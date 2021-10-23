package main

import (
	"log"
	"net/http"

	"api/database"
	"api/forms"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	db, err := database.Connect("dev")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r.GET("/forms", func(c *gin.Context) {
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
