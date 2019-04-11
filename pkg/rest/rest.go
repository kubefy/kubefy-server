package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/glog"
	"github.com/google/uuid"

	"github.com/kubefy/kubefy-server/pkg/kfunc"
	"github.com/kubefy/kubefy-server/pkg/kube"
	"github.com/kubefy/kubefy-server/pkg/model"
	"github.com/kubefy/kubefy-server/pkg/storage"
)

func getRequest(w http.ResponseWriter, r *http.Request, req interface{}) error {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	glog.Infof("body %+v", string(body))
	if err := json.Unmarshal(body, req); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		glog.Warningf("failed to parse: %v", err)
		w.WriteHeader(422)
		msg := "error: " + err.Error()
		if err := json.NewEncoder(w).Encode([]byte(msg)); err != nil {
			panic(err)
		}
	}
	return nil
}

func sendResponse(w http.ResponseWriter, rep interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(rep); err != nil {
		panic(err)
	}
}

func sendError(w http.ResponseWriter, rep interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(422)
	if err := json.NewEncoder(w).Encode(rep); err != nil {
		panic(err)
	}
}

func Root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Kubefy")
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var (
		req model.CreateUserRequest
		rep model.CreateUserResponse
	)
	if err := getRequest(w, r, &req); err != nil {
		glog.Warningf("failed to get request. %+v", err)
		rep.Error = err.Error()
		sendError(w, rep)
		return
	}

	// handle user here
	u, err := uuid.NewRandom()
	if err != nil {
		glog.Warningf("failed to get uuid. %+v", err)
		rep.Error = err.Error()
		sendError(w, rep)
		return
	}
	namespace := "kubefy-" + strings.ToLower(req.UserName) + "-" + u.String() + "-ns"
	glog.Infof("created user %v", namespace)
	if err = kube.CreateNamespace(namespace, req.UserName); err != nil {
		glog.Warningf("failed to create namespace. %+v", err)
		rep.Error = err.Error()
		sendError(w, rep)
		return
	}

	dockerSecretName := "docker"
	githubSecretName := "github"
	if len(req.DockerId) > 0 && len(req.DockerPassword) > 0 {
		if err = kube.CreateSecret(namespace, req.UserName, dockerSecretName, req.DockerId, req.DockerPassword); err != nil {
			glog.Warningf("failed to create docker secret. %+v", err)
			rep.Error = err.Error()
			sendError(w, rep)
			return
		}
	}
	if len(req.GithubId) > 0 && len(req.GithubPassword) > 0 {
		if err = kube.CreateSecret(namespace, req.UserName, githubSecretName, req.GithubId, req.GithubPassword); err != nil {
			glog.Warningf("failed to create github secret. %+v", err)
			rep.Error = err.Error()
			sendError(w, rep)
			return
		}
	}

	// response
	rep.UserName = namespace
	sendResponse(w, rep)
}

func CreateFunction(w http.ResponseWriter, r *http.Request) {
	var (
		req model.CreateFunctionRequest
		rep model.CreateFunctionResponse
	)
	if err := getRequest(w, r, &req); err != nil {
		return
	}
	gitUrl := req.GitRepo
	gitRevision := req.RepoRevision
	funcName := req.FunctionName
	image := req.ContainerImage
	namespace := req.UserName
	if len(gitUrl) > 0 {
		if err := kfunc.DeploySrc2Svc(namespace, gitUrl, gitRevision, image, funcName); err != nil {
			rep.Error = err.Error()
			sendError(w, rep)
			return
		}
	} else {
		if len(image) > 0 {
			if err := kfunc.DeployImg2Svc(namespace, image, funcName); err != nil {
				glog.Warningf("failed to create functions: %v", err)
				rep.Error = err.Error()
				sendError(w, rep)
				return
			}
		}
	}
	glog.Infof("created function %v", req.FunctionName)
	sendResponse(w, rep)
}

func GetFunction(w http.ResponseWriter, r *http.Request) {
	var (
		req model.GetFunctionRequest
		rep model.GetFunctionResponse
	)
	if err := getRequest(w, r, &req); err != nil {
		return
	}
	funcName := req.FunctionName
	namespace := req.UserName
	if ep, authoriy, err := kfunc.View(namespace, funcName); err != nil {
		glog.Warningf("failed to get function: %v", err)
		rep.Error = err.Error()
		sendError(w, rep)
		return
	} else {
		rep.Endpoints = ep
		rep.Authority = authoriy
	}

	sendResponse(w, rep)
}

func DeleteFunction(w http.ResponseWriter, r *http.Request) {
}

func CreateStorage(w http.ResponseWriter, r *http.Request) {
	var (
		req model.CreateStorageRequest
		rep model.CreateStorageResponse
	)
	if err := getRequest(w, r, &req); err != nil {
		return
	}

	userName := req.UserName
	if bucket, s3id, s3key, endpoint, err := storage.CreateStorage(userName); err != nil {
		rep.Error = err.Error()
		sendError(w, rep)
		return
	} else {
		glog.Infof("created bucket %v", bucket)
		rep.Bucket = bucket
		rep.S3Endpoint = endpoint
		rep.S3AccessKey = s3id
		rep.S3SecretKey = s3key
	}
	sendResponse(w, rep)
}
