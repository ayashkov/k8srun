package main

//go:generate mockgen --package mock --destination mock/runner.go --source runner/runner.go
//go:generate mockgen --package mock --destination mock/runner_factory.go --source runner/runner_factory.go
//go:generate mockgen --package mock --destination mock/client_config.go k8s.io/client-go/tools/clientcmd ClientConfig
//go:generate mockgen --package mock --destination mock/k8s_interface.go k8s.io/client-go/kubernetes Interface
//go:generate mockgen --package mock --destination mock/pod_interface.go k8s.io/client-go/kubernetes/typed/core/v1 PodInterface

import (
	"context"

	"github.com/ayashkov/k8srun/runner"
	"github.com/ayashkov/k8srun/service"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var runnerFactory = runner.NewRunnerFactory(
	clientcmd.NewNonInteractiveDeferredLoadingClientConfig,
	func(c *rest.Config) (kubernetes.Interface, error) {
		return kubernetes.NewForConfig(c)
	})

func main() {
	if err := newRunCommand().Execute(); err != nil {
		service.Os.Exit(1)
	}
}

func newRunCommand() *cobra.Command {
	var kubeconfig string

	job := runner.Job{
		Instance: service.Os.Getenv("AUTOSERV"),
		Name:     service.Os.Getenv("AUTO_JOB_NAME"),
	}
	cmd := &cobra.Command{
		Use:   "k8srun [flags] template [-- args ...]",
		Short: "AutoSys to Kubernetes bridge",
		Long: `This is a bridge between an AutoSys scheduler and
a Kubernetes cluster. The goal is to be able to
execute Kubernetes workload from AutoSys jobs.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			if job.Instance == "" || job.Name == "" {
				service.Log.Fatal(
					"both AUTOSERV and AUTO_JOB_NAME environment variables are required")
			}

			job.Template = args[0]
			job.Args = args[1:]

			cluster := runnerFactory.New(kubeconfig)
			exitCode, err := cluster.Run(context.Background(), &job,
				service.Os.Stdout())

			if err != nil {
				service.Log.Error(err)
				service.Os.Exit(128)
			}

			service.Os.Exit(exitCode)
		},
	}

	cmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "",
		"Kubernetes client configuration file")
	cmd.PersistentFlags().StringVarP(&job.Namespace, "namespace", "n", "",
		"The namespace for creating the pod")
	cmd.SetArgs(service.Os.Args()[1:])
	cmd.SetErr(service.Os.Stderr())
	cmd.SetIn(service.Os.Stdin())
	cmd.SetOut(service.Os.Stdout())

	return cmd
}
