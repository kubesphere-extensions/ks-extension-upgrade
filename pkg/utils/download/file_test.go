package download

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileDownloaderGet(t *testing.T) {
	d := NewFileDownloader(&FileDownloaderOptions{})
	_, err := d.Get("../../bin/redis-17.0.1.tgz")
	assert.Equal(t, err, nil)
}
