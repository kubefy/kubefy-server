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

package util

import (
	"github.com/golang/glog"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	cfg "github.com/kubefy/kubefy-server/pkg/config"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetNodeAddresses(defaultNumNodeAddr int) ([]string, error) {
	var (
		nodeAddresses []string
	)
	// for nodeport: get istio ingress port and node ip
	// get node ip
	nodes, err := cfg.KubeClientset.CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nodeAddresses, err
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
	return nodeAddresses, err
}

// CreateS3Client creates a s3 client using endpoint and creds
func CreateS3Client(endpoint string, s3id string, s3key string) *s3.S3 {

	creds := credentials.NewStaticCredentials(s3id, s3key, "")

	awsConfig := aws.NewConfig().
		WithRegion("us-east-1").
		WithCredentials(creds).
		WithEndpoint(endpoint).
		WithS3ForcePathStyle(true).
		WithDisableSSL(true).
		WithMaxRetries(20)

	ses := session.New()

	return s3.New(ses, awsConfig)
}

// CreateBucket creates  bucket using s3 client
func CreateBucket(s3client *s3.S3, bucket string) error {
	_, err := s3client.CreateBucket(&s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	})
	return err
}

// DeleteBucket deletes given bucket using s3 client
func DeleteBucket(s3client *s3.S3, bucket string) error {
	_, err := s3client.DeleteBucket(&s3.DeleteBucketInput{
		Bucket: aws.String(bucket),
	})
	return err
}
