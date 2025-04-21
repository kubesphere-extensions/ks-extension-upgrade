package core

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/klog/v2"
	kscorev1alpha1 "kubesphere.io/api/core/v1alpha1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	restconfig "sigs.k8s.io/controller-runtime/pkg/client/config"

	"github.com/kubesphere-extensions/upgrade/pkg/config"
	"github.com/kubesphere-extensions/upgrade/pkg/hooks"
	_ "github.com/kubesphere-extensions/upgrade/pkg/hooks/whizard-monitoring"
	"github.com/kubesphere-extensions/upgrade/pkg/utils/download"
)

type CoreHelper struct {
	config          *config.Config
	client          runtimeclient.Client
	scheme          *runtime.Scheme
	dynamicClient   *dynamic.DynamicClient
	chartDownloader *download.ChartDownloader
}

func NewCoreHelper() (*CoreHelper, error) {
	restConfig, err := restconfig.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get rest config: %s", err)
	}

	scheme := runtime.NewScheme()
	_ = apiextensionsv1.AddToScheme(scheme)
	_ = kscorev1alpha1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	client, err := runtimeclient.New(restConfig, runtimeclient.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %s", err)
	}

	config, err := config.NewConfig(context.Background(), client)
	if err != nil {
		return nil, errors.Errorf("failed to get config: %s", err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, errors.Errorf("failed to create dynamic client: %s", err)
	}

	if config.DownloadOptions == nil {
		config.DownloadOptions = download.NewDefaultOptions()
	}
	chartDownloader, err := download.NewChartDownloader(config.DownloadOptions)
	if err != nil {
		return nil, errors.Errorf("failed to create chartDownloader client: %s", err)
	}

	return &CoreHelper{
		dynamicClient:   dynamicClient,
		chartDownloader: chartDownloader,
		client:          client,
		config:          config,
		scheme:          scheme,
	}, nil
}

func (c *CoreHelper) Run(ctx context.Context) error {

	name := config.GetHookEnvInstallPlanName()
	if name == "" {
		klog.Info("install plan name is empty")
		return nil
	}

	extensionHookConfig, ok := c.config.ExtensionUpgradeHookConfigs[name]
	if !ok || !extensionHookConfig.Enabled {
		klog.Info("extension hook config not enabled")
		return nil
	}

	// apply crds
	if config.GetHookEnvAction() == config.ActionInstall && extensionHookConfig.InstallCrds ||
		config.GetHookEnvAction() == config.ActionUpgrade && extensionHookConfig.UpgradeCrds {

		klog.Info("force update of crd before extension installation or upgrade")

		if err := c.applyCRDsFromLocalChart(ctx); err != nil {
			return err
		}

	}

	if config.GetHookEnvInstallTag() == config.InstallTagExtension {
		installPlan := &kscorev1alpha1.InstallPlan{}
		if err := c.client.Get(ctx, runtimeclient.ObjectKey{Name: name}, installPlan); err != nil {
			return err
		}
		// merge and patch values
		if config.GetHookEnvAction() == config.ActionUpgrade && installPlan.Spec.Extension.Version != installPlan.Status.Version &&
			extensionHookConfig.MergeValues {

			klog.Info("force merge values before extension version upgrade")

			if err := c.mergeValuesFromExtensionChart(ctx, installPlan); err != nil {
				return err
			}
		}

	}

	return nil
}

func (c *CoreHelper) RunHooks(ctx context.Context) error {

	if hook, ok := hooks.GetHook(config.GetHookEnvInstallPlanName()); ok {
		klog.Infof("running hook: %s\n", config.GetHookEnvInstallPlanName())
		if err := hook.Run(ctx, c.client, c.config); err != nil {
			return fmt.Errorf("failed to run hook: %s", err)
		}
	}
	return nil
}
