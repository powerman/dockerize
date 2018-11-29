package main

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/powerman/check"
	_ "github.com/smartystreets/goconvey/convey" // get nice diff in web UI
)

var ( // nolint:gochecknoglobals
	testTimeFactor = floatGetenv("GO_TEST_TIME_FACTOR", 1.0)
	testSecond     = time.Duration(float64(time.Second) * testTimeFactor)
	testCtx        = context.Background()
)

func floatGetenv(name string, def float64) float64 {
	if v, err := strconv.ParseFloat(os.Getenv(name), 64); err == nil {
		return v
	}
	return def
}

type checkC struct{ *check.C }

func checkT(t *testing.T) *checkC                          { return &checkC{C: check.T(t)} }
func (c *checkC) NoErr(_ interface{}, err error)           { c.Helper(); c.Must(c.Nil(err)) }
func (c *checkC) NoErrFile(v *os.File, err error) *os.File { c.Helper(); c.Must(c.Nil(err)); return v }

func TestMain(m *testing.M) {
	check.TestMain(m)
}
