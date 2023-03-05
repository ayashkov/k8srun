package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	if err := runCommand().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runCommand() *cobra.Command {
	var kubeconfig string
	var namespace string
	var job string
	var cmd = &cobra.Command{
		Use:   "go-study [flags] image [-- command [args...]]",
		Short: "AutoSys to Kubernetes bridge",
		Long: `This is an attempt to implement a bridge between AutoSys
scheduler and a Kubernetes cluster. The goal is to be able
to execute Kubernetes workload from AutoSys jobs.`,
		Args: cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			command := args[1:]

			fmt.Println("Kubeconfig", kubeconfig)
			fmt.Println("Namespace", namespace)
			fmt.Println("Job", job)
			fmt.Println("Image", args[0])
			fmt.Println("Command", command)

			clientset := clientset(kubeconfig)
			pods, err := clientset.CoreV1().Pods("").List(context.TODO(),
				metav1.ListOptions{})

			if err != nil {
				panic(err.Error())
			}

			fmt.Printf("There are %d pods in the cluster\n",
				len(pods.Items))
		},
	}

	cmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig",
		filepath.Join(home(), ".kube", "config"),
		"Kubernetes client configuration file")
	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "",
		"The namespace for creating the pod")
	cmd.Flags().StringVarP(&job, "job", "j",
		normalize(os.Getenv("AUTO_JOB_NAME")),
		"The job name to use for naming the pod")

	return cmd
}

func home() string {
	if home := homedir.HomeDir(); home != "" {
		return home
	}

	return "/"
}

func normalize(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(strings.ToLower(s)),
		"_", "-")
}

func clientset(kubeconfig string) *kubernetes.Clientset {
	clientset, err := kubernetes.NewForConfig(config(kubeconfig))

	if err != nil {
		panic(err.Error())
	}

	return clientset
}

func config(kubeconfig string) *rest.Config {
	config, err := rest.InClusterConfig()

	if err == rest.ErrNotInCluster {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if err != nil {
		panic(err.Error())
	}

	return config
}
