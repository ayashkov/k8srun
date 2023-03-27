package runner

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type RunnerFactory interface {
	New(kubeconfig string) Runner
}

type DefaultRunnerFactory struct{}

func (DefaultRunnerFactory) New(kubeconfig string) Runner {
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

	return &defaultRunner{
		clentset:  clientset,
		namespace: namespace,
	}
}
