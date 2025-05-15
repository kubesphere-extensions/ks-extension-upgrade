package config

import (
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chartutil"

	"github.com/kubesphere-extensions/upgrade/pkg/utils/download"
)

type Config struct {
	DownloadOptions             *download.Options                     `json:"downloadOptions,omitempty"`
	ExtensionUpgradeHookConfigs map[string]ExtensionUpgradeHookConfig `json:"extensionUpgradeHookConfigs,omitempty"`
}

type ExtensionUpgradeHookConfig struct {
	Enabled bool `json:"enabled" yaml:"enabled"`
	// InstallCrds indicates whether to force installation of CRDs when the extension is first installed.
	InstallCrds bool `json:"installCrds,omitempty" yaml:"installCrds,omitempty"`
	// UpgradeCrds indicates whether to force upgrade of CRDs when the extension version is upgraded. It is invalid when only the extension value is updated.
	UpgradeCrds bool `json:"upgradeCrds,omitempty" yaml:"upgradeCrds,omitempty"`
	// MergeValues ​​indicates whether to merge values ​​when the extension version is upgraded. It is invalid when only the extension value is updated.
	MergeValues bool `json:"mergeValues,omitempty" yaml:"mergeValues,omitempty"`
	// FailurePolicy indicates the policy to use when an error occurs during the upgrade process.
	FailurePolicy FailurePolicy `json:"failurePolicy,omitempty" yaml:"failurePolicy,omitempty"`
	// DynamicOptions contains dynamic options for the extension.
	DynamicOptions DynamicOptions `json:"dynamicOptions,omitempty" yaml:"dynamicOptions,omitempty"`
}

type DynamicOptions map[string]interface{}

type FailurePolicy int

const (
	// IgnoreError indicates that the upgrade should continue even if an error occurs.
	IgnoreError FailurePolicy = iota
	// FailOnError indicates that the upgrade should fail if any error occurs.
	FailOnError
)

func NewConfig() *Config {
	return &Config{
		DownloadOptions: download.NewDefaultOptions(),
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

func LoadConfigFromHelmValues(values chartutil.Values) (*ExtensionUpgradeHookConfig, error) {
	if values == nil {
		return nil, nil
	}
	global, err := values.Table("global")
	if err != nil {
		return nil, err
	}
	upgradeConfig, err := global.Table("upgradeConfig")
	if err != nil {
		return nil, err
	}
	body, err := upgradeConfig.YAML()
	if err != nil {
		return nil, err
	}

	currentCfg := &ExtensionUpgradeHookConfig{}
	if err := yaml.Unmarshal([]byte(body), currentCfg); err != nil {
		return nil, err
	}

	return currentCfg, nil
}
