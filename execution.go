package main

import (
	"context"
	"io"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	coreAccessor "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Execution struct {
	job  *Job
	pods coreAccessor.PodInterface
	pod  *core.Pod
}

func (execution *Execution) CopyLogs(dst io.Writer) error {
	err := wait.PollImmediate(2*time.Second, time.Minute, func() (done bool, err error) {
		pod, err := execution.pods.Get(context.TODO(), execution.pod.Name,
			meta.GetOptions{})

		if err != nil {
			return false, err
		}

		execution.pod = pod

		phase := pod.Status.Phase

		if phase == core.PodRunning || phase == core.PodSucceeded ||
			phase == core.PodFailed {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return err
	}

	log, err := execution.pods.GetLogs(execution.pod.Name,
		&core.PodLogOptions{Follow: true}).Stream(context.TODO())

	if err != nil {
		return err
	}

	defer log.Close()

	_, err = io.Copy(dst, log)

	return err
}

func (execution *Execution) WaitForCompletion() (int, error) {
	var exitCode int

	err := wait.PollImmediate(2*time.Second, time.Minute, func() (done bool, err error) {
		pod, err := execution.pods.Get(context.TODO(), execution.pod.Name,
			meta.GetOptions{})

		if err != nil {
			return false, err
		}

		execution.pod = pod

		containerStatuses := pod.Status.ContainerStatuses

		if len(containerStatuses) == 0 {
			return false, nil
		}

		terminated := containerStatuses[0].State.Terminated

		if terminated != nil {
			exitCode = int(terminated.ExitCode)

			return true, nil
		}

		return false, nil
	})

	if err != nil {
		return 128, err
	}

	return exitCode, nil
}
