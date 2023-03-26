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
	job  *Job
	pods typedCore.PodInterface
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

func (execution *Execution) Delete() error {
	if execution.pod == nil {
		return nil
	}

	err := execution.pods.Delete(context.TODO(), execution.pod.Name,
		meta.DeleteOptions{})

	if err != nil {
		return fmt.Errorf("error deleting pod %q in %q namespace: %w",
			execution.pod.Name, execution.pod.Namespace, err)
	}

	service.Logger.Infof("deleted pod %q in %q namespace",
		execution.pod.Name, execution.pod.Namespace)

	execution.pod = nil

	return nil
}
