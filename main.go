package main

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

type investigatorStore struct {
	mutex         sync.RWMutex
	investigators []investigator
}

func (s *investigatorStore) add(gator investigator) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.investigators = append(s.investigators, gator)
}

func (s *investigatorStore) getAll() []investigator {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.investigators
}

var gatorStore = investigatorStore{
	investigators: investigators,
}

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
curl http://localhost:8080/investigators/ \
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

	gatorStore.add(newInvestigator)
	c.IndentedJSON(http.StatusCreated, newInvestigator)
}

func getInvestigators(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gatorStore.getAll())
}

func getInvestigatorByID(c *gin.Context) {
	id := c.Param("id")

	for _, gator := range gatorStore.getAll() {
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
	router.GET("/investigators", getInvestigators)
	router.GET("/investigators/:id", getInvestigatorByID)
	router.POST("/investigators", addInvestigator)

	router.GET("/ping", ping)

	router.Run("localhost:8080")
}
