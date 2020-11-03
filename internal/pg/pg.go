package pg

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/anrid/codecoach/internal/pg/dao/account"
	"github.com/anrid/codecoach/internal/pg/dao/user"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Postgres driver.
)

// InitDB ...
func InitDB(dbHost, dbPort, dbUser, dbPass, dbName string) *sqlx.DB {
	db := Connect(dbHost, dbPort, dbUser, dbPass, "" /* no database name */)

	db.MustExec(`DROP DATABASE IF EXISTS ` + dbName)
	db.MustExec(`CREATE DATABASE ` + dbName)

	db.Close()

	db = Connect(dbHost, dbPort, dbUser, dbPass, dbName)

	// Create tables.
	accountDAO := account.New(db)
	userDAO := user.New(db)

	accountDAO.CreateTable()
	userDAO.CreateTable()

	log.Println("Created database and tables")
	return db
}

// Connect ...
func Connect(dbHost, dbPort, dbUser, dbPass, dbName string) (db *sqlx.DB) {
	opts := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass)
	if dbName != "" {
		opts += " dbname=" + dbName
	}

	start := time.Now()
	var err error

	for {
		db, err = sqlx.Connect("postgres", opts)
		if err != nil {
			if !strings.Contains(err.Error(), "connection refused") {
				panic(err)
			}
		} else {
			break
		}

		log.Printf("waiting %d seconds for DB to start ..", 3)
		time.Sleep(3 * time.Second)

		if start.Add(20 * time.Second).Before(time.Now()) {
			panic("Waited too long ..")
		}
	}

	return db
}
