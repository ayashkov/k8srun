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

	job := Job{}
	cmd := &cobra.Command{
		Use:   "go-study [flags] image [-- command [args...]]",
		Short: "AutoSys to Kubernetes bridge",
		Long: `This is an attempt to implement a bridge between AutoSys
scheduler and a Kubernetes cluster. The goal is to be able
to execute Kubernetes workload from AutoSys jobs.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			job.Image = args[0]
			job.Command = args[1:]

			fmt.Println("Kubeconfig:", kubeconfig)
			fmt.Println("Namespace:", job.Namespace)
			fmt.Println("Job:", job.Name)
			fmt.Println("Image:", job.Image)
			fmt.Println("Command:", job.Command)

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
	cmd.PersistentFlags().StringVarP(&(job.Namespace), "namespace", "n",
		"", "The namespace for creating the pod")
	cmd.Flags().StringVarP(&(job.Name), "job", "j",
		os.Getenv("AUTO_JOB_NAME"),
		"The job name to use for naming the pod")

	return cmd
}
