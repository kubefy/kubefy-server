package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/golang/glog"

	"github.com/kubefy/kubefy-server/pkg/kfunc"
	"github.com/kubefy/kubefy-server/pkg/model"
)

func getRequest(w http.ResponseWriter, r *http.Request, req interface{}) error {
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, req); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		glog.Warningf("failed to parse: %v", err)
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
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

func Root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Kubefy")
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var (
		req model.CreateUserRequest
		rep model.CreateUserResponse
	)
	if err := getRequest(w, r, &req); err != nil {
		return
	}

	// handle user here

	// response
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
	dockerId := req.DockerId
	namespace := req.UserName
	if err := kfunc.Deploy(namespace, gitUrl, gitRevision, "docker.io", dockerId, funcName); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		glog.Warningf("failed to create functions: %v", err)
		w.WriteHeader(500)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}
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
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		glog.Warningf("failed to get function: %v", err)
		w.WriteHeader(500)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	} else {
		rep.Endpoints = ep
		rep.Authority = authoriy
	}

	sendResponse(w, rep)
}

func DeleteFunction(w http.ResponseWriter, r *http.Request) {
}
