package download

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOCIDownloaderGet(t *testing.T) {
	uri := "oci://hub.kubesphere.com.cn/kse-extensions/whizard-monitoring:1.0.0-rc.4"
	d, err := NewOCIDownloader(&OCIDownloaderOptions{})
	assert.Equal(t, err, nil)
	_, err = d.Get(uri)
	assert.Equal(t, err, nil)
}
