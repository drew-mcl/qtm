package deployment

type HelmDeployer interface {
	Deploy(appID string, phase int) DeploymentResult
}
