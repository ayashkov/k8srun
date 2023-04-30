package runner

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8sClient interface {
	NewClientConfig(loader clientcmd.ClientConfigLoader,
		overrides *clientcmd.ConfigOverrides) clientcmd.ClientConfig
	NewClientset(c *rest.Config) (kubernetes.Interface, error)
}

type defaultK8sClient struct{}

func (defaultK8sClient) NewClientConfig(loader clientcmd.ClientConfigLoader,
	overrides *clientcmd.ConfigOverrides) clientcmd.ClientConfig {
	return clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader,
		overrides)
}

func (defaultK8sClient) NewClientset(c *rest.Config) (kubernetes.Interface, error) {
	return kubernetes.NewForConfig(c)
}
