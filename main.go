package main

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

type investigator struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Class string `json:"class"`
}

type investigatorInput struct {
	Name  string `json:"name" binding:"required"`
	Class string `json:"class" binding:"required"`
}

/*
Test: This should successfully add roland to the data
curl http://localhost:8080/investigators/ \
    --include \
    --header "Content-Type: application/json" \
    --request "POST" \
    --data '{"id": "4","Name": "Roland Banks","Class": "Guardian"}'
*/

func getAllInvestigators(db *sql.DB) ([]investigator, error) {

	fullList := make([]investigator, 0)

	rows, err := db.Query("SELECT * FROM investigators")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	if fullList, err = parseRowsToGatorList(rows); err != nil {
		log.Println(err)
		return fullList, err
	}
	return fullList, nil
}

func getInvestigatorByID(db *sql.DB, id string) ([]investigator, error) {

	fullList := make([]investigator, 0)

	rows, err := db.Query("SELECT * FROM investigators WHERE id = $1", id)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	if fullList, err = parseRowsToGatorList(rows); err != nil {
		log.Println(err)
		return fullList, err
	}
	return fullList, nil
}

func parseRowsToGatorList(rows *sql.Rows) ([]investigator, error) {

	fullList := make([]investigator, 0)

	for rows.Next() {
		var id string
		var name string
		var class string
		if err := rows.Scan(&id, &name, &class); err != nil {
			log.Println(err)
			return fullList, err
		}
		newInvestigator := investigator{
			ID:    id,
			Name:  name,
			Class: class,
		}

		fullList = append(fullList, newInvestigator)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return fullList, nil
}

func addNewInvestigator(db *sql.DB, gator investigatorInput) (int, error) {
	id := 0
	err := db.QueryRow("INSERT INTO investigators (name,class) VALUES ($1,$2) RETURNING id",
		gator.Name, gator.Class,
	).Scan(&id)

	if err != nil {
		return id, err
	}
	return id, nil
}

func main() {

	dbConnStr := "user=gouser dbname=aeons password=gopassword sslmode=disable"
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	router := gin.Default()

	router.GET("/investigators", func(c *gin.Context) {

		allGators, err := getAllInvestigators(db)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusOK, allGators)
			return
		}
		c.JSON(http.StatusOK, allGators)
	})

	router.GET("/investigators/:id", func(c *gin.Context) {
		id := c.Param("id")
		foundGators, err := getInvestigatorByID(db, id)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, foundGators)
	})

	router.POST("/investigators", func(c *gin.Context) {

		var newGatorDTO investigatorInput

		if err := c.BindJSON(&newGatorDTO); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := addNewInvestigator(db, newGatorDTO)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, result)
	})

	router.Run("localhost:8080")
}
