package runner_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ayashkov/k8srun/mock"
	"github.com/ayashkov/k8srun/runner"
	"github.com/ayashkov/k8srun/service"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var logger *test.Hook

var ctx = context.Background()

func TestMain(m *testing.M) {
	service.Log, logger = test.NewNullLogger()
	m.Run()
}

func setUp(t *testing.T) *assert.Assertions {
	t.Cleanup(logger.Reset)

	return assert.New(t)
}

func Test_RunnerFactory_New_CreatesRunner_Normally(t *testing.T) {
	assert := setUp(t)
	ctrl := gomock.NewController(t)
	clientConfig := mock.NewMockClientConfig(ctrl)
	clientSet := mock.NewMockInterface(ctrl)
	clientConfigCreator := func(loader clientcmd.ClientConfigLoader,
		overrides *clientcmd.ConfigOverrides) clientcmd.ClientConfig {
		return clientConfig
	}
	clientSetCreator := func(c *rest.Config) (kubernetes.Interface, error) {
		return clientSet, nil
	}
	factory := runner.NewRunnerFactory(clientConfigCreator, clientSetCreator)

	assert.NotNil(factory)

	clientConfig.EXPECT().
		Namespace().
		Return("test-namespace", false, nil)
	clientConfig.EXPECT().
		ClientConfig().
		Return(&rest.Config{}, nil)

	runner := factory.New("")

	assert.NotNil(runner)
}

func Test_Execution_Delete_DeletesPod_WhenPodIsProvided(t *testing.T) {
	assert := setUp(t)
	pods := mock.NewMockPodInterface(gomock.NewController(t))
	execution := runner.Execution{
		Pod: &core.Pod{
			ObjectMeta: meta.ObjectMeta{
				Name:      "delete-me",
				Namespace: "namespace",
			},
		},
		Pods: pods,
	}

	pods.EXPECT().
		Delete(ctx, "delete-me", meta.DeleteOptions{})

	assert.Nil(execution.Delete(ctx))

	assert.Equal(1, len(logger.Entries))
	assert.Equal(logrus.InfoLevel, logger.LastEntry().Level)
	assert.Equal("deleted pod \"delete-me\" in \"namespace\" namespace",
		logger.LastEntry().Message)
}

func Test_Execution_Delete_DoesNothing_WhenNoPodIsProvided(t *testing.T) {
	assert := setUp(t)
	pods := mock.NewMockPodInterface(gomock.NewController(t))
	execution := runner.Execution{
		Pod:  nil,
		Pods: pods,
	}

	assert.Nil(execution.Delete(ctx))

	assert.Empty(logger.Entries)
}

func Test_Execution_Delete_ReturnsError_WhenDeleteFails(t *testing.T) {
	assert := setUp(t)
	pods := mock.NewMockPodInterface(gomock.NewController(t))
	execution := runner.Execution{
		Pod: &core.Pod{
			ObjectMeta: meta.ObjectMeta{
				Name:      "delete-me",
				Namespace: "namespace",
			},
		},
		Pods: pods,
	}

	pods.EXPECT().
		Delete(ctx, "delete-me", meta.DeleteOptions{}).
		Return(fmt.Errorf("delete error"))

	assert.Error(execution.Delete(ctx),
		"error deleting pod %q in %q namespace: delete error",
		"delete-me", "namespace")

	assert.Empty(logger.Entries)
}
