package rollback

// Rollbacker defines the interface for rolling back deployments
type Rollbacker interface {
	Rollback(releaseName string, phase int)
	IsRolledBack(releaseName string, phase int) bool
}
