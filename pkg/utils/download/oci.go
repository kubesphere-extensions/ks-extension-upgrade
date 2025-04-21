package download

import (
	"bytes"
	"fmt"
	"strings"

	"helm.sh/helm/v3/pkg/registry"
)

type OCIDownloaderOptions struct{}

var ociDefaultSchemes = []string{"oci"}

type OCIDownloader struct {
	client  *registry.Client
	Schemes []string
}

func NewOCIDownloader(_ *OCIDownloaderOptions) (*OCIDownloader, error) {
	client, err := registry.NewClient()
	if err != nil {
		return nil, err
	}
	return &OCIDownloader{
		client:  client,
		Schemes: ociDefaultSchemes,
	}, nil
}

func (o *OCIDownloader) Get(uri string) (*bytes.Buffer, error) {
	ref := strings.TrimPrefix(uri, fmt.Sprintf("%s://", registry.OCIScheme))

	var pullOpts []registry.PullOption
	requestingProv := strings.HasSuffix(ref, ".prov")
	if requestingProv {
		ref = strings.TrimSuffix(ref, ".prov")
		pullOpts = append(pullOpts,
			registry.PullOptWithChart(false),
			registry.PullOptWithProv(true))
	}

	result, err := o.client.Pull(ref, pullOpts...)
	if err != nil {
		return nil, err
	}

	if requestingProv {
		return bytes.NewBuffer(result.Prov.Data), nil
	}
	return bytes.NewBuffer(result.Chart.Data), nil
}

func (o *OCIDownloader) Provides(scheme string) bool {
	for _, i := range o.Schemes {
		if scheme == i {
			return true
		}
	}
	return false
}
