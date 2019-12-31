package deb

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/blakesmith/ar"
	"github.com/ulikunitz/xz"
	"github.com/ulikunitz/xz/lzma"
)

// Reader Deb包读取，类似tar的API
type Reader struct {
	arReader  *ar.Reader
	tarDir    string
	tarReader *tar.Reader
	body      io.Reader
}

// NewReader 解析读取deb包，类似tar的操作
func NewReader(r io.Reader) *Reader {
	return &Reader{arReader: ar.NewReader(r)}
}

// Next 类似 tar.Reader.Next
func (deb *Reader) Next() (*tar.Header, error) {
	if deb.tarReader != nil {
		header, err := deb.tarReader.Next()
		if err == nil {
			deb.body = deb.tarReader
			header.Name = filepath.Join(deb.tarDir, header.Name)
			return header, nil
		}
		if err != nil && err != io.EOF {
			return nil, fmt.Errorf("read tar %w", err)
		}
		deb.tarReader = nil
	}
	header, err := deb.arReader.Next()
	if err == io.EOF {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("ar read %w", err)
	}
	switch strings.SplitN(header.Name, ".", 2)[0] {
	case "debian-binary":
		b, err := ioutil.ReadAll(deb.arReader)
		if err != nil {
			return nil, fmt.Errorf("read debian binary %w", err)
		}
		if string(b) != "2.0\n" {
			return nil, fmt.Errorf("unknown version %v", string(b))
		}
		return deb.Next()
	case "control":
		deb.tarDir = "DEBIAN"
	case "data":
		deb.tarDir = ""
	}
	tr, err := decompression(header.Name, deb.arReader)
	if err != nil {
		return nil, fmt.Errorf("decompression control %w", err)
	}
	deb.tarReader = tr
	return deb.Next()
}

func (deb *Reader) Read(b []byte) (int, error) {
	return deb.body.Read(b)
}

// 根据文件后缀名，解压文件
func decompression(filename string, r io.Reader) (*tar.Reader, error) {
	var tarReader io.Reader
	var err error
	// See https://zh.wikipedia.org/wiki/Deb
	switch filepath.Ext(filename) {
	case ".gz":
		tarReader, err = gzip.NewReader(r)
	case ".xz":
		tarReader, err = xz.NewReader(r)
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
