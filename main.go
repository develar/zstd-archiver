package main

import (
	"path/filepath"
	"archive/tar"
	"os"
	"github.com/develar/go-fs-util"
	"github.com/develar/errors"
	"fmt"
	"strings"
	"io"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
)

func main() {
	app := kingpin.New("app-builder", "app-builder").Version("0.0.1")
	ConfigureCompressCommand(app)

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
}

func ConfigureCompressCommand(app *kingpin.Application) {
	command := app.Command("compress", "Compress")
	source := command.Flag("source", "Source directory.").Short('s').Required().String()
	output := command.Flag("output", "Output file.").Short('o').Required().String()
	command.Action(func(context *kingpin.ParseContext) error {
		return Compress(*source, *output)
	})
}

func Compress(source string, output string) (error) {
	outputFile, err := fsutil.CreateFile(output)
	if err != nil {
		return errors.WithStack(err)
	}

	tarWriter := tar.NewWriter(outputFile)
	defer tarWriter.Close()

	copyBuffer := make([]byte, 32 * 1024)

	return filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("error walking to %s: %v", path, err)
		}

		header, err := tar.FileInfoHeader(info, path)
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
}
