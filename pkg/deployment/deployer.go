package deployment

import (
	"context"
	"qtm/pkg/catalog"
	"qtm/pkg/session"
	"qtm/pkg/suite"
)

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
	Deploy(ctx context.Context, app suite.SuiteItem, phase int) DeploymentResult
	SetSessionManager(manager session.SessionManager)
	GetSessionManager() session.SessionManager
	SetCatalogSource(src catalog.CatalogSource)
	GetCatalogSource() catalog.CatalogSource
	SetSuiteSource(src suite.SuiteSource)
	GetSuiteSource() suite.SuiteSource
}

// deployApp is a function to handle the deployment of a single app
func DeployApp(ctx context.Context, d Deployer, app suite.SuiteItem, phase int, results chan<- DeploymentResult) {
	// Check for cancellation before starting deployment
	if ctx.Err() != nil {
		return
	}

	// Perform the actual deployment as part of this instantiation of the deployer
	result := d.Deploy(ctx, app, phase)

	// Add the app to the session if the deployment was successful
	if result.Status == Success {
		d.GetSessionManager().AddApp(app.Name, "1.0.0", phase) // Add the app to the session
	}

	// Send the result to the results channel
	results <- result
}
