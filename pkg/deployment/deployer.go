package deployment

import "context"

// DeploymentResult represents the result of a deployment attempt
type DeploymentResult struct {
	AppID    string
	Phase    int
	Status   DeploymentStatus
	ErrorMsg string
}

// DeploymentStatus represents the status of a deployment
type DeploymentStatus int

const (
	Pending DeploymentStatus = iota
	Success
	Fail
)

// Deployer defines the interface for deploying applications
type Deployer interface {
	Deploy(ctx context.Context, appName string, appGroup string, phase int) DeploymentResult
}
