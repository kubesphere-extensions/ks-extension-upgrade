package devops

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/version"
	"k8s.io/klog/v2"
	kscorev1alpha1 "kubesphere.io/api/core/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere-extensions/upgrade/pkg/config"
	"github.com/kubesphere-extensions/upgrade/pkg/hooks"
)

const (
	extensionName = "devops"
)

func init() {
	hooks.RegisterHook(extensionName, &Hook{})
}

type Hook struct{}

func (h *Hook) Run(ctx context.Context, c client.Client, _ *config.ExtensionUpgradeHookConfig) error {
	installPlan := &kscorev1alpha1.InstallPlan{}
	if err := c.Get(ctx, types.NamespacedName{Name: extensionName}, installPlan); err != nil {
		return fmt.Errorf("failed to get install plan %s: %v", extensionName, err)
	}

	expectVersion := version.MustParseSemantic(installPlan.Spec.Extension.Version)
	currentVersion := version.MustParseSemantic(installPlan.Status.Version)
	klog.Infof("devops extension currentVersion: %s, expectVersion: %s", currentVersion, expectVersion)

	v124 := version.MustParseSemantic("1.2.4-0")
	// Upgrade from < 1.2.4 to >= 1.2.4
	if currentVersion.LessThan(v124) && (expectVersion.GreaterThan(v124) || expectVersion.EqualTo(v124)) {
		return h.fixConflictResourceMetadata(ctx, c)
	}
	return nil
}

func (h *Hook) fixConflictResourceMetadata(ctx context.Context, c client.Client) error {
	resourceTypes := []struct {
		kind  string
		names []string
	}{
		{"GlobalRole", []string{"devops-anonymous", "devops-authenticated"}},
		{"GlobalRoleBinding", []string{"devops-anonymous", "devops-authenticated"}},
		{"RoleTemplate", []string{"workspace-view-devops", "workspace-create-devops", "workspace-manage-devops"}},
	}

	for _, resource := range resourceTypes {
		for _, roleName := range resource.names {
			u := &unstructured.Unstructured{}
			u.SetGroupVersionKind(schema.GroupVersionKind{
				Group:   "iam.kubesphere.io",
				Version: "v1beta1",
				Kind:    resource.kind,
			})

			key := client.ObjectKey{Name: roleName}
			if err := c.Get(ctx, key, u); err != nil {
				if errors.IsNotFound(err) {
					continue
				}
				return err
			}

			annotations := u.GetAnnotations()
			annotations["meta.helm.sh/release-name"] = "devops"
			u.SetAnnotations(annotations)

			if err := c.Update(ctx, u); err != nil {
				return err
			}
		}
	}

	return nil
}
