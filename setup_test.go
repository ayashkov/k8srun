package main

import (
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
)

var loggerHook *test.Hook

func TestMain(m *testing.M) {
	logger, loggerHook = test.NewNullLogger()
	m.Run()
}
