package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	if err := runCommand().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

type workload struct {
	Namespace string
	Job       string
	Image     string
	Command   []string
}

type cluster struct {
	clentset  *kubernetes.Clientset
	namespace string
}

func config(kubeconfig string) *cluster {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()

	rules.ExplicitPath = kubeconfig

	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		rules, &clientcmd.ConfigOverrides{})
	namespace, _, err := clientConfig.Namespace()

	if err != nil {
		panic(err.Error())
	}

	config, err := clientConfig.ClientConfig()

	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		panic(err.Error())
	}

	return &cluster{
		clentset:  clientset,
		namespace: namespace,
	}
}

func (cluster *cluster) run(workload *workload) int {
	namespace := workload.Namespace

	if namespace == "" {
		namespace = cluster.namespace
	}

	pods := cluster.clentset.CoreV1().Pods(namespace)
	podDef := &core.Pod{
		ObjectMeta: meta.ObjectMeta{
			GenerateName: normalize(workload.Job) + "-",
			Labels: map[string]string{
				"job": workload.Job,
			},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:            "job",
					Image:           workload.Image,
					Command:         workload.Command,
					ImagePullPolicy: core.PullAlways,
				},
			},
			RestartPolicy: core.RestartPolicyNever,
		},
	}

	created, err := pods.Create(context.TODO(), podDef, meta.CreateOptions{})

	if err != nil {
		panic(err)
	}

	fmt.Printf("Created pod %v/%v.\n", created.GetObjectMeta().GetNamespace(),
		created.GetObjectMeta().GetName())

	err = wait.PollImmediate(2*time.Second, time.Minute, func() (done bool, err error) {
		pod, err := pods.Get(context.TODO(), created.Name, meta.GetOptions{})

		if err != nil {
			return false, err
		}

		phase := pod.Status.Phase

		if phase == core.PodRunning || phase == core.PodSucceeded ||
			phase == core.PodFailed {
			return true, nil
		}

		return false, nil
	})

	if err != nil {
		panic(err)
	}

	log, err := pods.GetLogs(created.Name, &core.PodLogOptions{
		Follow: true,
	}).Stream(context.TODO())

	if err != nil {
		panic(err)
	}

	defer log.Close()

	_, err = io.Copy(os.Stdout, log)

	if err != nil {
		panic(err.Error())
	}

	var exitCode int

	err = wait.PollImmediate(2*time.Second, time.Minute, func() (done bool, err error) {
		pod, err := pods.Get(context.TODO(), created.Name, meta.GetOptions{})

		if err != nil {
			return false, err
		}

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
		panic(err)
	}

	return exitCode
}

func runCommand() *cobra.Command {
	var kubeconfig string

	workload := workload{}
	cmd := &cobra.Command{
		Use:   "go-study [flags] image [-- command [args...]]",
		Short: "AutoSys to Kubernetes bridge",
		Long: `This is an attempt to implement a bridge between AutoSys
scheduler and a Kubernetes cluster. The goal is to be able
to execute Kubernetes workload from AutoSys jobs.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			workload.Image = args[0]
			workload.Command = args[1:]

			fmt.Println("Kubeconfig:", kubeconfig)
			fmt.Println("Namespace:", workload.Namespace)
			fmt.Println("Job:", workload.Job)
			fmt.Println("Image:", workload.Image)
			fmt.Println("Command:", workload.Command)

			cluster := config(kubeconfig)

			os.Exit(cluster.run(&workload))
		},
	}

	cmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig",
		"", "Kubernetes client configuration file")
	cmd.PersistentFlags().StringVarP(&(workload.Namespace), "namespace",
		"n", "", "The namespace for creating the pod")
	cmd.Flags().StringVarP(&(workload.Job), "job", "j",
		os.Getenv("AUTO_JOB_NAME"),
		"The job name to use for naming the pod")

	return cmd
}

func normalize(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(strings.ToLower(s)),
		"_", "-")
}
