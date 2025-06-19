package archiver

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)


func Extract(src, dest string) error {
	if strings.HasSuffix(src, ".zip") {
		return unzip(src, dest)
	}
	if strings.HasSuffix(src, ".tar.gz") || strings.HasSuffix(src, ".tgz") {
		return untar(src, dest)
	}
	return fmt.Errorf("unsupported archive format: %s", src)
}

func unzip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func untar(src, dest string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		target := filepath.Join(dest, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(f, tr); err != nil {
				f.Close()
				return err
			}
			f.Close()
		}
	}
	return nil
}

func FindExecutable(dir, repoName, installName string) (string, error) {
	var foundPath string

	
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		isExecutable := info.Mode()&0111 != 0 || (runtime.GOOS == "windows" && strings.HasSuffix(strings.ToLower(info.Name()), ".exe"))

		if isExecutable {
			baseName := strings.TrimSuffix(info.Name(), ".exe")
			if strings.EqualFold(baseName, installName) || strings.EqualFold(baseName, repoName) {
				foundPath = path
				
				return io.EOF
			}
			
			if foundPath == "" {
				foundPath = path
			}
		}
		return nil
	})

	
	if err != nil && err != io.EOF {
		return "", err
	}

	if foundPath != "" {
		return foundPath, nil
	}

	return "", fmt.Errorf("no executable found in %s", dir)
}
