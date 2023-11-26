package rollback

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

// HelmRollbacker manages the  helm rollback process
type HelmRollbacker struct {
	rolledBack     map[string]map[int]bool
	mu             sync.Mutex
	logger         *zap.Logger
	RolledBackApps map[string]bool
}

func NewHelmRollbacker(logger *zap.Logger) *MockRollbacker {
	return &MockRollbacker{
		rolledBack:     make(map[string]map[int]bool),
		RolledBackApps: make(map[string]bool),
		logger:         logger,
	}
}

func (h *HelmRollbacker) HelmUninstall(releaseName string, phase int) {
	h.logger.Info("Performing rollback", zap.String("release_name", releaseName), zap.Int("phase", phase))
	settings := cli.New()
	actionConfig := new(action.Configuration)

	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), nil); err != nil {
		h.logger.Error("Failed to initialize helm action config", zap.Error(err))
	}

	uninstallAction := action.NewUninstall(actionConfig)
	response, err := uninstallAction.Run(releaseName)
	if err != nil {
		h.logger.Error("Failed to uninstall helm release", zap.Error(err))
	}

	h.logger.Info("Helm uninstall response", zap.Any("response", response))
}

// IsRolledBack checks if a specific app has been rolled back in a specific phase
func (h *HelmRollbacker) IsRolledBack(appID string, phase int) bool {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.logger.Info("Checking for rollback", zap.String("appID", appID), zap.Int("phase", phase))

	if phases, exists := h.rolledBack[appID]; exists {
		return phases[phase]
	}
	return false
}
