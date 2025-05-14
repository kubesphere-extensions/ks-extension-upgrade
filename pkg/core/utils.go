package core

import (
	"bytes"
	"context"
	"fmt"

	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog/v2"
	kscorev1alpha1 "kubesphere.io/api/core/v1alpha1"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	"github.com/kubesphere-extensions/upgrade/pkg/config"
)

func (c *CoreHelper) LoadChart(ctx context.Context, extensionVersion *kscorev1alpha1.ExtensionVersion) (*chart.Chart, error) {

	var chartBuf *bytes.Buffer
	var err error

	if extensionVersion.Spec.ChartURL != "" {
		if chartBuf, err = c.chartDownloader.Download(extensionVersion.Spec.ChartURL); err != nil {
			return nil, fmt.Errorf("failed to download chart %s: %v", extensionVersion.Spec.ChartURL, err)
		}
	} else if extensionVersion.Spec.ChartDataRef != nil {
		cm := corev1.ConfigMap{}

		if err := c.client.Get(ctx, types.NamespacedName{Name: extensionVersion.Spec.ChartDataRef.Name, Namespace: extensionVersion.Spec.ChartDataRef.Namespace}, &cm); err != nil {
			return nil, fmt.Errorf("failed to get configmap %s: %v", cm.Name, err)
		}
		chartBytes, ok := cm.BinaryData[extensionVersion.Spec.ChartDataRef.Key]
		if !ok {
			return nil, fmt.Errorf("failed to get chart data from configmap %s", cm.Name)
		}
		chartBuf = bytes.NewBuffer(chartBytes)
	}

	chart, err := loader.LoadArchive(chartBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to load chart data: %v", err)
	}
	return chart, nil

}

func (c *CoreHelper) applyCRDsFromLocalChart(ctx context.Context) error {

	var chartFile string
	if config.GetHookEnvInstallTag() == config.InstallTagExtension {
		chartFile = fmt.Sprintf("/tmp/helm-executor/%s.tgz", config.GetHookEnvInstallPlanName())
	} else {
		chartFile = fmt.Sprintf("/tmp/helm-executor/%s-agent.tgz", config.GetHookEnvInstallPlanName())
	}
	chartBuf, err := c.chartDownloader.Download(chartFile)
	if err != nil {
		return err
	}
	chart, err := loader.LoadArchive(chartBuf)
	if err != nil {
		return fmt.Errorf("failed to load chart data: %v", err)
	}

	crdClient := c.dynamicClient.Resource(apiextensionsv1.SchemeGroupVersion.WithResource("customresourcedefinitions"))
	codecs := serializer.NewCodecFactory(c.scheme)
	for _, chartCRD := range chart.CRDObjects() {
		obj, _, err := codecs.UniversalDeserializer().Decode(chartCRD.File.Data, nil, nil)
		if err != nil {
			return fmt.Errorf("failed to decode chart crd: %v", err)
		}
		crd, ok := obj.(*apiextensionsv1.CustomResourceDefinition)
		if !ok {
			continue
		}

		unStr, err := runtime.DefaultUnstructuredConverter.ToUnstructured(obj)
		if err != nil {
			return err
		}
		klog.Infof("applying crd: %s\n", crd.Name)
		if _, err := crdClient.Apply(ctx, crd.Name, &unstructured.Unstructured{Object: unStr}, metav1.ApplyOptions{FieldManager: "kubectl", Force: true}); err != nil {
			return err
		}
	}
	return nil
}

func (c *CoreHelper) mergeValuesFromExtensionChart(ctx context.Context, installPlan *kscorev1alpha1.InstallPlan) error {

	extensionVersion := &kscorev1alpha1.ExtensionVersion{}
	extensionVersionName := installPlan.Spec.Extension.Name + "-" + installPlan.Spec.Extension.Version
	if err := c.client.Get(ctx, runtimeclient.ObjectKey{Name: extensionVersionName}, extensionVersion); err != nil {
		return err
	}

	extensionChart, err := c.LoadChart(ctx, extensionVersion)
	if err != nil {
		return err
	}

	klog.Infof("installPlan values: %s\n", installPlan.Spec.Config)

	installPlanValues := make(map[string]interface{})
	if err := yaml.Unmarshal([]byte(installPlan.Spec.Config), &installPlanValues); err != nil {
		return fmt.Errorf("failed to unmarshal installPlan config: %v", err)
	}

	// Set the dependency to empty to avoid introducing subchart values
	extensionChart.SetDependencies()
	values, err := chartutil.MergeValues(extensionChart, installPlanValues)
	if err != nil {
		return fmt.Errorf("failed to merge values: %v", err)
	}
	mergedValues, err := yaml.Marshal(values)
	if err != nil {
		return fmt.Errorf("failed to marshal merged values: %v", err)
	}

	klog.Infof("merged values: %s\n", mergedValues)
	installPlan.Spec.Config = string(mergedValues)

	err = c.client.Update(ctx, installPlan, &runtimeclient.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to patch installPlan: %v", err)
	}
	return nil
}
