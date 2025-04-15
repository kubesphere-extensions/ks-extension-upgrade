package config

import "os"

const (
	HookEnvAction          = "HOOK_ACTION"
	HookEnvClusterRole     = "CLUSTER_ROLE"
	HookEnvClusterName     = "CLUSTER_NAME"
	HookEnvInstallTag      = "INSTALL_TAG"
	HookEnvInstallPlanName = "INSTALLPLAN_NAME"

	ActionInstall   = "install"
	ActionUpgrade   = "upgrade"
	ActionUninstall = "uninstall"

	InstallTagExtension = "extension"
	InstallTagAgent     = "agent"

	HookEnvConfigCMName      = "HOOK_CONFIG_CM_NAME"
	HookEnvConfigCMNamespace = "HOOK_CONFIG_CM_NAMESPACE"
	HookEnvConfigCMKey       = "HOOK_CONFIG_CM_KEY"
	defaultConfigCMName      = "extension-upgrade-config"
	defaultConfigCMNamespace = "kubesphere-system"
	defaultConfigCMKey       = "config.yaml"
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
func GetHookEnvInstallTag() string {
	return os.Getenv(HookEnvInstallTag)
}
func GetHookEnvInstallPlanName() string {
	return os.Getenv(HookEnvInstallPlanName)
}

func GetHookEnvConfigCMName() string {
	name := os.Getenv(HookEnvConfigCMName)
	if name == "" {
		return defaultConfigCMName
	}
	return name
}
func GetHookEnvConfigCMNamespace() string {
	ns := os.Getenv(HookEnvConfigCMNamespace)
	if ns == "" {
		return defaultConfigCMNamespace
	}
	return ns
}
func GetHookEnvConfigCMKey() string {
	key := os.Getenv(HookEnvConfigCMKey)
	if key == "" {
		return defaultConfigCMKey
	}
	return key
}
