package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"time"

	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

var populateDB = flag.String("populateDB", "", "set flag all if you have an empty db, set flag gator if you only want gators, set flag classes if you only want classes")

type class struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type campaign struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type scenario struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Position int      `json:"campaign_position"`
	Campaign campaign `json:"campaign"`
}

type investigator struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Class class  `json:"class"`
}

type investigatorInput struct {
	Name  string `json:"name" binding:"required"`
	Class string `json:"class" binding:"required"`
}

type session struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Scenario scenario `json:"scenario"`
}

type play struct {
	ID                 int          `json:"id"`
	Name               string       `json:"name"`
	Session            session      `json:"session"`
	PlayedInvestigator investigator `json:"investigator"`
	SelfPlayed         bool         `json:"self_played"`
}

var seedFuncs = map[string]func(context.Context, *sql.DB) error{
	"all":     populateDatabase,
	"gators":  populateInvestigators,
	"classes": populateClasses,
}

func getAllInvestigators(ctx context.Context, db *sql.DB) ([]investigator, error) {

	fullList := make([]investigator, 0)

	rows, err := db.QueryContext(
		ctx, `
		SELECT i.investigator_id, i.investigator_name, c.class_id, c.class_name
		FROM investigators i
		JOIN classes c
		ON i.investigator_class = c.class_id
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if fullList, err = parseRowsToGatorList(rows); err != nil {
		return fullList, err
	}
	return fullList, nil
}

func parseRowsToGatorList(rows *sql.Rows) ([]investigator, error) {

	fullList := make([]investigator, 0)

	for rows.Next() {

		gator, err := parseRowToGator(rows)
		if err != nil {
			return nil, err
		}
		fullList = append(fullList, gator)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return fullList, nil
}

func parseRowToGator(rows *sql.Rows) (investigator, error) {

	var gatorID int
	var gatorName string
	var classID int
	var className string

	if err := rows.Scan(&gatorID, &gatorName, &classID, &className); err != nil {
		return investigator{}, err
	}
	newInvestigator := investigator{
		ID:   gatorID,
		Name: gatorName,
		Class: class{
			ID:   classID,
			Name: className,
		},
	}

	return newInvestigator, nil
}

func getInvestigatorByID(ctx context.Context, db *sql.DB, id string) (investigator, error) {

	var investigatorID int
	var investigatorName string
	var classID int
	var className string

	err := db.QueryRowContext(
		ctx, `
		SELECT i.investigator_id, i.investigator_name, c.class_id, c.class_name 
		FROM investigators i
		JOIN classes c
		ON i.investigator_class = c.class_id
		WHERE i.investigator_id = $1`,
		id).Scan(&investigatorID, &investigatorName, &classID, &className)

	if err != nil {
		return investigator{}, err
	}

	return investigator{
			ID:   investigatorID,
			Name: investigatorName,
			Class: class{
				ID:   classID,
				Name: className,
			},
		},
		nil
}

/*
Test Curl:
curl -X POST localhost:8080/investigators \
  -H "Content-Type: application/json" \
  -d '{"name": "Michael McGlenn", "class": "Rogue"}'

*/

func addNewInvestigator(ctx context.Context, db *sql.DB, gatorDTO investigatorInput) (int, error) {
	id := 0

	class_id, err := getIDForClassName(ctx, db, gatorDTO.Class)
	if err != nil {
		return 0, err
	}

	log.Printf("Inserting investigator %s \n", gatorDTO.Name)
	err = db.QueryRowContext(ctx,
		"INSERT INTO investigators (investigator_name,investigator_class) VALUES ($1,$2) RETURNING investigator_id",
		gatorDTO.Name, class_id,
	).Scan(&id)

	if err != nil {
		return id, err
	}
	log.Printf("Successfully inserted investigator %s \n", gatorDTO.Name)
	return id, nil
}

func addClass(ctx context.Context, db *sql.DB, newClassName string) (int, error) {
	var id int
	log.Printf("Inserting class %s \n", newClassName)
	err := db.QueryRowContext(ctx,
		"INSERT INTO classes (class_name) VALUES ($1) RETURNING class_id",
		newClassName,
	).Scan(&id)

	if err != nil {
		return 0, err
	}
	log.Printf("Inserted class %s \n", newClassName)
	return id, nil
}

func getIDForClassName(ctx context.Context, db *sql.DB, className string) (int, error) {
	id := 0
	log.Printf("Fetching ID for class name %s \n", className)
	err := db.QueryRowContext(ctx, "SELECT class_id FROM classes WHERE class_name=$1", className).Scan(&id)
	if err != nil {
		return 0, err
	}
	log.Printf("Got ID %d for class name %s \n", id, className)
	return id, nil

}

func populateDatabase(ctx context.Context, db *sql.DB) error {

	err := populateClasses(ctx, db)
	if err != nil {
		return err
	}

	err = populateInvestigators(ctx, db)
	if err != nil {
		return err
	}

	return nil
}

func populateInvestigators(ctx context.Context, db *sql.DB) error {

	seedInvestigators := []investigatorInput{
		{Name: "Roland Banks", Class: "Guardian"},
		{Name: "Daisy Walkers", Class: "Seeker"},
		{Name: "Skids O'Toole", Class: "Rogue"},
		{Name: "Agnes Baker", Class: "Mystic"},
		{Name: "Wendy Adams", Class: "Survivor"},
		{Name: "Zoey Samras", Class: "Guardian"},
		{Name: "Rex Murphy", Class: "Seeker"},
		{Name: "Jenny Barnes", Class: "Rogue"},
		{Name: "Jim Culver", Class: "Mystic"},
		{Name: `"Ashcan" Pete`, Class: "Survivor"},
		{Name: "Mark Harrigan", Class: "Guardian"},
		{Name: "Minh Thi Phan ", Class: "Seeker"},
		{Name: "Sefina Rousseau", Class: "Rogue"},
		{Name: "Akachi Onyele", Class: "Mystic"},
		{Name: "William Yorick", Class: "Survivor"},
		{Name: "Lola Hayes", Class: "Neutral"},
	}

	for _, v := range seedInvestigators {
		_, err := addNewInvestigator(ctx, db, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func populateClasses(ctx context.Context, db *sql.DB) error {

	classes := [6]string{"Guardian", "Seeker", "Rogue", "Mystic", "Survivor", "Neutral"}

	for _, v := range classes {
		_, err := addClass(ctx, db, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	flag.Parse()

	dbConnStr := "user=gouser dbname=aeons password=gopassword sslmode=disable"
	db, err := sql.Open("postgres", dbConnStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		log.Fatal(err)
	}

	seedFunc, ok := seedFuncs[*populateDB]
	if !ok {
		log.Println("Empty or unknown seeding. Continuing without seeding.")
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := seedFunc(ctx, db); err != nil {
			log.Print(err)
			return
		}
	}

	router := gin.Default()

	router.GET("/investigators", func(c *gin.Context) {

		allGators, err := getAllInvestigators(c, db)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, allGators)
	})

	router.GET("/investigators/:id", func(c *gin.Context) {
		id := c.Param("id")
		foundGators, err := getInvestigatorByID(c, db, id)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, foundGators)
	})

	router.POST("/investigators", func(c *gin.Context) {

		var newGatorDTO investigatorInput

		if err := c.BindJSON(&newGatorDTO); err != nil {
			log.Println(err)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		result, err := addNewInvestigator(c, db, newGatorDTO)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, result)
	})

	router.Run("localhost:8080")
}
