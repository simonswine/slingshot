package utils

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type TarObject struct {
	Header *tar.Header
	Body   *[]byte
}

// tar a list of files and directories
func tarListOfObjects(objects []TarObject) (tarData []byte, err error) {

	buf := new(bytes.Buffer)

	tarWriter := tar.NewWriter(buf)

	for _, object := range objects {
		if errWriteHeader := tarWriter.WriteHeader(object.Header); errWriteHeader != nil {
			err = errWriteHeader
			return
		}
		if _, errWrite := tarWriter.Write(*object.Body); errWrite != nil {
			err = errWrite
			return
		}
	}
	err = tarWriter.Close()
	tarData = buf.Bytes()
	return
}

func walkDirToObjects(fullPath string, rootPath string) (objects []TarObject, err error) {
	err = filepath.Walk(
		fullPath,
		func(path string, info os.FileInfo, err error) error {
			relativePath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return err
			}
			if len(relativePath) == 0 || relativePath == "." {
				return nil
			}

			object := TarObject{
				Header: &tar.Header{
					Name:    relativePath,
					Mode:    int64(info.Mode()),
					ModTime: info.ModTime(),
					Size:    info.Size(),
				},
			}
			if !info.IsDir() {
				object.Header.Typeflag = tar.TypeReg

				// read file content
				fileBytes, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				object.Body = &fileBytes
			} else {
				object.Header.Typeflag = tar.TypeDir
			}

			objects = append(objects, object)
			return nil
		},
	)
	return
}

func TarFromFile(fileName string, fileBody []byte, fileMode int64) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tarArchive := tar.NewWriter(buf)

	fileHeader := &tar.Header{
		Name: fileName,
		Mode: fileMode,
		Size: int64(len(fileBody)),
	}

	if err := tarArchive.WriteHeader(fileHeader); err != nil {
		return nil, err
	}

	if _, err := tarArchive.Write(fileBody); err != nil {
		return nil, err
	}

	err := tarArchive.Close()

	return bytes.NewReader(buf.Bytes()), err
}

func FirstFileFromTar(reader io.Reader) (fileBody []byte, fileName string, err error) {
	tarReader := tar.NewReader(reader)
	for {
		metaData, errReader := tarReader.Next()
		if err != nil {
			if errReader == io.EOF {
				break
			}
			err = errReader
			return
		}

		switch metaData.Typeflag {

		case tar.TypeReg:
			_, err = tarReader.Read(fileBody)
			if err != nil {
				return
			}
			fileName = metaData.Name
			return
		default:
		}
	}
	err = fmt.Errorf("Reached end of tar without finding a regular file")
	return
}
