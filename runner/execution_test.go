package runner_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ayashkov/k8srun/mock"
	"github.com/ayashkov/k8srun/runner"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var ctx = context.Background()

func setUp(t *testing.T) *assert.Assertions {
	t.Cleanup(logger.Reset)

	return assert.New(t)
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
