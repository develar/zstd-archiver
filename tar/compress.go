package tarutil

import (
	"github.com/develar/go-fs-util"
	"bufio"
	"archive/tar"
	"path/filepath"
	"os"
	"fmt"
	"strings"
	"io"
	"github.com/develar/errors"
	"github.com/DataDog/zstd"
)

func Compress(source string, output string, blockSize int, isSystemIndependentFileInfo bool) (error) {
	outputFile, err := fsutil.CreateFile(output)
	if err != nil {
		return errors.WithStack(err)
	}

	defer outputFile.Close()

	// 22 doesn't produce significant size difference, but took ~1.5x more time
	zstdWriter := zstd.NewWriterLevel(outputFile, 19)
	// without buffer, compression is not good.
	bufferWriter := bufio.NewWriterSize(zstdWriter, blockSize)
	tarWriter := tar.NewWriter(bufferWriter)
	copyBuffer := make([]byte, 32 * 1024)

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking to %s: %v", path, err)
		}

		var header *tar.Header
		if isSystemIndependentFileInfo {
			header, err = SystemIndependentFileInfoHeader(info, path)
		} else {
			header, err = tar.FileInfoHeader(info, path)
		}

		if err != nil {
			return fmt.Errorf("%s: making header: %v", path, err)
		}
		header.Name = strings.TrimPrefix(path, source)
		if info.IsDir() {
			header.Name += "/"
		}

		err = tarWriter.WriteHeader(header)
		if err != nil {
			return fmt.Errorf("%s: writing header: %v", path, err)
		}

		if info.IsDir() {
			return nil
		}

		if header.Typeflag == tar.TypeReg {
			file, err := os.Open(path)
			if err != nil {
				return fmt.Errorf("%s: open: %v", path, err)
			}
			defer file.Close()
			_, err = io.CopyBuffer(tarWriter, file, copyBuffer)
			if err != nil {
				return fmt.Errorf("%s: copying contents: %v", path, err)
			}
		}
		return nil
	})

	err = tarWriter.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	err = bufferWriter.Flush()
	if err != nil {
		return errors.WithStack(err)
	}

	err = zstdWriter.Close()
	if err != nil {
		return errors.WithStack(err)
	}

	return errors.WithStack(err)
}

func SystemIndependentFileInfoHeader(fileInfo os.FileInfo, link string) (*tar.Header, error) {
	if fileInfo == nil {
		return nil, errors.New("archive/tar: FileInfo is nil")
	}

	fileMode := fileInfo.Mode()
	h := &tar.Header{
		Name: fileInfo.Name(),
		Mode: int64(fileMode.Perm()), // or'd with c_IS* constants later
	}

	switch {
	case fileMode.IsRegular():
		h.Typeflag = tar.TypeReg
		h.Size = fileInfo.Size()
	case fileInfo.IsDir():
		h.Typeflag = tar.TypeDir
		h.Name += "/"
	case fileMode&os.ModeSymlink != 0:
		h.Typeflag = tar.TypeSymlink
		h.Linkname = link
	case fileMode&os.ModeDevice != 0:
		if fileMode&os.ModeCharDevice != 0 {
			h.Typeflag = tar.TypeChar
		} else {
			h.Typeflag = tar.TypeBlock
		}
	case fileMode&os.ModeNamedPipe != 0:
		h.Typeflag = tar.TypeFifo
	case fileMode&os.ModeSocket != 0:
		return nil, fmt.Errorf("archive/tar: sockets not supported")
	default:
		return nil, fmt.Errorf("archive/tar: unknown file mode %v", fileMode)
	}

	if sys, ok := fileInfo.Sys().(*tar.Header); ok {
		if sys.Typeflag == tar.TypeLink {
			// hard link
			h.Typeflag = tar.TypeLink
			h.Size = 0
			h.Linkname = sys.Linkname
		}
	}
	return h, nil
}