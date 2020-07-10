package deb

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"path/filepath"

	"github.com/ulikunitz/xz/lzma"
	"github.com/xi2/xz"
)

// 根据文件后缀名，解压文件
func decompression(filename string, r io.Reader) (*tar.Reader, error) {
	var tarReader io.Reader
	var err error
	// See https://zh.wikipedia.org/wiki/Deb
	switch filepath.Ext(filename) {
	case ".gz":
		tarReader, err = gzip.NewReader(r)
	case ".xz":
		tarReader, err = xz.NewReader(r, 0)
	case ".lzma":
		tarReader, err = lzma.NewReader(r)
	case ".bz2":
		tarReader = bzip2.NewReader(r)
	case ".tar":
		tarReader = r
	default:
		return nil, fmt.Errorf("unknown extension")
	}
	if err != nil {
		return nil, fmt.Errorf("unzip reader %w", err)
	}
	return tar.NewReader(tarReader), nil
}
