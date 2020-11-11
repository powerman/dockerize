package main

import (
	"context"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/powerman/check"

	// Get nice diff in web UI.
	_ "github.com/smartystreets/goconvey/convey"
)

var (
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
	f := c.NoErrFile(ioutil.TempFile("", "gotest"))
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
