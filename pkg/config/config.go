package config

import (
	serving_clientset "github.com/knative/serving/pkg/client/clientset/versioned"
	rook_clientset "github.com/rook/rook/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
)

var (
	KubeClientset    *kubernetes.Clientset
	ServingClientset *serving_clientset.Clientset
	RookClientset    *rook_clientset.Clientset
	BuildTemplate    string
)
