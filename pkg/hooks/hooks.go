package hooks

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kubesphere-extensions/upgrade/pkg/config"
)

type Hook interface {
	Run(ctx context.Context, cli client.Client, cfg *config.Config) error
}

var hookRegistry = make(map[string]Hook)

func RegisterHook(name string, hook Hook) {
	if _, exists := hookRegistry[name]; exists {
		panic("hook already registered: " + name)
	}
	hookRegistry[name] = hook
}

func GetHook(name string) (Hook, bool) {
	hook, exists := hookRegistry[name]
	return hook, exists
}
