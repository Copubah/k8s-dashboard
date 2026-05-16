package kube

import (
	"fmt"

	"k8s-dashboard/backend/internal/config"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Client wraps client-go so HTTP handlers do not need to know how cluster
// credentials are discovered.
type Client struct {
	Set       kubernetes.Interface
	Namespace string
}

func NewClient(cfg config.Config) (*Client, error) {
	restConfig, err := buildRESTConfig(cfg)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("build clientset: %w", err)
	}

	return &Client{Set: clientset, Namespace: cfg.Namespace}, nil
}

func buildRESTConfig(cfg config.Config) (*rest.Config, error) {
	if cfg.InCluster {
		restConfig, err := rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("load in-cluster config: %w", err)
		}
		return restConfig, nil
	}

	loadingRules := &clientcmd.ClientConfigLoadingRules{ExplicitPath: cfg.Kubeconfig}
	overrides := &clientcmd.ConfigOverrides{CurrentContext: cfg.KubeContext}
	clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, overrides)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}
	return restConfig, nil
}

func NewFakeClient(namespace string, clientset kubernetes.Interface) *Client {
	return &Client{Set: clientset, Namespace: namespace}
}
