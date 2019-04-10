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

package kube

import (
	cfg "github.com/kubefy/kubefy-server/pkg/config"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateNamespace creates a kube namespace
func CreateNamespace(namespace, label string) error {
	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
			Labels: map[string]string{
				"kubefy.io/username": label,
			},
		},
	}

	_, err := cfg.KubeClientset.CoreV1().Namespaces().Create(ns)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

// CreateSecret creates a kube secret
func CreateSecret(namespace, label, secretName, key, value string) error {
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
			Labels: map[string]string{
				"kubefy.io/username": label,
			},
		},
		Data: map[string][]byte{
			key: []byte(value),
		},
	}
	_, err := cfg.KubeClientset.CoreV1().Secrets(namespace).Create(secret)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}
