package main

import (
	"os"

	"github.com/anrid/codecoach/internal/config"
	oauth_c "github.com/anrid/codecoach/internal/controller/oauth"
	user_c "github.com/anrid/codecoach/internal/controller/user"
	"github.com/anrid/codecoach/internal/pg"
	account_d "github.com/anrid/codecoach/internal/pg/dao/account"
	user_d "github.com/anrid/codecoach/internal/pg/dao/user"
	"github.com/anrid/codecoach/internal/pkg/httpserver"
	github_oauth "github.com/anrid/codecoach/internal/usecase/github"
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
	userDAO := user_d.New(db)
	accountDAO := account_d.New(db)

	// Setup use cases.
	oauthUC := github_oauth.New(c)

	// Setup HTTP server.
	serv := httpserver.New()

	// Setup controllers.
	userCtrl := user_c.New(accountDAO, userDAO, c)
	oauthCtrl := oauth_c.New(oauthUC)

	// Setup routes.
	userCtrl.SetupRoutes(serv)
	oauthCtrl.SetupRoutes(serv)

	// Start server.
	serv.Start(c.Host)
}
