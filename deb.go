package deb


import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/blakesmith/ar"
	"github.com/mholt/archiver"
)

// Reader Deb包读取，类似tar的API
type Reader struct {
	arReader  *ar.Reader
	tarDir    string
	tarReader archiver.Reader
	body      io.Reader
}

// NewReader 解析读取deb包，类似tar的操作
func NewReader(r io.Reader) *Reader {
	return &Reader{arReader: ar.NewReader(r)}
}

// Next 类似 tar.Reader.Next
func (deb *Reader) Next() (*tar.Header, error) {
	if deb.tarReader != nil {
		file, err := deb.tarReader.Read()
		if err == nil {
			deb.body = file
			header, ok := file.Header.(*tar.Header)
			if !ok {
				return nil, fmt.Errorf("unknown header")
			}
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
	if header.Name == "debian-binary" {
		b, err := ioutil.ReadAll(deb.arReader)
		if err != nil {
			return nil, fmt.Errorf("read debian binary %w", err)
		}
		if string(b) != "2.0\n" {
			return nil, fmt.Errorf("unknown version %v", string(b))
		}
		return deb.Next()
	}
	tr, err := decompression(header.Name, deb.arReader)
	if err != nil {
		return nil, fmt.Errorf("decompression control %w", err)
	}
	deb.tarDir = strings.SplitN(header.Name, ".", 2)[0]
	deb.tarReader = tr
	return deb.Next()
}

func (deb *Reader) Read(b []byte) (int, error) {
	return deb.body.Read(b)
}

// 根据文件后缀名，解压文件
func decompression(filename string, r io.Reader) (archiver.Reader, error) {
	type Archiver interface {
		archiver.Reader
		archiver.Archiver
	}
	archivers := []Archiver{
		archiver.NewTar(), archiver.NewTarGz(), archiver.NewTarXz(), archiver.NewTarBz2(),
	}
	for i := range archivers {
		err := archivers[i].CheckExt(filename)
		if err != nil {
			continue
		}
		err = archivers[i].Open(r, 0)
		if err != nil {
			return nil, fmt.Errorf("open archive %w", err)
		}
		return archivers[i], nil
	}
	return nil, fmt.Errorf("unknown extension")
}
