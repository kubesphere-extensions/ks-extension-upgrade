package download

import (
	"bytes"
	"os"
)

type FileDownloaderOptions struct{}

var fileDefaultSchemes = []string{"file"}

type FileDownloader struct {
	Schemes []string
}

func NewFileDownloader(_ *FileDownloaderOptions) *FileDownloader {
	return &FileDownloader{
		Schemes: fileDefaultSchemes,
	}
}

func (f *FileDownloader) Get(uri string) (*bytes.Buffer, error) {
	data, err := os.ReadFile(uri)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(data), err
}

func (f *FileDownloader) Provides(scheme string) bool {
	for _, i := range f.Schemes {
		if scheme == i {
			return true
		}
	}
	return false
}
