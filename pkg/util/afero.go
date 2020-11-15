package util

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/afero/tarfs"
	"github.com/spf13/afero/zipfs"
)

func NewAferoFsByPath(path string) (afero.Fs, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open file")
	}

	info, err := file.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "unable to read stat")
	}

	if info.IsDir() {
		return nil, fmt.Errorf("\"%s\" is a directory", path)
	}

	if strings.HasSuffix(path, ".zip") {
		zipReader, err := zip.NewReader(file, info.Size())
		if err != nil {
			return nil, errors.Wrap(err, "unable to read zip")
		}
		return zipfs.New(zipReader), nil
	}

	if strings.HasSuffix(path, "tar") {
		return tarfs.New(tar.NewReader(file)), nil
	}

	if strings.HasSuffix(path, "tar.gz") {
		gReader, err := gzip.NewReader(file)
		if err != nil {
			return nil, errors.Wrap(err, "unable to read gzip")
		}
		return tarfs.New(tar.NewReader(gReader)), nil
	}

	if strings.HasSuffix(path, ".lua") {
		fs := afero.NewMemMapFs()
		if err := afero.WriteReader(fs, "/"+filepath.Base(path), file); err != nil {
			return nil, errors.Wrap(err, "unable to read "+path)
		}
		return fs, nil
	}

	return nil, errors.New("currently only .zip, .tar, .tar.gz are supported")
}
