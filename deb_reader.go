package deb

import (
	"archive/tar"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/myml/ar"
)

// ErrNoSupportVersion 不支持的debian包版本，目前只支持2.0
var (
	ErrNoSupportVersion = errors.New("Not support version")
	ErrUnknownExtension = errors.New("Unknown extension name")
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
	header.Name = filepath.Base(header.Name)
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
	default:
		return deb.Next()
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
