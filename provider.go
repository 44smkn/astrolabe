package main

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type Provider interface{}

type KubernetesProvider struct {
	Client *kubernetes.Clientset
}

func NewKubernetesProvider(kubeconfig string) (*KubernetesProvider, error) {
	if _, err := os.Stat(kubeconfig); os.IsNotExist(err) {
		return nil, errors.New(fmt.Sprintf("%s does not exist\n", kubeconfig))
	}
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &KubernetesProvider{Client: clientset}, nil
}
