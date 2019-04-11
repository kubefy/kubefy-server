package main

import (
	"flag"
	"github.com/gorilla/mux"
	"net/http"

	cfg "github.com/kubefy/kubefy-server/pkg/config"
	restcall "github.com/kubefy/kubefy-server/pkg/rest"

	"github.com/golang/glog"

	//metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	serving_clientset "github.com/knative/serving/pkg/client/clientset/versioned"
	rook_clientset "github.com/rook/rook/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	defaultNamespace = "default"
)

var (
	namespace  string
	kubeConfig string
)

func main() {
	flag.StringVar(&namespace, "namespace", defaultNamespace, "The configuration namespace")
	flag.StringVar(&kubeConfig, "kubeconfig", "", "Absolute path to the kubeconfig")
	flag.StringVar(&cfg.BuildTemplate, "build-template", "", "Knative build template")
	flag.StringVar(&cfg.RookCephCluster, "rook-ceph-cluster", "rook-ceph", "Rook Ceph cluster namespace")
	flag.StringVar(&cfg.RookCephObjectStore, "rook-ceph-object-store", "", "Rook Ceph Object Store name")
	flag.Parse()
	flag.Set("logtostderr", "true")

	initClients()
	startServer()
}

func initClients() {
	var config *rest.Config
	var err error
	if len(kubeConfig) > 0 {
		config, err = clientcmd.BuildConfigFromFlags("", kubeConfig)
	} else {
		config, err = rest.InClusterConfig()
	}

	if err != nil {
		glog.Fatal(err.Error())
	}
	// create kube clientset
	cfg.KubeClientset = kubernetes.NewForConfigOrDie(config)
	// create serving clientset
	cfg.ServingClientset = serving_clientset.NewForConfigOrDie(config)
	// create rook clientset
	cfg.RookClientset = rook_clientset.NewForConfigOrDie(config)
}

func startServer() {
	router := mux.NewRouter()

	router.HandleFunc("/", restcall.Root).Methods("GET")
	router.HandleFunc("/users", restcall.CreateUser).Methods("POST")

	router.HandleFunc("/functions", restcall.CreateFunction).Methods("POST")
	router.HandleFunc("/functions", restcall.GetFunction).Methods("GET")
	router.HandleFunc("/functions", restcall.DeleteFunction).Methods("DELETE")

	router.HandleFunc("/storage", restcall.CreateStorage).Methods("POST")
	//	router.HandleFunc("/storage", restcall.DeleteStorage).Methods("DELETE")

	glog.Fatal(http.ListenAndServe(":8888", router))
}
