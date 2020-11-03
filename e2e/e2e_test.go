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
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
)

const testHost = "localhost:10099"

type ts struct {
	suite.Suite
	db *sqlx.DB
	ud *user_dao.DAO
	ad *account_dao.DAO
	uc *ctrl.Controller
	s  *httpserver.HTTPServer
	l  *zap.Logger
}

func TestE2E(t *testing.T) {
	suite.Run(t, new(ts))
}

func (su *ts) SetupSuite() {
	// Setup global zap logger.
	su.l, _ = zap.NewDevelopment()
	zap.ReplaceGlobals(su.l)

	zap.S().Infow("setup")

	c := config.New()

	// Override server host for testing.
	c.Host = ":10099"

	// Override database schema for testing.
	c.DBName = "codecoach_test"

	// Override token expires at.
	c.TokenExpires = 3 * time.Second

	su.db = pg.InitDB(c.DBHost, c.DBPort, c.DBUser, c.DBPass, c.DBName)

	// Setup DAOs.
	su.ud = user_dao.New(su.db)
	su.ad = account_dao.New(su.db)

	// Setup HTTP server.
	su.s = httpserver.New()

	// Setup controller.
	su.uc = ctrl.New(su.ad, su.ud, c)

	// Setup routes.
	su.uc.SetupRoutes(su.s)

	// Start server.
	go func() {
		su.s.Echo.Start(c.Host)
	}()

	zap.S().Infow("setup complete")
}

func (su *ts) TearDownSuite() {
	zap.S().Infow("tear down")

	su.db.Close()
	su.s.Echo.Shutdown(context.Background())
	su.l.Sync()

	zap.S().Infow("tear down complete")
}

func (su *ts) TestAll() {
	r := su.Require()
	AllTests(r, Options{
		Host:                testHost,
		TestTokenExpiration: true,
	})
}
