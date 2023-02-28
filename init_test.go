package main

import (
	"context"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/powerman/check"
	_ "github.com/smartystreets/goconvey/convey"
)

var (
	testTimeFactor = floatGetenv("GO_TEST_TIME_FACTOR", 1.0)
	testSecond     = time.Duration(testTimeFactor) * time.Second //nolint:revive // By design.
	testCtx        = context.Background()
)

func floatGetenv(name string, def float64) float64 {
	if v, err := strconv.ParseFloat(os.Getenv(name), 64); err == nil {
		return v
	}
	return def
}

type checkC struct{ *check.C }

func checkT(t *testing.T) *checkC                          { t.Helper(); return &checkC{C: check.T(t)} }
func (c *checkC) NoErr(_ any, err error)                   { c.Helper(); c.Must(c.Nil(err)) }
func (c *checkC) NoErrInt(v int, err error) int            { c.Helper(); c.Must(c.Nil(err)); return v }
func (c *checkC) NoErrBuf(v []byte, err error) []byte      { c.Helper(); c.Must(c.Nil(err)); return v }
func (c *checkC) NoErrFile(v *os.File, err error) *os.File { c.Helper(); c.Must(c.Nil(err)); return v }
func (c *checkC) NoErrListen(v net.Listener, err error) net.Listener {
	c.Helper()
	c.Must(c.Nil(err))
	return v
}

func (c *checkC) TempPath() string {
	c.Helper()
	f := c.NoErrFile(os.CreateTemp("", "gotest"))
	c.Nil(f.Close())
	c.Nil(os.Remove(f.Name()))
	return f.Name()
}

func TestMain(m *testing.M) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") == "" { // don't do this again in subprocess
		// Vars for use in testdata/ templates.
		os.Setenv("A", "10")
		os.Setenv("B", "20")
		os.Unsetenv("C")
	}

	check.TestMain(m)
}
