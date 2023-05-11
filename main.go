package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/davidhintelmann/gin-postgresql/connect"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/pgxpool"
)

// user, password, database name for postgresql instance, and ssl mode
const user, dbname, sslmode = "david", "AdventureWorks2014", "disable"

// be careful not to expose your password to the public
var password string // "***REMOVED***"

// use background context globally to pass between functions
var ctx = context.Background()

// Persons represents data from AdventureWorks OLTP DB

type People struct {
	Persons []Person `json:"people"`
}

type Person struct {
	ID         int            `json:"id"`
	Title      sql.NullString `json:"title"`
	FirstName  string         `json:"firstname"`
	MiddleName sql.NullString `json:"middlename"`
	LastName   string         `json:"lastname"`
	Suffix     sql.NullString `json:"suffix"`
	Scode      string         `json:"scode"`
	Ccode      string         `json:"ccode"`
	State      string         `json:"state"`
	Country    string         `json:"country"`
}

type GinNames struct {
	ID        int            `json:"id"`
	FirstName string         `json:"firstname"`
	LastName  string         `json:"lastname"`
	Job       sql.NullString `json:"job"`
}

type Error struct {
	Err string `json:"error"`
}

// password.go in connect directory has a function 'ImportPassword()'
// which will return the password for local PostgreSQL instance
func init() {
	password = connect.ImportPassword()
}

func main() {
	log.SetFlags(log.Ldate | log.Lshortfile)
	router := gin.Default()
	router.GET("/people", getPersons)
	router.POST("/postnames", postNames)
	router.GET("/people/:country", getCountries)

	router.Run("localhost:8080")
}

func getPersons(c *gin.Context) {
	db, err := connect.ConnectPSQL(ctx, user, password, dbname, sslmode)

	if err != nil {
		log.Fatalf("error after connecting to PostgreSQL: %v\n", err)
	}
	defer db.Close()

	err = db.Ping(ctx)
	if err != nil {
		log.Fatalf("error after pinging PostgreSQL database: %v\n", err)
	}

	query := `SELECT Person.Person.BusinessEntityID
,Person.Person.Title
,Person.Person.FirstName
,Person.Person.MiddleName
,Person.Person.LastName
,Person.Person.Suffix
,Person.StateProvince.StateProvinceCode
,Person.StateProvince.CountryRegionCode
,Person.StateProvince.Name
,Person.CountryRegion.Name
FROM Person.Person
JOIN Person.BusinessEntityAddress ON Person.Person.BusinessEntityID = Person.BusinessEntityAddress.BusinessEntityID
JOIN Person.Address ON Person.BusinessEntityAddress.AddressID = Person.Address.AddressID
JOIN Person.StateProvince ON Person.Address.StateProvinceID = Person.StateProvince.StateProvinceID
JOIN Person.CountryRegion ON Person.StateProvince.CountryRegionCode = Person.CountryRegion.CountryRegionCode;`
	query = fmt.Sprintf(query)

	// Execute query
	rows, err := db.Query(ctx, query)
	if err != nil {
		log.Fatal("Error reading table: " + err.Error())
	}
	defer rows.Close()

	people := People{}
	// Iterate through the result set.
	for rows.Next() {
		var person Person
		// Get values from row.
		err = rows.Scan(
			&person.ID,
			&person.Title,
			&person.FirstName,
			&person.MiddleName,
			&person.LastName,
			&person.Suffix,
			&person.Scode,
			&person.Ccode,
			&person.State,
			&person.Country,
		)

		if err != nil {
			log.Fatal("Error reading rows: " + err.Error())
		}

		// people = append(people, person)
		// person, err = json.MarshalIndent(person, "", "\t")
		people.AddPerson(person)
	}
	c.IndentedJSON(http.StatusOK, people)
}

func (peeps *People) AddPerson(per Person) {
	peeps.Persons = append(peeps.Persons, per)
}

// func ConnectPSQL(ctx context.Context, user string, password string, dbname string, sslmode string) (*pgxpool.Pool, error) {
// 	fmt.Print("Connecting to postgresql...\n")
// 	connectionString := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", user, password, dbname, sslmode)
// 	dbpool, err := pgxpool.New(ctx, connectionString)
// 	if err != nil {
// 		log.Printf("error during intial connection: %v\n", err)
// 		return nil, err
// 	}
// 	defer dbpool.Close()
// 	return dbpool, nil
// }

func postNames(c *gin.Context) {
	var ginnames GinNames
	db, err := connect.ConnectPSQL(ctx, user, password, "GinTest", sslmode)

	if err != nil {
		log.Fatalf("error after ConnectPSQL() in postNames() function: %v", err)
	}
	defer db.Close()

	err = db.Ping(ctx)
	if err != nil {
		log.Fatalf("error after pinging PostgreSQL database with dbname 'GinTest': %v\n", err)
	}

	// Call BindJSON to bind the received JSONS
	if err := c.BindJSON(&ginnames); err != nil {
		log.Fatalf("error after BindJSON() in postNames() function: %v", err)
		return
	}

	var tsql string
	// Insert into postnames Table in the GinTest database
	if ginnames.Job.Valid {
		query := `INSERT INTO post.postnames
(id, firstname, lastname, job)
VALUES (%d, '%s', '%s', '%s')`

		tsql = fmt.Sprintf(query, ginnames.ID, ginnames.FirstName, ginnames.LastName, ginnames.Job.String)
	} else {
		query := `INSERT INTO post.postnames
(id, firstname, lastname)
VALUES (%d, '%s', '%s')`

		tsql = fmt.Sprintf(query, ginnames.ID, ginnames.FirstName, ginnames.LastName)
	}

	row := db.QueryRow(ctx, tsql)
	err = row.Scan()
	fmt.Println(err)
	// no rows in result set
	if strings.HasPrefix(err.Error(), "ERROR:") {
		log.Printf("Error posting table: %v\n", row.Scan())
		errMsg := fmt.Sprintf("id '%d' already exists, please try another one", ginnames.ID)
		c.IndentedJSON(http.StatusUnprocessableEntity, gin.H{"error": errMsg})
		return
	}
	c.IndentedJSON(http.StatusCreated, ginnames)
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getCountries(c *gin.Context) {
	country := strings.ToUpper(c.Param("country"))

	db, err := connect.ConnectPSQL(ctx, user, password, dbname, sslmode)

	if err != nil {
		log.Fatalf("error after connecting to PostgreSQL: %v\n", err)
	}
	defer db.Close()

	err = db.Ping(ctx)
	if err != nil {
		log.Fatalf("error after pinging PostgreSQL database: %v\n", err)
	}

	query := `SELECT Person.Person.BusinessEntityID
,Person.Person.Title
,Person.Person.FirstName
,Person.Person.MiddleName
,Person.Person.LastName
,Person.Person.Suffix
,Person.StateProvince.StateProvinceCode
,Person.StateProvince.CountryRegionCode
,Person.StateProvince.Name
,Person.CountryRegion.Name
FROM Person.Person
JOIN Person.BusinessEntityAddress ON Person.Person.BusinessEntityID = Person.BusinessEntityAddress.BusinessEntityID
JOIN Person.Address ON Person.BusinessEntityAddress.AddressID = Person.Address.AddressID
JOIN Person.StateProvince ON Person.Address.StateProvinceID = Person.StateProvince.StateProvinceID
JOIN Person.CountryRegion ON Person.StateProvince.CountryRegionCode = Person.CountryRegion.CountryRegionCode
WHERE Person.StateProvince.CountryRegionCode = '%s';`
	query = fmt.Sprintf(query, country)

	// Execute query
	rows, err := db.Query(ctx, query)
	if err != nil {
		log.Fatal("Error reading table: " + err.Error())
	}
	defer rows.Close()

	people := People{}
	// Iterate through the result set.
	for rows.Next() {
		var person Person
		// Get values from row.
		err = rows.Scan(
			&person.ID,
			&person.Title,
			&person.FirstName,
			&person.MiddleName,
			&person.LastName,
			&person.Suffix,
			&person.Scode,
			&person.Ccode,
			&person.State,
			&person.Country,
		)

		if err != nil {
			log.Fatal("Error reading rows: " + err.Error())
		}

		// people = append(people, person)
		// person, err = json.MarshalIndent(person, "", "\t")
		people.AddPerson(person)
	}
	c.IndentedJSON(http.StatusOK, people)
}
