package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"

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
			image := args[0]
			command := args[1:]

			fmt.Println("Kubeconfig", kubeconfig)
			fmt.Println("Namespace", namespace)
			fmt.Println("Job", job)
			fmt.Println("Image", image)
			fmt.Println("Command", command)

			clientset := createClientset(kubeconfig)

			createPod(clientset, namespace, job, image, command)
		},
	}

	cmd.PersistentFlags().StringVar(&kubeconfig, "kubeconfig",
		filepath.Join(home(), ".kube", "config"),
		"Kubernetes client configuration file")
	cmd.PersistentFlags().StringVarP(&namespace, "namespace", "n", "",
		"The namespace for creating the pod")
	cmd.Flags().StringVarP(&job, "job", "j", os.Getenv("AUTO_JOB_NAME"),
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

func createClientset(kubeconfig string) *kubernetes.Clientset {
	clientset, err := kubernetes.NewForConfig(configure(kubeconfig))

	if err != nil {
		panic(err.Error())
	}

	return clientset
}

func configure(kubeconfig string) *rest.Config {
	// rules := clientcmd.NewDefaultClientConfigLoadingRules()

	// rules.ExplicitPath = ""

	// overrides := &clientcmd.ConfigOverrides{}
	// k8s := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules,
	// 	overrides)

	// fmt.Println(k8s.ClientConfig())

	// os.Exit(1)

	config, err := rest.InClusterConfig()

	if err == rest.ErrNotInCluster {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	}

	if err != nil {
		panic(err.Error())
	}

	return config
}

func createPod(clientset *kubernetes.Clientset, namespace string,
	name string, image string, command []string) {
	podClient := clientset.CoreV1().Pods(meta.NamespaceDefault)
	pod := &core.Pod{
		ObjectMeta: meta.ObjectMeta{
			GenerateName: normalize(name) + "-",
			Namespace:    namespace,
			Labels: map[string]string{
				"job": name,
			},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:            "job",
					Image:           image,
					Command:         command,
					ImagePullPolicy: core.PullAlways,
				},
			},
			RestartPolicy: core.RestartPolicyNever,
		},
	}

	result, err := podClient.Create(context.TODO(), pod, meta.CreateOptions{})

	if err != nil {
		panic(err)
	}

	fmt.Printf("Created pod %v/%v.\n", result.GetObjectMeta().GetNamespace(),
		result.GetObjectMeta().GetName())
}
