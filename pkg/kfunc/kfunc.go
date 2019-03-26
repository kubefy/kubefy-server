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

	cfg "github.com/kubefy/kubefy-server/pkg/config"

	build_api "github.com/knative/build/pkg/apis/build/v1alpha1"
	serving_api "github.com/knative/serving/pkg/apis/serving/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Deploy(gitUrl, gitRevision, containerUrl, containerUser, funcName string) error {
	if len(gitUrl) == 0 || len(funcName) == 0 || len(containerUser) == 0 {
		return fmt.Errorf("git repo, container registry user name, or function name is missing")
	}

	if len(gitRevision) == 0 {
		gitRevision = "master"
	}
	if len(containerUrl) == 0 {
		containerUrl = "docker.io"
	}

	svc := &serving_api.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: funcName,
		},
		Spec: serving_api.ServiceSpec{
			RunLatest: &serving_api.RunLatestType{
				Configuration: serving_api.ConfigurationSpec{
					Build: &serving_api.RawExtension{
						BuildSpec: &build_api.BuildSpec{
							Source: &build_api.SourceSpec{
								Git: &build_api.GitSourceSpec{
									Url:      gitUrl,
									Revision: gitRevision,
								},
							},
							Template: &build_api.TemplateInstantiationSpec{
								Name: cfg.BuildTemplate,
								Arguments: []build_api.ArgumentSpec{
									build_api.ArgumentSpec{
										Name:  "IMAGE",
										Value: containerUrl + "/" + funcName,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	_, err := cfg.ServingClientset.ServingV1alpha1().Services("defaults").Create(svc)

	return err
}

func View(funcName string) (string, error) {
	if len(funcName) == 0 {
		return "", fmt.Errorf("function name is missing")
	}
	return "succeed", nil
}
