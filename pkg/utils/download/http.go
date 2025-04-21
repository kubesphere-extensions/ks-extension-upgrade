package download

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

var (
	httpDefaultSchemes       = []string{"http", "https"}
	defaultHttpTimeout int64 = 10
)

type HttpDownloaderOptions struct {
	Timeout            int64  `json:"timeout,omitempty" yaml:"timeout,omitempty"`
	CaBundle           string `json:"caBundle,omitempty" yaml:"caBundle,omitempty"`
	InsecureSkipVerify bool   `json:"insecureSkipVerify,omitempty" yaml:"insecureSkipVerify,omitempty"`
}

type HttpDownloader struct {
	timeout            int64
	insecureSkipVerify bool
	defaultClient      *http.Client
	pool               map[string]*http.Client
	certPool           *x509.CertPool
	Schemes            []string
}

func NewHttpDownloader(options *HttpDownloaderOptions) (*HttpDownloader, error) {
	transport := &http.Transport{}
	var timeout = defaultHttpTimeout
	if options != nil && options.Timeout > 0 {
		timeout = options.Timeout
	}
	certPool, err := x509.SystemCertPool()
	if err != nil {
		return nil, err
	}
	if options != nil && options.CaBundle != "" {
		caBundle, err := base64.StdEncoding.DecodeString(options.CaBundle)
		if err != nil {
			return nil, errors.Errorf("can't decode CaBundle: %v", err)
		}
		certPool.AppendCertsFromPEM(caBundle)
	}
	insecureSkipVerify := false
	if options != nil {
		insecureSkipVerify = options.InsecureSkipVerify
	}
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: insecureSkipVerify, RootCAs: certPool}

	return &HttpDownloader{
		timeout:            timeout,
		insecureSkipVerify: insecureSkipVerify,
		defaultClient: &http.Client{
			Transport: transport,
			Timeout:   time.Second * time.Duration(timeout),
		},
		pool:     make(map[string]*http.Client),
		Schemes:  httpDefaultSchemes,
		certPool: certPool,
	}, nil
}

func (h *HttpDownloader) Get(uri string) (*bytes.Buffer, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	if u.User != nil && u.User.String() != "" {
		if passwd, isSet := u.User.Password(); isSet {
			req.SetBasicAuth(u.User.Username(), passwd)
		} else {
			req.SetBasicAuth(u.User.Username(), "")
		}
	}

	var client *http.Client
	cacheKey := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
	if cacheClient, exists := h.pool[cacheKey]; !exists {
		tlsConfig := &tls.Config{}
		if u.Scheme == "http" || h.insecureSkipVerify {
			tlsConfig.InsecureSkipVerify = true
		} else {
			tlsConfig.InsecureSkipVerify = false
			tlsConfig.RootCAs = h.certPool
		}
		cacheClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: tlsConfig,
			},
			Timeout: time.Second * time.Duration(h.timeout),
		}
		h.pool[cacheKey] = cacheClient
		client = cacheClient
	} else {
		client = cacheClient
	}
	if client == nil {
		client = h.defaultClient
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to fetch %s : %s", uri, resp.Status)
	}

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, resp.Body)
	return buf, err
}

func (h *HttpDownloader) Provides(scheme string) bool {
	for _, i := range h.Schemes {
		if scheme == i {
			return true
		}
	}
	return false
}
