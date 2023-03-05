package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	var runCommand = &cobra.Command{
		Use:   "go-study [flags] image [-- command [args...]]",
		Short: "AutoSys to Kubernetes bridge",
		Long: `This is an attempt to implement a bridge between AutoSys
scheduler and a Kubernetes cluster. The goal is to be able to execute
Kubernetes workload from AutoSys jobs.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			namespace := cmd.Flags().Lookup("namespace").Value.String()
			job := cmd.Flags().Lookup("job").Value.String()
			command := ""

			if len(args) > 1 {
				command = args[1]
			}

			fmt.Println("Namespace", namespace)
			fmt.Println("Job", job)
			fmt.Println("Image", args[0])
			fmt.Println("Command", command)
		},
	}

	runCommand.PersistentFlags().StringP("namespace", "n", "",
		"The namespace for creating the pod")
	runCommand.Flags().StringP("job", "j",
		normalize(os.Getenv("AUTO_JOB_NAME")),
		"The job name to use for naming the pod")

	if err := runCommand.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func normalize(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(strings.ToLower(s)),
		"_", "-")
}
