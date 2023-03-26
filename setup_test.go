package main

import (
	"testing"

	"github.com/ayashkov/k8srun/mock"
	"github.com/ayashkov/k8srun/service"
	"github.com/sirupsen/logrus/hooks/test"
)

var logger *test.Hook

var mockOs *mock.MockOsServices

func TestMain(m *testing.M) {
	mockOs = mock.NewMockOsServices()
	service.Os = mockOs
	service.Log, logger = test.NewNullLogger()
	service.Log.ExitFunc = mockOs.Exit
	m.Run()
}
