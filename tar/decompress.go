package tarutil

import (
	"os"
	"github.com/pkg/errors"
	"io"
	"github.com/DataDog/zstd"
	"archive/tar"
	"path/filepath"
	"fmt"
	"github.com/develar/go-fs-util"
)

func Decompress(input string, output string) (error) {
	err := os.MkdirAll(output, 0777)
	if err != nil {
		return errors.WithStack(err)
	}

	inputFile, err := os.Open(input)
	if err != nil {
		return errors.WithStack(err)
	}

	defer inputFile.Close()

	zstdReader := zstd.NewReader(inputFile)
	defer zstdReader.Close()

	copyBuffer := make([]byte, 32*1024)

	tarReader := tar.NewReader(zstdReader)
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return errors.WithStack(err)
		}

		err = untarFile(tarReader, header, output, copyBuffer)
		if err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

func untarFile(tarReader *tar.Reader, header *tar.Header, outputBase string, copyBuffer []byte) error {
	outputFile := filepath.Join(outputBase, header.Name)

	switch header.Typeflag {
	case tar.TypeDir:
		perm := header.FileInfo().Mode().Perm()
		err := os.MkdirAll(outputFile, perm)
		if err != nil {
			return errors.WithStack(err)
		}

		if perm != 0755 {
			err := os.Chmod(outputFile, perm)
			if err != nil {
				return errors.WithStack(err)
			}
		}
		return nil
	case tar.TypeReg, tar.TypeRegA, tar.TypeChar, tar.TypeBlock, tar.TypeFifo:
		return fsutil.WriteFile(tarReader, outputFile, header.FileInfo(), copyBuffer)
	case tar.TypeSymlink:
		return os.Symlink(outputFile, header.Linkname)
	case tar.TypeLink:
		return os.Link(outputFile, header.Linkname)
	default:
		return fmt.Errorf("%s: unknown type flag: %c", header.Name, header.Typeflag)
	}
}
