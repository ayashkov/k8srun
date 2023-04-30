package runner

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type RunnerFactory interface {
	New(kubeconfig string) Runner
}

type defaultRunnerFactory struct {
	newClientConfig func(loader clientcmd.ClientConfigLoader,
		overrides *clientcmd.ConfigOverrides) clientcmd.ClientConfig
	newClientSet func(c *rest.Config) (kubernetes.Interface, error)
}

func NewRunnerFactory(clientConfigCreator func(
	loader clientcmd.ClientConfigLoader,
	overrides *clientcmd.ConfigOverrides) clientcmd.ClientConfig,
	clientSetCreator func(c *rest.Config) (kubernetes.Interface, error)) RunnerFactory {
	return &defaultRunnerFactory{
		newClientConfig: clientConfigCreator,
		newClientSet:    clientSetCreator,
	}
}

func (factory *defaultRunnerFactory) New(kubeconfig string) Runner {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()

	rules.ExplicitPath = kubeconfig

	clientConfig := factory.newClientConfig(rules,
		&clientcmd.ConfigOverrides{})
	namespace, _, err := clientConfig.Namespace()

	if err != nil {
		panic(err.Error())
	}

	config, err := clientConfig.ClientConfig()

	if err != nil {
		panic(err.Error())
	}

	clientset, err := factory.newClientSet(config)

	if err != nil {
		panic(err.Error())
	}

	return &defaultRunner{
		clentset:  clientset,
		namespace: namespace,
	}
}
