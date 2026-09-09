package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type investigator struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Class string `json:"class"`
}

var investigators = []investigator{
	{ID: "1", Name: "Luke Robinson", Class: "Mystic"},
	{ID: "2", Name: "Daisy Walker", Class: "Seeker"},
	{ID: "3", Name: "Norman Withers", Class: "Seeker"},
}

/*
Test: This should successfully add roland to the data
curl http://localhost:8080/investigators/add \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '{"id": "4","Name": "Roland Banks","Class": "Guardian"}'
*/

func addInvestigator(c *gin.Context) {

	var newInvestigator investigator

	if err := c.BindJSON(&newInvestigator); err != nil {
		return
	}

	investigators = append(investigators, newInvestigator)
	c.IndentedJSON(http.StatusCreated, newInvestigator)
}

func getInvestigators(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, investigators)
}

func getInvestigatorById(c *gin.Context) {
	id := c.Param("id")

	for _, gator := range investigators {
		if gator.ID == id {
			c.IndentedJSON(http.StatusOK, gator)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "investigator not found"})
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func main() {
	router := gin.Default()
	router.GET("/investigators/get", getInvestigators)
	router.GET("/investigators/get/:id", getInvestigatorById)
	router.POST("/investigators/add", addInvestigator)

	router.GET("ping", ping)

	router.Run("localhost:8080")
}
