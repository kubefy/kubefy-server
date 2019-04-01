//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package kfunc

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/glog"

	cfg "github.com/kubefy/kubefy-server/pkg/config"
	"github.com/kubefy/kubefy-server/pkg/model"

	build_api "github.com/knative/build/pkg/apis/build/v1alpha1"
	serving_api "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultContainerRegistry = "docker.io"
	defaultIstioNamespace    = "istio-system"
	defaultIstioGatewaySvc   = "istio-ingressgateway"
	defaultBuildTemplate     = "buildah"
	defaultNumNodeAddr       = 3
)

// DeploySrc2Svc deploys a git repo to a Knative Service
func DeploySrc2Svc(namespace, gitUrl, gitRevision, imageUrl, funcName string) error {
	if len(gitUrl) == 0 || len(funcName) == 0 || len(imageUrl) == 0 {
		return fmt.Errorf("git repo, imageUrl, or function name is missing")
	}

	if len(gitRevision) == 0 {
		gitRevision = "master"
	}
	buildTemplate := defaultBuildTemplate
	if len(cfg.BuildTemplate) != 0 {
		buildTemplate = cfg.BuildTemplate
	}

	svc := &serving_api.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      funcName,
			Namespace: namespace,
		},
		Spec: serving_api.ServiceSpec{
			RunLatest: &serving_api.RunLatestType{
				Configuration: serving_api.ConfigurationSpec{
					Build: &serving_api.RawExtension{
						Object: &build_api.Build{
							TypeMeta: metav1.TypeMeta{
								APIVersion: build_api.SchemeGroupVersion.String(),
								Kind:       "Build",
							},
							Spec: build_api.BuildSpec{
								Source: &build_api.SourceSpec{
									Git: &build_api.GitSourceSpec{
										Url:      gitUrl,
										Revision: gitRevision,
									},
								},
								Template: &build_api.TemplateInstantiationSpec{
									Name: buildTemplate,
									Arguments: []build_api.ArgumentSpec{
										build_api.ArgumentSpec{
											Name:  "IMAGE",
											Value: imageUrl,
										},
									},
								},
							},
						},
					},
					RevisionTemplate: serving_api.RevisionTemplateSpec{
						Spec: serving_api.RevisionSpec{
							Container: corev1.Container{
								Image: imageUrl,
							},
						},
					},
				},
			},
		},
	}

	_, err := cfg.ServingClientset.ServingV1alpha1().Services(namespace).Create(svc)

	return err
}

// DeployImg2Svc deploys a container image to a Knative Service
func DeployImg2Svc(namespace, imageUrl, funcName string) error {
	if len(imageUrl) == 0 || len(funcName) == 0 {
		return fmt.Errorf("container image or function name is missing")
	}

	svc := &serving_api.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      funcName,
			Namespace: namespace,
		},
		Spec: serving_api.ServiceSpec{
			RunLatest: &serving_api.RunLatestType{
				Configuration: serving_api.ConfigurationSpec{
					RevisionTemplate: serving_api.RevisionTemplateSpec{
						Spec: serving_api.RevisionSpec{
							Container: corev1.Container{
								Image: imageUrl,
							},
						},
					},
				},
			},
		},
	}

	_, err := cfg.ServingClientset.ServingV1alpha1().Services(namespace).Create(svc)

	return err
}

func View(namespace, funcName string) ([]model.Endpoint, string, error) {
	var (
		endpoints     []model.Endpoint
		authoriy      string
		nodeAddresses []string
	)

	if len(funcName) == 0 {
		return endpoints, authoriy, fmt.Errorf("function name is missing")
	}

	// for nodeport: get istio ingress port and node ip
	//FIXME use domain if nodeport not configured
	// get node ip
	nodes, err := cfg.KubeClientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return endpoints, authoriy, err
	}

	numNodes := 0
	for _, n := range nodes.Items {
		if numNodes >= defaultNumNodeAddr {
			break
		}

		st := n.Status
		if st.Conditions[len(st.Conditions)-1].Type != corev1.NodeReady {
			glog.Infof("wrong status %v", st.Conditions[len(st.Conditions)].Type)
			continue
		}
		for _, i := range st.Addresses {
			if i.Type == corev1.NodeExternalIP || i.Type == corev1.NodeHostName {
				nodeAddresses = append(nodeAddresses, i.Address)
				numNodes++
			}
		}
	}

	// get istio ingress nodeport
	svc, err := cfg.KubeClientset.CoreV1().Services(defaultIstioNamespace).Get(defaultIstioGatewaySvc, metav1.GetOptions{})
	if err != nil {
		return endpoints, authoriy, err
	}
	// construct endpoint
	for _, p := range svc.Spec.Ports {
		if strings.ToLower(p.Name) == "http2" || strings.ToLower(p.Name) == "https" {
			port := strconv.FormatInt(int64(p.NodePort), 10)
			ep := []string{}
			for _, addr := range nodeAddresses {
				if len(addr) != 0 {
					e := addr + ":" + port
					ep = append(ep, e)
				}
			}
			endpoints = append(endpoints, model.Endpoint{
				Endpoint: ep,
				Protocol: p.Name,
			})
		}
	}
	glog.Infof("endpoints %v", endpoints)
	// get service domain
	servingSvc, err := cfg.ServingClientset.ServingV1alpha1().Services(namespace).Get(funcName, metav1.GetOptions{})
	if err != nil {
		return endpoints, authoriy, err
	}
	authoriy = servingSvc.Status.Domain
	return endpoints, authoriy, nil
}
