package runner

import (
	"k8s.io/client-go/tools/clientcmd"
)

type RunnerFactory interface {
	New(kubeconfig string) (Runner, error)
}

type defaultRunnerFactory struct{}

var Client K8sClient = defaultK8sClient{}

func NewRunnerFactory() RunnerFactory {
	return &defaultRunnerFactory{}
}

func (factory *defaultRunnerFactory) New(kubeconfig string) (Runner, error) {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()

	rules.ExplicitPath = kubeconfig

	clientConfig := Client.NewClientConfig(rules,
		&clientcmd.ConfigOverrides{})
	namespace, _, err := clientConfig.Namespace()

	if err != nil {
		panic(err.Error())
	}

	config, err := clientConfig.ClientConfig()

	if err != nil {
		panic(err.Error())
	}

	clientset, err := Client.NewClientset(config)

	if err != nil {
		panic(err.Error())
	}

	return &defaultRunner{
		clentset:  clientset,
		namespace: namespace,
	}, nil
}
