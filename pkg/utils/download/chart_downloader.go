package download

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

type Options struct {
	GlobalRegistryUrl string                 `json:"globalRegistryUrl" yaml:"globalRegistryUrl"`
	FileOptions       *FileDownloaderOptions `json:"file" yaml:"file"`
	HttpOptions       *HttpDownloaderOptions `json:"http" yaml:"http"`
	OCIOptions        *OCIDownloaderOptions  `json:"oci" yaml:"oci"`
}

type Downloader interface {
	Get(uri string) (*bytes.Buffer, error)
	Provides(uri string) bool
}

type ChartDownloader struct {
	globalRegistryUrl url.URL

	defaultDownloader Downloader
	downloader        []Downloader
}

func NewDefaultOptions() *Options {
	return &Options{
		HttpOptions: &HttpDownloaderOptions{
			Timeout:            5,
			InsecureSkipVerify: true,
		},
	}
}

func NewChartDownloader(options *Options) (*ChartDownloader, error) {
	if options == nil {
		return nil, errors.New("fail to load download options. Field `config.download` is nil")
	}
	defaultDownloader := NewFileDownloader(options.FileOptions)
	httpDownloader, err := NewHttpDownloader(options.HttpOptions)
	if err != nil {
		return nil, err
	}
	ociDownloader, err := NewOCIDownloader(options.OCIOptions)
	if err != nil {
		return nil, err
	}
	globalRegistryUrl, err := url.Parse(options.GlobalRegistryUrl)
	if err != nil {
		return nil, err
	}
	return &ChartDownloader{
		globalRegistryUrl: *globalRegistryUrl,

		defaultDownloader: defaultDownloader,
		downloader: []Downloader{
			httpDownloader,
			ociDownloader,
		},
	}, nil
}

func (c *ChartDownloader) Download(uri string) (*bytes.Buffer, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		u.Scheme = "file"
	}
	for _, downloader := range c.downloader {
		if downloader.Provides(u.Scheme) {
			return downloader.Get(uri)
		}
	}
	return c.defaultDownloader.Get(uri)
}

func (c *ChartDownloader) DownloadByNameVersion(chartName, chartVersion string) (*bytes.Buffer, error) {
	scheme := c.globalRegistryUrl.Scheme
	if scheme == "" {
		scheme = "file"
	}
	var subPath string
	if scheme == "oci" {
		subPath = fmt.Sprintf("/%s:%s", chartName, chartVersion)
	} else {
		subPath = fmt.Sprintf("/%s-%s.tgz", chartName, chartVersion)
	}

	chartUri := strings.TrimRight(c.globalRegistryUrl.String(), "/") + subPath

	return c.Download(chartUri)
}
