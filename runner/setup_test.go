package runner

import (
	"testing"

	"github.com/ayashkov/k8srun/service"
	"github.com/sirupsen/logrus/hooks/test"
)

var loggerHook *test.Hook

func TestMain(m *testing.M) {
	service.Logger, loggerHook = test.NewNullLogger()
	m.Run()
}
