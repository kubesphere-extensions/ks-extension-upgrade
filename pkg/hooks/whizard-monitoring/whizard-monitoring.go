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
	"k8s.io/klog/v2"
	kscorev1alpha1 "kubesphere.io/api/core/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
	klog.Infof("whizard-monitoring currentVersion: %s, expectVersion: %s", currentVersion, expectVersion)

	v12 := version.MustParseSemantic("1.2.0-0")
	// Upgrade from < 1.2.0 to >= 1.2.0
	if currentVersion.LessThan(v12) && (expectVersion.GreaterThan(v12) || expectVersion.EqualTo(v12)) {

		// Do not block the upgrade process
		if cfg, err := genWhizardMonitoringSmoothUpgradeConfig(installPlan.Spec.Config); err == nil {
			patch := client.MergeFrom(installPlan.DeepCopy())
			installPlan.Spec.Config = cfg
			if err := hook.client.Patch(ctx, installPlan, patch); err != nil {
				klog.Errorf("failed to patch installPlan: %v", err)
			}
		} else {
			klog.Errorf("failed to generate whizard-monitoring smooth upgrade config: %v", err)
		}
		if err := hook.installWhizardMonitoringProExtension(ctx, installPlan); err != nil {
			klog.Errorf("failed to install whizard-monitoring-pro extension: %v", err)
		}
	}

	return nil
}

type upgradeHook struct {
	client          client.Client
	cfg             *config.ExtensionUpgradeHookConfig
	chartDownloader *download.ChartDownloader
}

func genWhizardMonitoringSmoothUpgradeConfig(config string) (string, error) {
	klog.Info("generate whizard monitoring smooth upgrade config, remove tag from image")

	installPlanValues, err := chartutil.ReadValues([]byte(config))
	if err != nil {
		return "", fmt.Errorf("failed to parse installPlan config: %v", err)
	}

	if prometheusTag, err := installPlanValues.PathValue("kube-prometheus-stack.prometheus.prometheusSpec.image.tag"); err == nil && prometheusTag.(string) == "v2.51.2" {
		delete(installPlanValues.AsMap()["kube-prometheus-stack"].(map[string]interface{})["prometheus"].(map[string]interface{})["prometheusSpec"].(map[string]interface{})["image"].(map[string]interface{}), "tag")
	}

	if prometheusOperaterTag, err := installPlanValues.PathValue("kube-prometheus-stack.prometheusOperator.image.tag"); err == nil && prometheusOperaterTag.(string) == "v0.75.1" {
		delete(installPlanValues.AsMap()["kube-prometheus-stack"].(map[string]interface{})["prometheusOperator"].(map[string]interface{})["image"].(map[string]interface{}), "tag")
	}

	if admissionWebhooksPathTag, err := installPlanValues.PathValue("kube-prometheus-stack.prometheusOperator.admissionWebhooks.patch.image.tag"); err == nil && admissionWebhooksPathTag.(string) == "v20221220-controller-v1.5.1-58-g787ea74b6" {
		delete(installPlanValues.AsMap()["kube-prometheus-stack"].(map[string]interface{})["prometheusOperator"].(map[string]interface{})["admissionWebhooks"].(map[string]interface{})["patch"].(map[string]interface{})["image"].(map[string]interface{}), "tag")
	}

	if prometheusConfigReloaderTag, err := installPlanValues.PathValue("kube-prometheus-stack.prometheusOperator.prometheusConfigReloader.image.tag"); err == nil && prometheusConfigReloaderTag.(string) == "v0.75.1" {
		delete(installPlanValues.AsMap()["kube-prometheus-stack"].(map[string]interface{})["prometheusOperator"].(map[string]interface{})["prometheusConfigReloader"].(map[string]interface{})["image"].(map[string]interface{}), "tag")
	}

	if kubeStateMetricsTag, err := installPlanValues.PathValue("kube-prometheus-stack.kube-state-metrics.image.tag"); err == nil && kubeStateMetricsTag.(string) == "v2.12.0" {
		delete(installPlanValues.AsMap()["kube-prometheus-stack"].(map[string]interface{})["kube-state-metrics"].(map[string]interface{})["image"].(map[string]interface{}), "tag")
	}

	if kubeRBACProxyTag, err := installPlanValues.PathValue("kube-prometheus-stack.kube-state-metrics.kubeRBACProxy.image.tag"); err == nil && kubeRBACProxyTag.(string) == "v0.18.0" {
		delete(installPlanValues.AsMap()["kube-prometheus-stack"].(map[string]interface{})["kube-state-metrics"].(map[string]interface{})["kubeRBACProxy"].(map[string]interface{})["image"].(map[string]interface{}), "tag")
	}

	if nodeExporterTag, err := installPlanValues.PathValue("kube-prometheus-stack.prometheus-node-exporter.image.tag"); err == nil && nodeExporterTag.(string) == "v1.8.1" {
		delete(installPlanValues.AsMap()["kube-prometheus-stack"].(map[string]interface{})["prometheus-node-exporter"].(map[string]interface{})["image"].(map[string]interface{}), "tag")
	}

	if kubeRbacProxyTag, err := installPlanValues.PathValue("kube-prometheus-stack.prometheus-node-exporter.kubeRBACProxy.image.tag"); err == nil && kubeRbacProxyTag.(string) == "v0.18.0" {
		delete(installPlanValues.AsMap()["kube-prometheus-stack"].(map[string]interface{})["prometheus-node-exporter"].(map[string]interface{})["kubeRBACProxy"].(map[string]interface{})["image"].(map[string]interface{}), "tag")
	}

	// subchart  whizard-monitoring-helper has been renamed to wiztelemetry-monitoring-helper in v1.2.0
	if _, err := installPlanValues.PathValue("whizard-monitoring-helper"); err != nil {
		installPlanValues["wiztelemetry-monitoring-helper"] = installPlanValues["whizard-monitoring-helper"]
	}

	return installPlanValues.YAML()
}

func checkWhizardConfig(whizardMonitoringCfg string) (interface{}, error) {
	klog.Info("Check whether whizard is installed and get its config")

	whizardMonitoringValues, err := chartutil.ReadValues([]byte(whizardMonitoringCfg))
	if err != nil {
		return nil, fmt.Errorf("failed to parse whizard monitoring config: %v", err)
	}

	needCreateExtension := true
	if whizardEnabled, err := whizardMonitoringValues.PathValue("whizard.enabled"); err != nil || !whizardEnabled.(bool) {
		needCreateExtension = false
	}
	if whizardAgentProxyEnabled, err := whizardMonitoringValues.PathValue("whizardAgentProxy.enabled"); err != nil || !whizardAgentProxyEnabled.(bool) {
		needCreateExtension = false
	}

	if !needCreateExtension {
		return nil, fmt.Errorf("whizard and whizardAgentProxy are not enabled, skip install whizard-monitoring-pro extension")
	}

	return whizardMonitoringValues.PathValue("whizard-agent-proxy.config.gatewayUrl")
}

func (h *upgradeHook) installWhizardMonitoringProExtension(ctx context.Context, whizardMonitoringInstallPlan *kscorev1alpha1.InstallPlan) error {

	whizardGatewayUrl, err := checkWhizardConfig(whizardMonitoringInstallPlan.Spec.Config)
	if err != nil {
		klog.Errorf("failed to check whizard config: %v", err)
		return err
	}

	klog.Info("whizard and whizardAgentProxy are enabled, install whizard-monitoring-pro extension")
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
		whizardMonitoringProDefaultValues := chartutil.Values(chart.Values)
		whizardMonitoringProDefaultValues["whizard-agent-proxy"] = map[string]interface{}{
			"config": map[string]interface{}{
				"gatewayUrl": whizardGatewayUrl,
			},
		}
		whizardMonitoringProExtensionConfig, err := whizardMonitoringProDefaultValues.YAML()
		if err != nil {
			return fmt.Errorf("failed to marshal whizard monitoring pro extension config: %v", err)
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
				ClusterScheduling: whizardMonitoringInstallPlan.Spec.ClusterScheduling,
				Config:            string(whizardMonitoringProExtensionConfig),
				Extension: kscorev1alpha1.ExtensionRef{
					Name:    "whizard-monitoring-pro",
					Version: whizardMonitoringProExtension.Status.RecommendedVersion,
				},
			},
		}
		if err := h.client.Create(ctx, whizardMonitoringProInstallPlan); err != nil {
			return err
		}
	}

	return nil
}
