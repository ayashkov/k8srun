package runner_test

import (
	"testing"

	"github.com/ayashkov/k8srun/service"
	"github.com/sirupsen/logrus/hooks/test"
)

var logger *test.Hook

func TestMain(m *testing.M) {
	service.Log, logger = test.NewNullLogger()
	m.Run()
}
