package config

import "os"

const (
	HookEnvAction      = "HOOK_ACTION"
	HookEnvClusterRole = "CLUSTER_ROLE"
	HookEnvClusterName = "CLUSTER_NAME"
	HookEnvReleaseName = "RELEASE_NAME"

	ActionInstall   = "install"
	ActionUpgrade   = "upgrade"
	ActionUninstall = "uninstall"
)

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
