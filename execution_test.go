package main

import (
	"context"
	"fmt"
	"testing"

	"github.com/ayashkov/k8srun/mocks"
	gomock "github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_Execution_Delete_DeletesPod_WhenPodIsProvided(t *testing.T) {
	pods := mocks.NewMockPodInterface(gomock.NewController(t))
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

	assert.Nil(t, execution.Delete())
	assert.Equal(t, 1, len(loggerHook.Entries))
	assert.Equal(t, logrus.InfoLevel, loggerHook.LastEntry().Level)
	assert.Equal(t, "deleted pod \"delete-me\" in \"namespace\" namespace",
		loggerHook.LastEntry().Message)

	loggerHook.Reset()
}

func Test_Execution_Delete_DoesNothing_WhenNoPodIsProvided(t *testing.T) {
	pods := mocks.NewMockPodInterface(gomock.NewController(t))
	execution := Execution{
		pod:  nil,
		pods: pods,
	}

	assert.Nil(t, execution.Delete())
	assert.Equal(t, 0, len(loggerHook.Entries))

	loggerHook.Reset()
}

func Test_Execution_Delete_ReturnsError_WhenDeleteFails(t *testing.T) {
	pods := mocks.NewMockPodInterface(gomock.NewController(t))
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

	assert.Error(t, execution.Delete(),
		"error deleting pod %q in %q namespace: delete error",
		"delete-me", "namespace")
	assert.Equal(t, 0, len(loggerHook.Entries))

	loggerHook.Reset()
}
