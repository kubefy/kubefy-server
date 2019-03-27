package model

type CreateUserRequest struct {
	UserName string `json:"userName"`
	Bucket   string `json:"bucket,omitempty"`
	// docker setting
	DockerId       string `json:"dockerId"`
	DockerPassword string `json:"dockerPassword,omitempty"`
	// github setting
	GithubId       string `json:"githubId,omitempty"`
	GithubPassword string `json:"githubPassword,omitempty"`
}

type CreateUserResponse struct {
	UserName    string `json:"userName"`
	Bucket      string `json:"bucket,omitempty"`
	S3Endpoint  string `json:"s3endpoint,omitempty"`
	S3AccessKey string `json:"s3access,omitempty"`
	S3SecretKey string `json:"s3secret,omitempty"`
}

type CreateFunctionRequest struct {
	CreateUserRequest
	FunctionName   string `json:"functionName"`
	GitRepo        string `json:"repo"`
	RepoRevision   string `json:"revision,omitempty"`
	ContainerImage string `json:"image,omitempty"`
}

type CreateFunctionResponse struct {
}

type GetFunctionRequest struct {
	CreateUserRequest
	FunctionName string `json:"functionName"`
}

type GetFunctionResponse struct {
	Endpoints []Endpoint `json:"endpoints"`
	Authority string     `json:"authoriy"`
}

type Endpoint struct {
	Endpoint string `json:"endpoint"`
	Protocol string `json:"protocol"`
}
