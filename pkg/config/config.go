package config

import (
	"context"
	"encoding/json"
	"fmt"

	"dario.cat/mergo"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere-extensions/upgrade/pkg/utils/download"
)

type Config struct {
	LogLevel                    klog.Level                            `json:"logLevel,omitempty"`
	DownloadOptions             *download.Options                     `json:"downloadOptions,omitempty"`
	ExtensionUpgradeHookConfigs map[string]ExtensionUpgradeHookConfig `json:"extensionUpgradeHookConfigs,omitempty"`
}

type ExtensionUpgradeHookConfig struct {
	Enabled bool `json:"enabled,omitempty"`
	// InstallCrds indicates whether to force installation of CRDs when the extension is first installed.
	InstallCrds bool `json:"installCrds,omitempty"`
	// UpgradeCrds indicates whether to force upgrade of CRDs when the extension version is upgraded. It is invalid when only the extension value is updated.
	UpgradeCrds bool `json:"upgradeCrds,omitempty"`
	// MergeValues ​​indicates whether to merge values ​​when the extension version is upgraded. It is invalid when only the extension value is updated.
	MergeValues bool `json:"mergeValues,omitempty"`
	// FailurePolicy indicates the policy to use when an error occurs during the upgrade process.
	FailurePolicy FailurePolicy `json:"failurePolicy,omitempty"`
	// DynamicOptions contains dynamic options for the extension.
	DynamicOptions DynamicOptions `json:"dynamicOptions,omitempty"`
}

type DynamicOptions map[string]interface{}

type FailurePolicy int

const (
	// IgnoreError indicates that the upgrade should continue even if an error occurs.
	IgnoreError FailurePolicy = iota
	// FailOnError indicates that the upgrade should fail if any error occurs.
	FailOnError
)

func defaultConfig() *Config {
	return &Config{
		LogLevel: klog.Level(0),
		ExtensionUpgradeHookConfigs: map[string]ExtensionUpgradeHookConfig{
			"whizard-monitoring": {
				Enabled:       true,
				InstallCrds:   true,
				UpgradeCrds:   true,
				MergeValues:   false,
				FailurePolicy: IgnoreError,
			},
			"whizard-monitoring-pro": {
				Enabled:       true,
				InstallCrds:   true,
				UpgradeCrds:   true,
				MergeValues:   false,
				FailurePolicy: IgnoreError,
			},
			"whizard-alerting": {
				Enabled:       true,
				InstallCrds:   true,
				UpgradeCrds:   true,
				MergeValues:   false,
				FailurePolicy: IgnoreError,
			},
			"network": {
				Enabled:       true,
				InstallCrds:   true,
				UpgradeCrds:   true,
				MergeValues:   false,
				FailurePolicy: IgnoreError,
			},
		},
	}
}

func NewConfig(ctx context.Context, client runtimeclient.Client) (*Config, error) {
	defaultCfg := defaultConfig()
	cfg := &Config{}

	if client != nil {
		cm := &corev1.ConfigMap{}

		namespacedName := types.NamespacedName{
			Namespace: GetHookEnvConfigCMNamespace(),
			Name:      GetHookEnvConfigCMName(),
		}

		if err := client.Get(ctx, namespacedName, cm); err != nil {
			if errors.IsNotFound(err) {
				klog.Infof("configmap %s not found, using default config", namespacedName.String())
				return defaultCfg, nil
			}
			klog.Errorf("failed to get configmap %s: %v, using default config", namespacedName.String(), err)
			return defaultCfg, err
		}
		if _, ok := cm.Data[GetHookEnvConfigCMKey()]; !ok {
			klog.Errorf("configmap key not found, using default config")
			return defaultCfg, fmt.Errorf("configmap key: %s not found", GetHookEnvConfigCMKey())
		}

		if err := json.Unmarshal([]byte(cm.Data[GetHookEnvConfigCMKey()]), cfg); err != nil {
			klog.Errorf("failed to unmarshal configmap data: %v, using default config", err)
			return defaultCfg, err
		}

		if err := mergo.Merge(cfg, defaultCfg, mergo.WithOverride); err != nil {
			klog.Errorf("failed to merge config: %v, using default config", err)
			return defaultCfg, err
		}
	}
	klog.V(3).Infof("config: %v", cfg)

	return cfg, nil
}
