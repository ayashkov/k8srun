package runner

import (
	"context"
	"fmt"
	"testing"

	"github.com/ayashkov/k8srun/mock"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func setUp(t *testing.T) *assert.Assertions {
	logger.Reset()

	return assert.New(t)
}

func Test_Execution_Delete_DeletesPod_WhenPodIsProvided(t *testing.T) {
	assert := setUp(t)
	pods := mock.NewMockPodInterface(gomock.NewController(t))
	execution := Execution{
		pod: &core.Pod{
			ObjectMeta: meta.ObjectMeta{
				Name:      "delete-me",
				Namespace: "namespace",
			},
		},
		pods: pods,
	}

	pods.EXPECT().
		Delete(context.TODO(), "delete-me", meta.DeleteOptions{})

	assert.Nil(execution.Delete())

	assert.Equal(1, len(logger.Entries))
	assert.Equal(logrus.InfoLevel, logger.LastEntry().Level)
	assert.Equal("deleted pod \"delete-me\" in \"namespace\" namespace",
		logger.LastEntry().Message)
}

func Test_Execution_Delete_DoesNothing_WhenNoPodIsProvided(t *testing.T) {
	assert := setUp(t)
	pods := mock.NewMockPodInterface(gomock.NewController(t))
	execution := Execution{
		pod:  nil,
		pods: pods,
	}

	assert.Nil(execution.Delete())

	assert.Empty(logger.Entries)
}

func Test_Execution_Delete_ReturnsError_WhenDeleteFails(t *testing.T) {
	assert := setUp(t)
	pods := mock.NewMockPodInterface(gomock.NewController(t))
	execution := Execution{
		pod: &core.Pod{
			ObjectMeta: meta.ObjectMeta{
				Name:      "delete-me",
				Namespace: "namespace",
			},
		},
		pods: pods,
	}

	pods.EXPECT().
		Delete(context.TODO(), "delete-me", meta.DeleteOptions{}).
		Return(fmt.Errorf("delete error"))

	assert.Error(execution.Delete(),
		"error deleting pod %q in %q namespace: delete error",
		"delete-me", "namespace")

	assert.Empty(logger.Entries)
}
