package deb

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/myml/ar"
	"github.com/ulikunitz/xz"
)

// Writer Deb包Writer，类似tar的API
type Writer struct {
	tempDir     string
	arWriter    *ar.Writer
	controlFile *os.File
	dataFile    *os.File
	gzipWriter  io.WriteCloser
	tarWriter   *tar.Writer
}

func NewWriter(w io.Writer) *Writer {
	deb := &Writer{arWriter: ar.NewWriter(w)}
	return deb
}

func (w *Writer) WriteHeader(head *tar.Header) error {
	var err error
	if len(w.tempDir) == 0 {
		w.tempDir, err = ioutil.TempDir("", "archive-deb_")
		if err != nil {
			return err
		}
	}
	if head.Name == "." || head.Name == "DEBIAN" {
		return nil
	}
	if strings.HasPrefix(head.Name, "DEBIAN") {
		head.Name = head.Name[len("DEBIAN/"):]
		if w.controlFile == nil {
			w.controlFile, err = os.Create(filepath.Join(w.tempDir, "control.tar.gz"))
			if err != nil {
				return err
			}
			w.gzipWriter = gzip.NewWriter(w.controlFile)
			w.tarWriter = tar.NewWriter(w.gzipWriter)
		}
	} else {
		if w.dataFile == nil {
			w.dataFile, err = os.Create(filepath.Join(w.tempDir, "data.tar.xz"))
			if err != nil {
				return err
			}
			err = w.close()
			if err != nil {
				return err
			}
			w.gzipWriter, err = xz.NewWriter(w.dataFile)
			if err != nil {
				return fmt.Errorf("new xz writer")
			}
			w.tarWriter = tar.NewWriter(w.gzipWriter)
		}
	}
	head.Name = "./" + head.Name
	head.Format = tar.FormatGNU
	return w.tarWriter.WriteHeader(head)
}

func (w *Writer) Write(b []byte) (int, error) {
	return w.tarWriter.Write(b)
}

func (w *Writer) Close() error {
	err := w.close()
	if err != nil {
		return fmt.Errorf("writer close %w", err)
	}
	err = w.arWriter.WriteGlobalHeader()
	if err != nil {
		return fmt.Errorf("writer global header %w", err)
	}
	err = w.arWriter.WriteHeader(&ar.Header{Name: "debian-binary", ModTime: time.Now(), Mode: 0655, Size: 4})
	if err != nil {
		return fmt.Errorf("write debian-binary header")
	}
	_, err = w.arWriter.Write([]byte("2.0\n"))
	if err != nil {
		return fmt.Errorf("write debian-binary data")
	}
	for _, f := range []*os.File{w.controlFile, w.dataFile} {
		err = f.Close()
		if err != nil {
			return fmt.Errorf("close file %w", err)
		}
		f, err := os.Open(f.Name())
		defer f.Close()
		if err != nil {
			return fmt.Errorf("read file %w", err)
		}
		stat, err := f.Stat()
		if err != nil {
			return fmt.Errorf("stat %w", err)
		}
		err = w.arWriter.WriteHeader(&ar.Header{
			Name:    stat.Name(),
			ModTime: stat.ModTime(),
			Mode:    int64(stat.Mode()),
			Size:    stat.Size(),
		})
		if err != nil {
			return fmt.Errorf("write header %w", err)
		}
		_, err = io.Copy(w.arWriter, f)
		if err != nil {
			return fmt.Errorf("write data %w", err)
		}
	}
	return nil
}

func (w *Writer) close() error {
	err := w.tarWriter.Close()
	if err != nil {
		return err
	}
	err = w.gzipWriter.Close()
	if err != nil {
		return err
	}
	return nil
}
