package download

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChartDownloaderDownload(t *testing.T) {
	chartDownloader, err := NewChartDownloader(NewDefaultOptions())
	assert.Equal(t, err, nil)
	_, err = chartDownloader.Download("../../bin/redis-17.0.1.tgz")
	assert.Equal(t, err, nil)
	_, err = chartDownloader.Download("https://charts.kubesphere.io/test/ks-core-0.6.12.tgz")
	assert.Equal(t, err, nil)
	_, err = chartDownloader.Download("oci://hub.kubesphere.com.cn/kse-extensions/whizard-monitoring:1.0.0-rc.4")
	assert.Equal(t, err, nil)
}
