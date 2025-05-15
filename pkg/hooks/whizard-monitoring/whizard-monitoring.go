package whizardmonitoring

import (
	"bytes"
	"context"
	"fmt"

	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/version"
	kscorev1alpha1 "kubesphere.io/api/core/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/kubesphere-extensions/upgrade/pkg/config"
	"github.com/kubesphere-extensions/upgrade/pkg/hooks"
	"github.com/kubesphere-extensions/upgrade/pkg/utils/download"
)

const (
	WhizardMonitoringProExtensionName = "whizard-monitoring-pro"

	extensionHookName = "whizard-monitoring"
)

func init() {
	hooks.RegisterHook(extensionHookName, &WhizardMonitoringHook{})
}

type WhizardMonitoringHook struct{}

func (h *WhizardMonitoringHook) Run(ctx context.Context, cli client.Client, cfg *config.ExtensionUpgradeHookConfig) error {

	hook := &upgradeHook{
		client: cli,
		cfg:    cfg,
	}

	chartDownloader, err := download.NewChartDownloader(download.NewDefaultOptions())
	if err != nil {
		return fmt.Errorf("failed to create chartDownloader client: %v", err)
	}
	hook.chartDownloader = chartDownloader

	installPlan := &kscorev1alpha1.InstallPlan{}
	if err := cli.Get(ctx, types.NamespacedName{Name: extensionHookName}, installPlan); err != nil {
		return fmt.Errorf("failed to get install plan %s: %v", extensionHookName, err)
	}

	expectVersion := version.MustParseSemantic(installPlan.Spec.Extension.Version)
	currentVersion := version.MustParseSemantic(installPlan.Status.Version)

	v12 := version.MustParseSemantic("1.2.0")
	// Upgrade from < 1.2.0 to >= 1.2.0
	if currentVersion.LessThan(v12) && expectVersion.GreaterThan(v12) {
		if err := hook.installWhizardMonitoringProExtension(ctx, installPlan); err != nil {
			return fmt.Errorf("failed to run %s hook: %v", extensionHookName, err)
		}
	}

	return nil
}

type upgradeHook struct {
	client          client.Client
	cfg             *config.ExtensionUpgradeHookConfig
	chartDownloader *download.ChartDownloader
}

func (h *upgradeHook) installWhizardMonitoringProExtension(ctx context.Context, installPlan *kscorev1alpha1.InstallPlan) error {

	installPlanValues, err := chartutil.ReadValues([]byte(installPlan.Spec.Config))
	if err != nil {
		return err
	}

	whizardEnabled, err := installPlanValues.PathValue("whizard.enabled")
	if err != nil {
		return err
	}

	whizardAgentProxyEnabled, err := installPlanValues.PathValue("whizardAgentProxy.enabled")
	if err != nil {
		return err
	}
	if whizardEnabled.(bool) && whizardAgentProxyEnabled.(bool) {
		whizardMonitoringProExtension := &kscorev1alpha1.Extension{}
		if err := h.client.Get(ctx, client.ObjectKey{Name: WhizardMonitoringProExtensionName}, whizardMonitoringProExtension); err != nil {
			return err
		}
		if whizardMonitoringProExtension.Status.State == "" && whizardMonitoringProExtension.Status.RecommendedVersion != "" {

			whizardMonitoringProExtensionVersion := &kscorev1alpha1.ExtensionVersion{}
			if err := h.client.Get(ctx, client.ObjectKey{
				Name: WhizardMonitoringProExtensionName + "-" + whizardMonitoringProExtension.Status.RecommendedVersion,
			}, whizardMonitoringProExtensionVersion); err != nil {
				return err
			}

			var chartBuf *bytes.Buffer

			if whizardMonitoringProExtensionVersion.Spec.ChartURL != "" {
				if chartBuf, err = h.chartDownloader.Download(whizardMonitoringProExtensionVersion.Spec.ChartURL); err != nil {
					return fmt.Errorf("failed to download chart %s: %v", whizardMonitoringProExtensionVersion.Spec.ChartURL, err)
				}
			} else if whizardMonitoringProExtensionVersion.Spec.ChartDataRef != nil {
				cm := corev1.ConfigMap{}

				if err := h.client.Get(ctx, types.NamespacedName{Name: whizardMonitoringProExtensionVersion.Spec.ChartDataRef.Name, Namespace: whizardMonitoringProExtensionVersion.Spec.ChartDataRef.Namespace}, &cm); err != nil {
					return fmt.Errorf("failed to get configmap %s: %v", cm.Name, err)
				}
				chartBytes, ok := cm.BinaryData[whizardMonitoringProExtensionVersion.Spec.ChartDataRef.Key]
				if !ok {
					return fmt.Errorf("failed to get chart data from configmap %s", cm.Name)
				}
				chartBuf = bytes.NewBuffer(chartBytes)
			}

			chart, err := loader.LoadArchive(chartBuf)
			if err != nil {
				return fmt.Errorf("failed to load chart data: %v", err)
			}

			whizardMonitoringProExtensionDefaultValues := chartutil.Values(chart.Values)

			whizardGatewayUrl, err := installPlanValues.PathValue("whizard-agent-proxy.config.gatewayUrl")
			if err != nil {
				return err
			}

			whizardMonitoringProExtensionDefaultValues["whizard"] = map[string]interface{}{
				"agentProxy": map[string]interface{}{
					"config": map[string]interface{}{
						"gatewayUrl": whizardGatewayUrl,
					},
				},
			}
			whizardMonitoringProExtensionConfig, err := yaml.Marshal(whizardMonitoringProExtensionDefaultValues)
			if err != nil {
				return err
			}
			// install whizard-monitoring-pro extension
			whizardMonitoringProInstallPlan := &kscorev1alpha1.InstallPlan{
				ObjectMeta: metav1.ObjectMeta{
					Name: "whizard-monitoring-pro",
					Annotations: map[string]string{
						"kubesphere.io/creator": "wiztelemetry-upgrade",
					},
				},

				Spec: kscorev1alpha1.InstallPlanSpec{
					ClusterScheduling: installPlan.Spec.ClusterScheduling,
					Config:            string(whizardMonitoringProExtensionConfig),
					Extension: kscorev1alpha1.ExtensionRef{
						Name:    "whizard-monitoring-pro",
						Version: whizardMonitoringProExtension.Status.RecommendedVersion,
					},
				},
			}
			if err := h.client.Create(ctx, whizardMonitoringProInstallPlan); err != nil {
				return nil
			}
		}
	}

	return nil
}
