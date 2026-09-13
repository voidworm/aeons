package main

import (
	"context"
	"database/sql"
	"flag"
	"log"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"

	"github.com/gin-gonic/gin"
)

var populateDB = flag.Bool("populateDB", false, "set flag if you want to populate the db on a first run after db reset")

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

		gator, err := parseRowToGator(rows)
		if err != nil {
			return nil, err
		}
		fullList = append(fullList, gator)
	}
	if err := rows.Err(); err != nil {
		log.Println(err)
		return nil, err
	}

	return fullList, nil
}

func parseRowToGator(rows *sql.Rows) (investigator, error) {

	var gator_id int
	var gator_name string
	var class_id int
	var class_name string

	if err := rows.Scan(&gator_id, &gator_name, &class_id, &class_name); err != nil {
		log.Println(err)
		return investigator{}, err
	}
	newInvestigator := investigator{
		ID:   gator_id,
		Name: gator_name,
		Class: class{
			ID:   class_id,
			Name: class_name,
		},
	}

	return newInvestigator, nil
}

func getInvestigatorByID(ctx context.Context, db *sql.DB, id string) (investigator, error) {

	var investigator_id int
	var investigator_name string
	var class_id int
	var class_name string

	err := db.QueryRowContext(
		ctx, `
		SELECT i.investigator_id, i.investigator_name, c.class_id, c.class_name 
		FROM investigators i
		JOIN classes c
		ON i.investigator_class = c.class_id
		WHERE i.investigator_id = $1`,
		id).Scan(&investigator_id, &investigator_name, &class_id, &class_name)

	if err != nil {
		log.Println(err)
		return investigator{}, err
	}

	return investigator{
			ID:   investigator_id,
			Name: investigator_name,
			Class: class{
				ID:   class_id,
				Name: class_name,
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
func populateDatabase(db *sql.DB) error {

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err := addClass(ctx, db, "Guardian")

	if err != nil {
		return err
	}
	addClass(ctx, db, "Seeker")
	addClass(ctx, db, "Rogue")
	addClass(ctx, db, "Mystic")
	addClass(ctx, db, "Survivor")
	addClass(ctx, db, "Neutral")

	_, err = addNewInvestigator(ctx, db, investigatorInput{Name: "Roland Banks", Class: "Guardian"})
	if err != nil {
		return err
	}

	addNewInvestigator(ctx, db, investigatorInput{Name: "Daisy Walkers", Class: "Seeker"})
	addNewInvestigator(ctx, db, investigatorInput{Name: "Skids O'Toole", Class: "Rogue"})
	addNewInvestigator(ctx, db, investigatorInput{Name: "Agnes Baker", Class: "Mystic"})
	addNewInvestigator(ctx, db, investigatorInput{Name: "Wendy Adams", Class: "Survivor"})
	addNewInvestigator(ctx, db, investigatorInput{Name: "Zoey Samras", Class: "Guardian"})
	addNewInvestigator(ctx, db, investigatorInput{Name: "Rex Murphy", Class: "Seeker"})
	addNewInvestigator(ctx, db, investigatorInput{Name: "Jenny Barnes", Class: "Rogue"})
	addNewInvestigator(ctx, db, investigatorInput{Name: "Jim Culver", Class: "Mystic"})
	addNewInvestigator(ctx, db, investigatorInput{Name: `"Ashcan" Pete`, Class: "Survivor"})
	addNewInvestigator(ctx, db, investigatorInput{Name: "Mark Harrigan", Class: "Guardian"})
	addNewInvestigator(ctx, db, investigatorInput{Name: "Minh Thi Phan ", Class: "Seeker"})
	addNewInvestigator(ctx, db, investigatorInput{Name: "Sefina Rousseau", Class: "Rogue"})
	addNewInvestigator(ctx, db, investigatorInput{Name: "Akachi Onyele", Class: "Mystic"})
	addNewInvestigator(ctx, db, investigatorInput{Name: "William Yorick", Class: "Survivor"})
	addNewInvestigator(ctx, db, investigatorInput{Name: "Lola Hayes", Class: "Neutral"})

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

	if *populateDB {
		if err = populateDatabase(db); err != nil {
			log.Fatal(err)
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
