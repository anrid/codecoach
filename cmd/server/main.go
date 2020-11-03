package main

import (
	"os"

	"github.com/anrid/codecoach/internal/config"
	user_ctrl "github.com/anrid/codecoach/internal/controller/user"
	"github.com/anrid/codecoach/internal/pg"
	account_dao "github.com/anrid/codecoach/internal/pg/dao/account"
	user_dao "github.com/anrid/codecoach/internal/pg/dao/user"
	"github.com/anrid/codecoach/internal/pkg/httpserver"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

func main() {
	// Setup global zap logger.
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	// Handle flags.
	initDB := pflag.Bool("init-db", false, "Drop and recreate database")

	pflag.Parse()

	// Load config.
	c := config.New()

	// Connect to and initialize db.
	if *initDB {
		pg.InitDB(c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName)
		os.Exit(0)
	}

	// Connect to database.
	db := pg.Connect(c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName)
	defer db.Close()

	// Setup DAOs.
	userDAO := user_dao.New(db)
	accountDAO := account_dao.New(db)

	// Setup HTTP server.
	serv := httpserver.New()

	// Setup controllers.
	userCtrl := user_ctrl.New(accountDAO, userDAO, c)

	// Setup routes.
	userCtrl.SetupRoutes(serv)

	// Start server.
	serv.Start(c.Host)
}
