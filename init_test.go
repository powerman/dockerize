package main

import (
	"testing"

	"github.com/powerman/check"
	_ "github.com/smartystreets/goconvey/convey" // get nice diff in web UI
)

func TestMain(m *testing.M) { check.TestMain(m) }
