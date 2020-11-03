package main

import (
	"fmt"

	"github.com/anrid/codecoach/e2e"
	"github.com/anrid/codecoach/internal/config"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

type testingT struct{}

func (testingT) Errorf(format string, args ...interface{}) {
	fmt.Printf(format, args...)
}

func (testingT) FailNow() {
	// https://www.youtube.com/watch?v=RlnlDKznIaw <-- It's NECK n' NECK!
	panic("B00M!! TETRIS FOR JEFF!")
}

func main() {
	c := config.New()

	defaultHost := c.Host
	if defaultHost[0:1] == `:` {
		defaultHost = "localhost" + defaultHost
	}

	host := pflag.String("host", defaultHost, "CodeCoach API server host, default: config.Host")

	pflag.Parse()

	t := new(testingT)
	r := require.New(t)

	e2e.AllTests(r, e2e.Options{
		Host:                *host,
		TestTokenExpiration: false,
	})

	fmt.Printf("ALL TESTS PASS!\n")
}
