package config

import "os"

const (
	HookEnvAction      = "HOOK_ACTION"
	HookEnvClusterRole = "CLUSTER_ROLE"
	HookEnvClusterName = "CLUSTER_NAME"
	HookEnvReleaseName = "RELEASE_NAME"
	HookEnvChartPath   = "CHART_PATH"

	ActionInstall   = "install"
	ActionUpgrade   = "upgrade"
	ActionUninstall = "uninstall"
)

func GetHookEnvChartPath() string {
	return os.Getenv(HookEnvChartPath)
}

func GetHookEnvAction() string {
	return os.Getenv(HookEnvAction)
}

func GetHookEnvClusterRole() string {
	return os.Getenv(HookEnvClusterRole)
}

func GetHookEnvClusterName() string {
	return os.Getenv(HookEnvClusterName)
}

func GetHookEnvReleaseName() string {
	return os.Getenv(HookEnvReleaseName)
}
