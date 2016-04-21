package utils

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"errors"
)

type TarObject struct {
	Header *tar.Header
	Body   *[]byte
}

// tar a list of files and directories
func TarListOfObjects(objects []TarObject) (tarData []byte, err error) {

	buf := new(bytes.Buffer)

	tarWriter := tar.NewWriter(buf)

	for _, object := range objects {
		if errWriteHeader := tarWriter.WriteHeader(object.Header); errWriteHeader != nil {
			err = errWriteHeader
			return
		}
		if object.Body != nil {
			if _, errWrite := tarWriter.Write(*object.Body); errWrite != nil {
				err = errWrite
				return
			}
		}
	}
	err = tarWriter.Close()
	tarData = buf.Bytes()
	return
}

func WalkDirToObjects(fullPath string, rootPath string) (objects []TarObject, err error) {
	err = filepath.Walk(
		fullPath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}

			relativePath, err := filepath.Rel(rootPath, path)
			if err != nil {
				return err
			}

			// filter cwd
			if len(relativePath) == 0 || relativePath == "." {
				return nil
			}

			// file path not existing
			if info == nil {
				return nil
			}

			object := TarObject{}
			object.Header, err = tar.FileInfoHeader(info, info.Name())
			if err != nil {
				return err
			}

			object.Header.Name = relativePath
			if info.IsDir() {
				object.Header.Name += "/"
			} else {
				// read file content
				fileBytes, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				object.Body = &fileBytes
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

func MergeTar(tarArray[][]byte) ([]byte, error){

	if len(tarArray) == 0 {
		err := errors.New("No tar found")
		return []byte{}, err
	}

	if len(tarArray) == 1 {
		return tarArray[0], nil
	}

	buf := new(bytes.Buffer)
	tarMerged := tar.NewWriter(buf)

	for _, tarSingle := range tarArray {
		tarReader := tar.NewReader(bytes.NewReader(tarSingle))
		for {
			metaData, err := tarReader.Next()
			if err != nil {
				if err == io.EOF {
					break
				}
				return []byte{}, err
			}

			err = tarMerged.WriteHeader(metaData)
			if err != nil {
				return []byte{}, err
			}

			if metaData.Size > 0 {
				_, err := io.CopyN(tarMerged, tarReader, metaData.Size)
				if err != nil {
					return []byte{}, err
				}
			}

		}

	}

	tarMerged.Close()

	return buf.Bytes(), nil
}
