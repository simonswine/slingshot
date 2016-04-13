package utils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
)

func UnTar(data []byte, destDir string) error {
	b := bytes.NewBuffer(data)
	return unTarHelper(b, destDir)
}

func UnTarGz(data []byte, destDir string) error {
	b := bytes.NewBuffer(data)
	reader, err := gzip.NewReader(b)
	if err != nil {
		return err
	}

	return unTarHelper(reader, destDir)
}

func unTarHelper(reader io.Reader, destDir string) error {

	tarReader := tar.NewReader(reader)

	for {
		metaData, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		fileName := path.Join(destDir, metaData.Name)

		switch metaData.Typeflag {

		case tar.TypeDir:
			// directories
			err = os.MkdirAll(fileName, os.FileMode(metaData.Mode))
			if err != nil {
				return err
			}

		case tar.TypeReg:
			// regular files
			writer, err := os.Create(fileName)
			if err != nil {
				return err
			}
			io.Copy(writer, tarReader)
			writer.Close()

			// fix file permissions
			err = os.Chmod(fileName, os.FileMode(metaData.Mode))
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf(
				"Unknown file type %c for file %s",
				metaData.Typeflag,
				fileName,
			)
		}
	}
	return nil
}
