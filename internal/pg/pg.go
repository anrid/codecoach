package pg

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/anrid/codecoach/internal/config"
	"github.com/anrid/codecoach/internal/pg/dao/account"
	"github.com/anrid/codecoach/internal/pg/dao/user"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Postgres driver.
	"go.uber.org/zap"
)

// InitDB ...
func InitDB(c *config.Config, drop bool) *sqlx.DB {
	db := Connect(c, "")

	var databases []string
	err := db.Select(&databases, `SHOW DATABASES`)
	if err != nil {
		panic(err)
	}

	var found bool
	for _, d := range databases {
		if d == c.DBName {
			zap.S().Infof("found existing database %s", c.DBName)
			found = true
			break
		}
	}

	if found && drop {
		// Drop existing database.
		zap.S().Infof("dropping existing database %s", c.DBName)
		db.MustExec(`DROP DATABASE ` + c.DBName)
		found = false
	}

	if !found {
		// Create new database.
		zap.S().Infof("creating new database %s", c.DBName)
		db.MustExec(`CREATE DATABASE ` + c.DBName)
	}

	// Reconnect.
	db.Close()
	db = Connect(c, c.DBName)

	if !found {
		// Create tables.
		var created int

		accountDAO := account.New(db)
		userDAO := user.New(db)

		created += accountDAO.CreateTable()
		created += userDAO.CreateTable()

		zap.S().Infof("created database %s and %d tables", c.DBName, created)
	}

	return db
}

// RunPinger ...
func RunPinger(db *sqlx.DB) {
	// Ping in a loop.
	t := time.NewTicker(3 * time.Second)

	go func() {
		var failedPings int
		for {
			<-t.C
			err := db.Ping()
			if err != nil {
				failedPings++
				zap.S().Infow("db ping failed", "total", failedPings, "error", err.Error())
			} else {
				if failedPings > 0 {
					zap.S().Info("db ping successful")
				}
				failedPings = 0
			}
		}
	}()
}

// Connect ...
func Connect(c *config.Config, name string) (db *sqlx.DB) {
	opts := fmt.Sprintf("host=%s port=%s user=%s password=%s sslmode=disable", c.DBHost, c.DBPort, c.DBUser, c.DBPass)
	if name != "" {
		opts += " dbname=" + name
	}

	start := time.Now()
	wait := start.Add(30 * time.Second)
	var err error

	for {
		db, err = sqlx.Connect("postgres", opts)
		if err == nil {
			// No error? Great, we're good to go.
			break
		}

		if err == io.EOF {
			// Ignore EOF errors.
		} else if strings.Contains(err.Error(), "connection refused") {
			// Ignore `connection refused` errors.
		} else {
			// Panic on all other errors.
			zap.S().Panicw("could not connect to database", "error", err)
		}

		// Stay a while, and listen.
		zap.S().Infow("waiting for DB to start", "delay", 3, "opts", opts, "total_wait", time.Since(start))
		time.Sleep(3 * time.Second)

		// Don't stay too long though, that'd be rude!
		if wait.Before(time.Now()) {
			zap.S().Panicw("waited too long to connect to database", "opts", opts, "total_wait", time.Since(start))
		}
	}

	// Sane defaults.
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db
}
