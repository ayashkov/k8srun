package runner

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/ayashkov/k8srun/service"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	typedCore "k8s.io/client-go/kubernetes/typed/core/v1"
)

type Execution struct {
	Job  *Job
	Pods typedCore.PodInterface
	Pod  *core.Pod
}

func (execution *Execution) CopyLogs(ctx context.Context, dst io.Writer) error {
	err := wait.PollImmediate(2*time.Second, time.Minute, func() (done bool, err error) {
		pod, err := execution.Pods.Get(ctx, execution.Pod.Name,
			meta.GetOptions{})

		if err != nil {
			return false, err
		}

		execution.Pod = pod

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

	log, err := execution.Pods.GetLogs(execution.Pod.Name,
		&core.PodLogOptions{Follow: true}).Stream(ctx)

	if err != nil {
		return err
	}

	defer log.Close()

	_, err = io.Copy(dst, log)

	return err
}

func (execution *Execution) WaitForCompletion(ctx context.Context) (int, error) {
	var exitCode int

	err := wait.PollImmediate(2*time.Second, time.Minute, func() (done bool, err error) {
		pod, err := execution.Pods.Get(ctx, execution.Pod.Name,
			meta.GetOptions{})

		if err != nil {
			return false, err
		}

		execution.Pod = pod

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

func (execution *Execution) Delete(ctx context.Context) error {
	if execution.Pod == nil {
		return nil
	}

	err := execution.Pods.Delete(ctx, execution.Pod.Name,
		meta.DeleteOptions{})

	if err != nil {
		return fmt.Errorf("error deleting pod %q in %q namespace: %w",
			execution.Pod.Name, execution.Pod.Namespace, err)
	}

	service.Log.Infof("deleted pod %q in %q namespace",
		execution.Pod.Name, execution.Pod.Namespace)

	execution.Pod = nil

	return nil
}
