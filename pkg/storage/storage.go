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

package storage

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang/glog"

	cfg "github.com/kubefy/kubefy-server/pkg/config"
	"github.com/kubefy/kubefy-server/pkg/model"
	"github.com/kubefy/kubefy-server/pkg/util"

	rookceph "github.com/rook/rook/pkg/apis/ceph.rook.io/v1"

	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	defaultNumNodeAddr = 3
)

func CreateStorage(userName string) (bucket string, s3id string, s3key string, endpoints []model.Endpoint, err error) {
	var (
		addresses []string
	)

	user := &rookceph.CephObjectStoreUser{
		ObjectMeta: metav1.ObjectMeta{
			Name:      userName,
			Namespace: cfg.RookCephCluster,
			Labels: map[string]string{
				"kubefy.io/username": userName,
			},
		},
		Spec: rookceph.ObjectStoreUserSpec{
			Store:       cfg.RookCephObjectStore,
			DisplayName: "Kubefy user " + userName,
		},
	}
	_, err = cfg.RookClientset.CephV1().CephObjectStoreUsers(cfg.RookCephCluster).Create(user)
	if err != nil && !errors.IsAlreadyExists(err) {
		return
	}
	// watch and get s3 secret
	secretFilter := fmt.Sprintf("rook_object_store=%s,user=%s", cfg.RookCephObjectStore, userName)
	listOpts := metav1.ListOptions{LabelSelector: secretFilter}
	ticker := time.NewTicker(500 * time.Millisecond)
	done := make(chan bool)
	go func() {
		time.Sleep(30 * time.Second)
		done <- true
	}()
	for {
		select {
		case <-done:
			break
		case <-ticker.C:
			secrets, listErr := cfg.KubeClientset.CoreV1().Secrets(cfg.RookCephCluster).List(listOpts)
			if listErr != nil {
				err = listErr
				return
			}

			for _, secret := range secrets.Items {
				if key, ok := secret.Data["AccessKey"]; !ok {
					continue
				} else {
					if sec, ok := secret.Data["SecretKey"]; !ok {
						continue
					} else {
						s3id = string(key)
						s3key = string(sec)
						break
					}
				}
			}
		}
		if len(s3id) != 0 || len(s3key) != 0 {
			ticker.Stop()
			break
		}

	}
	if len(s3id) == 0 || len(s3key) == 0 {
		err = fmt.Errorf("failed to find s3 credentials")
		return
	}
	// get endpoint
	svcFilter := fmt.Sprintf("rook_object_store=%s", cfg.RookCephObjectStore)
	listOpts = metav1.ListOptions{LabelSelector: svcFilter}
	glog.Info(svcFilter)
	svcs, listErr := cfg.KubeClientset.CoreV1().Services(cfg.RookCephCluster).List(listOpts)
	if listErr != nil {
		err = listErr
		return
	}
	for _, svc := range svcs.Items {
		tp := svc.Spec.Type
		switch tp {
		case v1.ServiceTypeClusterIP:
			addresses = append(addresses, svc.Spec.ClusterIP)
		case v1.ServiceTypeNodePort:
			nodeAddresses, getErr := util.GetNodeAddresses(defaultNumNodeAddr)
			if getErr != nil {
				err = getErr
				return
			}
			addresses = nodeAddresses

		case v1.ServiceTypeLoadBalancer:
			for _, addr := range svc.Status.LoadBalancer.Ingress {
				addresses = append(addresses, addr.IP)
			}

		}
		for _, p := range svc.Spec.Ports {
			if strings.ToLower(p.Name) == "http" || strings.ToLower(p.Name) == "https" {
				ep := []string{}

				port := p.Port
				if tp != v1.ServiceTypeClusterIP {
					port = p.NodePort
				}
				port64 := strconv.FormatInt(int64(port), 10)
				for _, addr := range addresses {
					if len(addr) != 0 {
						e := addr + ":" + port64
						ep = append(ep, e)
					}
				}

				endpoints = append(endpoints, model.Endpoint{
					Endpoint: ep,
					Protocol: p.Name,
				})
			}
		}

	}
	if len(endpoints) == 0 {
		err = fmt.Errorf("no cluster ip")
		glog.Infof("%v", err)
		return
	}
	// create bucket
	bucket = userName
	for _, ep := range endpoints {
		for _, addr := range ep.Endpoint {
			endpoint := fmt.Sprintf("%s://%s", ep.Protocol, addr)
			s3client := util.CreateS3Client(endpoint, s3id, s3key)
			if err = util.CreateBucket(s3client, bucket); err != nil {
				glog.Infof("created bucket %s", bucket)
				return
			}
		}
	}
	return
}
