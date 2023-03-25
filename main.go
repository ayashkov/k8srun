package main

import (
	"os"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var logger *logrus.Logger = logrus.New()

func main() {
	if err := runCommand().Execute(); err != nil {
		os.Exit(1)
	}
}

func runCommand() *cobra.Command {
	var kubeconfig string

	job := Job{
		Instance: os.Getenv("AUTOSERV"),
		Name:     os.Getenv("AUTO_JOB_NAME"),
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
				logger.Fatal(
					"both AUTOSERV and AUTO_JOB_NAME environment variables are required")
			}

			job.Template = args[0]
			job.Args = args[1:]

			cluster := NewCluster(kubeconfig)
			exitCode, err := cluster.Run(&job, os.Stdout)

			if err != nil {
				logger.Error(err)
			}

			os.Exit(exitCode)
		},
	}

	cmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig", "",
		"Kubernetes client configuration file")
	cmd.PersistentFlags().StringVarP(&job.Namespace, "namespace", "n", "",
		"The namespace for creating the pod")

	return cmd
}
