package core

import (
	"context"
	"fmt"
	"strings"

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
	releaseName     string
	extensionName   string
	cfg             *config.ExtensionUpgradeHookConfig
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

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %s", err)
	}

	chartDownloader, err := download.NewChartDownloader(config.NewConfig().DownloadOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create chartDownloader client: %s", err)
	}
	releaseName := config.GetHookEnvReleaseName()
	extensionName := releaseName
	if strings.HasSuffix(releaseName, "-agent") {
		extensionName = strings.TrimSuffix(releaseName, "-agent")
	}
	c := &CoreHelper{
		releaseName:     releaseName,
		extensionName:   extensionName,
		dynamicClient:   dynamicClient,
		chartDownloader: chartDownloader,
		client:          client,
		scheme:          scheme,
	}

	chart, err := c.loadLocalChartFile(fmt.Sprintf("%s.tgz", c.releaseName), "values.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to load chart: %s", err)
	}
	cfg, err := config.LoadConfigFromHelmValues(chart.Values)
	if err != nil {
		klog.Errorf("failed to load config from helm values: %s", err)

		klog.Info("try to use default config")
		if defaultCfg, ok := config.NewConfig().ExtensionUpgradeHookConfigs[extensionName]; ok {
			cfg = &defaultCfg
		}
	}

	klog.Infof("extension %s upgrade config: %+v", c.extensionName, cfg)
	c.cfg = cfg

	return c, nil
}

func (c *CoreHelper) Run(ctx context.Context) error {

	if c.cfg == nil || !c.cfg.Enabled {
		klog.Info("config not found, skip extension upgrade")
		return nil
	}

	// apply crds
	if config.GetHookEnvAction() == config.ActionInstall && c.cfg.InstallCrds ||
		config.GetHookEnvAction() == config.ActionUpgrade && c.cfg.UpgradeCrds {

		klog.Info("force update of crd before extension installation or upgrade")

		if err := c.applyCRDsFromLocalChart(ctx); err != nil {
			return err
		}

	}

	if !strings.HasSuffix(c.releaseName, "-agent") {
		installPlan := &kscorev1alpha1.InstallPlan{}
		if err := c.client.Get(ctx, runtimeclient.ObjectKey{Name: c.extensionName}, installPlan); err != nil {
			return err
		}
		// merge and patch values
		if config.GetHookEnvAction() == config.ActionUpgrade && installPlan.Spec.Extension.Version != installPlan.Status.Version &&
			c.cfg.MergeValues {

			klog.Info("force merge values before extension version upgrade")

			if err := c.mergeValuesFromExtensionChart(ctx, installPlan); err != nil {
				return err
			}
		}

	}

	return nil
}

func (c *CoreHelper) RunHooks(ctx context.Context) error {

	if c.cfg == nil || !c.cfg.Enabled {
		klog.Info("config not found, skip extension upgrade hook")
		return nil
	}

	if hook, ok := hooks.GetHook(c.extensionName); ok {
		klog.Infof("running hook: %s\n", c.extensionName)
		if err := hook.Run(ctx, c.client, c.cfg); err != nil {
			return fmt.Errorf("failed to run hook: %s", err)
		}
	}
	return nil
}
