package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Job struct {
	Namespace string
	Name      string
	Image     string
	Command   []string
}

type Cluster struct {
	clentset  *kubernetes.Clientset
	namespace string
}

func NewCluster(kubeconfig string) *Cluster {
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

	return &Cluster{
		clentset:  clientset,
		namespace: namespace,
	}
}

func (cluster *Cluster) Start(job *Job) (*Execution, error) {
	execution := Execution{job: job}
	namespace := job.Namespace

	if namespace == "" {
		namespace = cluster.namespace
	}

	execution.pods = cluster.clentset.CoreV1().Pods(namespace)
	podDef := &core.Pod{
		ObjectMeta: meta.ObjectMeta{
			GenerateName: normalize(job.Name) + "-",
			Labels: map[string]string{
				"job": job.Name,
			},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:            "job",
					Image:           job.Image,
					Command:         job.Command,
					ImagePullPolicy: core.PullAlways,
				},
			},
			RestartPolicy: core.RestartPolicyNever,
		},
	}

	var err error

	execution.pod, err = execution.pods.Create(context.TODO(), podDef,
		meta.CreateOptions{})

	if err != nil {
		return nil, err
	}

	fmt.Printf("Created pod %q in %q namespace\n", execution.pod.Name,
		execution.pod.Namespace)

	return &execution, nil
}

func (cluster *Cluster) Run(job *Job, out io.Writer) (int, error) {
	execution, err := cluster.Start(job)

	if err != nil {
		return 128, err
	}

	defer execution.Delete()

	err = execution.CopyLogs(out)

	if err != nil {
		return 128, err
	}

	return execution.WaitForCompletion()
}

func normalize(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(strings.ToLower(s)),
		"_", "-")
}
