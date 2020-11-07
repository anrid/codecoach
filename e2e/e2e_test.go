package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/anrid/codecoach/internal/config"
	ctrl "github.com/anrid/codecoach/internal/controller/user"
	"github.com/anrid/codecoach/internal/pg"
	account_dao "github.com/anrid/codecoach/internal/pg/dao/account"
	user_dao "github.com/anrid/codecoach/internal/pg/dao/user"
	"github.com/anrid/codecoach/internal/pkg/httpserver"
	user_uc "github.com/anrid/codecoach/internal/usecase/user"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

const testHost = "localhost:10099"

type ts struct {
	suite.Suite
	db     *sqlx.DB
	aD     *account_dao.DAO
	uD     *user_dao.DAO
	serv   *httpserver.HTTPServer
	logger *zap.Logger
}

func TestE2E(t *testing.T) {
	suite.Run(t, new(ts))
}

func (su *ts) SetupSuite() {
	// Setup global zap logger.
	su.logger, _ = zap.NewDevelopment()
	zap.ReplaceGlobals(su.logger)

	zap.S().Infow("setup")

	c := config.New()

	// Override server host for testing.
	c.Host = ":10099"

	// Override database schema for testing.
	c.DBName = "codecoach_test"

	// Override token expires at.
	c.TokenExpires = 1 * time.Second

	su.db = pg.InitDB(c, true /* drop and recreate db every time */)

	// Setup DAOs.
	su.aD = account_dao.New(su.db)
	su.uD = user_dao.New(su.db)

	// Setup use cases.
	userUC := user_uc.New(c, su.aD, su.uD)

	// Setup HTTP server.
	su.serv = httpserver.New(su.uD)

	// Setup controller.
	userCtrl := ctrl.New(userUC)

	// Setup routes.
	userCtrl.SetupRoutes(su.serv)

	// Start server.
	go func() {
		_ = su.serv.Echo.Start(c.Host)
	}()

	zap.S().Infow("setup complete")
}

func (su *ts) TearDownSuite() {
	zap.S().Infow("tear down")

	su.db.Close()
	_ = su.serv.Echo.Shutdown(context.Background())
	_ = su.logger.Sync()

	zap.S().Infow("tear down complete")
}

func (su *ts) TestAll() {
	r := su.Require()
	AllTests(r, Options{
		Host:                testHost,
		TestTokenExpiration: true,
	})
}
