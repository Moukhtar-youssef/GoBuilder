package internal

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func CompressResults(results []buildResult, spinner *Spinner) {
	for _, r := range results {
		if !r.success {
			continue
		}
		var err error
		if r.platform.GOOS == "windows" {
			err = compressZip(r.output)
		} else {
			err = compressTarGz(r.output)
		}
		if err != nil {
			spinner.BufferWarn(
				fmt.Sprintf("Failed to compress %s: %v", filepath.Base(r.output), err),
			)
		}
	}
}

func compressZip(srcPath string) error {
	zipPath := strings.TrimSuffix(srcPath, ".exe") + ".zip"
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	w := zip.NewWriter(zipFile)
	defer w.Close()

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}

	entry, err := w.Create(filepath.Base(srcPath))
	if err != nil {
		src.Close()
		return err
	}

	if _, err := io.Copy(entry, src); err != nil {
		src.Close()
		return err
	}

	src.Close()
	return os.Remove(srcPath)
}

func compressTarGz(srcPath string) error {
	tgzFile, err := os.Create(srcPath + ".tar.gz")
	if err != nil {
		return err
	}
	defer tgzFile.Close()

	gzWriter := gzip.NewWriter(tgzFile)
	tarWriter := tar.NewWriter(gzWriter)

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}

	info, err := src.Stat()
	if err != nil {
		src.Close()
		return err
	}

	if err := tarWriter.WriteHeader(&tar.Header{
		Name: filepath.Base(srcPath),
		Mode: 0o755,
		Size: info.Size(),
	}); err != nil {
		src.Close()
		return err
	}

	if _, err := io.Copy(tarWriter, src); err != nil {
		src.Close()
		return err
	}
	src.Close()

	// Must flush in order: tar → gzip, otherwise gzip footer is missing
	if err := tarWriter.Close(); err != nil {
		return err
	}
	if err := gzWriter.Close(); err != nil {
		return err
	}

	return os.Remove(srcPath)
}
