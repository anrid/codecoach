package main

import (
	"github.com/anrid/codecoach/internal/config"
	oauth_c "github.com/anrid/codecoach/internal/controller/oauth"
	user_c "github.com/anrid/codecoach/internal/controller/user"
	"github.com/anrid/codecoach/internal/pg"
	account_d "github.com/anrid/codecoach/internal/pg/dao/account"
	user_d "github.com/anrid/codecoach/internal/pg/dao/user"
	"github.com/anrid/codecoach/internal/pkg/httpserver"
	github_oauth "github.com/anrid/codecoach/internal/usecase/github"
	user_uc "github.com/anrid/codecoach/internal/usecase/user"
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
	initDB := pflag.Bool("init-db", false, "Drop and recreate database and all tables")

	pflag.Parse()

	// Load config.
	c := config.New()

	// Connect to and initialize db.
	db := pg.InitDB(c, *initDB)
	defer db.Close()

	pg.RunPinger(db)

	// Setup DAOs.
	accountDAO := account_d.New(db)
	userDAO := user_d.New(db)

	// Setup use cases.
	oauthUC := github_oauth.New(c)
	userUC := user_uc.New(c, accountDAO, userDAO)

	// Setup HTTP server.
	serv := httpserver.New(userDAO)

	// Setup controllers.
	userCtrl := user_c.New(userUC)
	oauthCtrl := oauth_c.New(oauthUC)

	// Setup routes.
	userCtrl.SetupRoutes(serv)
	oauthCtrl.SetupRoutes(serv)

	// Start server.
	serv.Start(c.Host)
}
