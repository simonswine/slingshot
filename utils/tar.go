package utils

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
)

func TarFromFile(fileName string, fileBody []byte, fileMode int64) (io.Reader, error) {
	buf := new(bytes.Buffer)
	tarArchive := tar.NewWriter(buf)

	fileHeader := &tar.Header{
		Name: fileName,
		Mode: fileMode,
		Size: int64(len(fileBody)),
	}

	if err := tarArchive.WriteHeader(fileHeader); err != nil{
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
