package runner

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/ayashkov/k8srun/service"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const INSTANCE = "k8srun.yashkov.org/instance"

const PREFIX = "k8srun.yashkov.org/prefix"

type Runner interface {
	Run(ctx context.Context, job *Job, out io.Writer) (int, error)
	Start(ctx context.Context, job *Job) (*Execution, error)
}

type defaultRunner struct {
	clentset  kubernetes.Interface
	namespace string
}

func (runner *defaultRunner) Start(ctx context.Context,
	job *Job) (*Execution, error) {
	template, err := runner.getPodTemplate(ctx, job)

	if err != nil {
		return nil, err
	}

	execution := Execution{Job: job}

	execution.Pods = runner.clentset.CoreV1().Pods(template.Namespace)

	def := &core.Pod{
		ObjectMeta: template.Template.ObjectMeta,
		Spec:       template.Template.Spec,
	}

	def.ObjectMeta.Namespace = ""
	def.ObjectMeta.Name = ""
	def.ObjectMeta.GenerateName = generateName(job.Name)
	def.Spec.Containers[0].Args = job.Args

	execution.Pod, err = execution.Pods.Create(ctx, def, meta.CreateOptions{})

	if err != nil {
		return nil, err
	}

	service.Log.Infof("created pod %q in %q namespace",
		execution.Pod.Name, execution.Pod.Namespace)

	return &execution, nil
}

func (runner *defaultRunner) Run(ctx context.Context, job *Job,
	out io.Writer) (int, error) {
	execution, err := runner.Start(ctx, job)

	if err != nil {
		return -1, err
	}

	defer func() {
		if err := execution.Delete(ctx); err != nil {
			service.Log.Error(err)
		}
	}()

	err = execution.CopyLogs(ctx, out)

	if err != nil {
		return -1, err
	}

	return execution.WaitForCompletion(ctx)
}

func (runner *defaultRunner) getPodTemplate(ctx context.Context,
	job *Job) (*core.PodTemplate, error) {
	namespace := job.Namespace

	if namespace == "" {
		namespace = runner.namespace
	}

	template, err := runner.clentset.
		CoreV1().
		PodTemplates(namespace).
		Get(ctx, job.Template, meta.GetOptions{})

	if err != nil {
		return nil, err
	}

	if err = checkAnnotation(template, INSTANCE,
		strings.ToLower(job.Instance)); err != nil {
		return nil, err
	}

	if err = checkAnnotation(template, PREFIX, prefix(job.Name)); err != nil {
		return nil, err
	}

	if nConts := len(template.Template.Spec.Containers); nConts != 1 {
		return nil,
			fmt.Errorf("only one container per pod is supported, %q has %v",
				template.Name, nConts)
	}

	return template, nil
}

func checkAnnotation(template *core.PodTemplate, name string, value string) error {
	if strings.ToLower(template.Annotations[name]) != value {
		return fmt.Errorf("template annotation %v does not match %q",
			name, value)
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
