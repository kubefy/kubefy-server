package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type CreateUserRequest struct {
	Name   string `json:"name"`
	Bucket string `json:"bucket,omitempty"`
	// docker setting
	DockerId       string `json:"dockerId,omitempty"`
	DockerPassword string `json:"dockerPassword,omitempty"`
	// github setting
	GithubId       string `json:"githubId,omitempty"`
	GithubPassword string `json:"githubPassword,omitempty"`
}

type CreateUserResponse struct {
	Name        string `json:"name"`
	Bucket      string `json:"bucket,omitempty"`
	S3Endpoint  string `json:"s3endpoint,omitempty"`
	S3AccessKey string `json:"s3access,omitempty"`
	S3SecretKey string `json:"s3secret,omitempty"`
}

type CreateFunctionRequest struct {
	Repo     string `json:"repo"`
	Revision string `json:"revision,omitempty"`
	Image    string `json:"image"`
}

type CreateFunctionResponse struct {
	Endpoint  string `json:"endpoint"`
	Authority string `json:"authoriy"`
}

func Root(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Kubefy")
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var (
		req CreateUserRequest
		rep CreateUserResponse
	)
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &req); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}
	// handle user here

	// response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(rep); err != nil {
		panic(err)
	}
}

func CreateFunction(w http.ResponseWriter, r *http.Request) {
	var (
		req CreateFunctionRequest
		rep CreateFunctionResponse
	)
	body, err := ioutil.ReadAll(io.LimitReader(r.Body, 1048576))
	if err != nil {
		panic(err)
	}
	if err := r.Body.Close(); err != nil {
		panic(err)
	}
	if err := json.Unmarshal(body, &req); err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
	}
	// handle function here

	// response
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(rep); err != nil {
		panic(err)
	}
}
