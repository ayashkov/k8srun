package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := runCommand().Execute(); err != nil {
		fmt.Println(err)
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
		Use:   "go-study [flags] template [-- args ...]",
		Short: "AutoSys to Kubernetes bridge",
		Long: `This is an attempt to implement a bridge between AutoSys
scheduler and a Kubernetes cluster. The goal is to be able
to execute Kubernetes workload from AutoSys jobs.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			job.Template = args[0]
			job.Args = args[1:]

			fmt.Println("Kubeconfig:", kubeconfig)
			fmt.Println("Instance:", job.Instance)
			fmt.Println("Job:", job.Name)
			fmt.Println("Namespace:", job.Namespace)
			fmt.Println("Template:", job.Template)
			fmt.Println("Args:", job.Args)

			if job.Instance == "" || job.Name == "" {
				fmt.Println("Both AUTOSERV and AUTO_JOB_NAME environment variables are required")

				os.Exit(1)
			}

			cluster := NewCluster(kubeconfig)
			exitCode, err := cluster.Run(&job, os.Stdout)

			if err != nil {
				fmt.Println("Error:", err.Error())
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
