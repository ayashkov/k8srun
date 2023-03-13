package main

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

const INSTANCE = "k8srun.yashkov.org/instance"

const PREFIX = "k8srun.yashkov.org/prefix"

type Job struct {
	Instance  string
	Name      string
	Namespace string
	Template  string
	Args      []string
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
	template, err := cluster.getPodTemplate(job)

	if err != nil {
		return nil, err
	}

	execution := Execution{job: job}

	execution.pods = cluster.clentset.CoreV1().Pods(template.Namespace)

	def := &core.Pod{
		ObjectMeta: template.Template.ObjectMeta,
		Spec:       template.Template.Spec,
	}

	def.ObjectMeta.Namespace = ""
	def.ObjectMeta.Name = ""
	def.ObjectMeta.GenerateName = generateName(job.Name)
	def.Spec.Containers[0].Args = job.Args

	execution.pod, err = execution.pods.Create(context.TODO(), def,
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

func (cluster *Cluster) getPodTemplate(job *Job) (*core.PodTemplate, error) {
	namespace := job.Namespace

	if namespace == "" {
		namespace = cluster.namespace
	}

	template, err := cluster.clentset.CoreV1().PodTemplates(namespace).Get(
		context.TODO(), job.Template, meta.GetOptions{})

	if err != nil {
		return nil, err
	}

	if err = checkLabel(template, INSTANCE, strings.ToLower(job.Instance)); err != nil {
		return nil, err
	}

	if err = checkLabel(template, PREFIX, prefix(job.Name)); err != nil {
		return nil, err
	}

	return template, nil
}

func checkLabel(template *core.PodTemplate, name string, value string) error {
	if strings.ToLower(template.Labels[name]) != value {
		return fmt.Errorf("template label %v does not match %q", name,
			value)
	}

	return nil
}

func prefix(name string) string {
	re := regexp.MustCompile("^[[:alnum:]]+")

	return strings.ToLower(re.FindString(strings.TrimSpace(name)))
}

func generateName(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(strings.ToLower(s)), "_",
		"-") + "-"
}
