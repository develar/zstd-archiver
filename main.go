package main

import (
	"os"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"time"
	"github.com/develar/zstd-archiver/tar"
)

func main() {
	app := kingpin.New("zstd-archiver", "app-builder").Version("0.0.1")
	ConfigureCompressCommand(app)
	ConfigureDeCompressCommand(app)

	_, err := app.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
}

func ConfigureCompressCommand(app *kingpin.Application) {
	command := app.Command("compress", "Compress").Alias("c")
	source := command.Flag("source", "Source directory.").Short('s').Required().String()
	output := command.Flag("output", "Output file.").Short('o').Required().String()
	// 64MB produces the same size, but 16MB produces 200KB more, so, use 32 as default
	blockSize := command.Flag("block-size", "Output file.").Default("32MB").Bytes()
	isSystemIndependentFileInfo := command.Flag("system-independent", "Ignore time, xattr, PAX and other system dependent info").Bool()

	command.Action(func(context *kingpin.ParseContext) error {
		start := time.Now()
		err := tarutil.Compress(*source, *output, int(*blockSize), *isSystemIndependentFileInfo)
		elapsed := time.Since(start)
		log.Printf("Compress took %s", elapsed)
		return err
	})
}

func ConfigureDeCompressCommand(app *kingpin.Application) {
	command := app.Command("decompress", "Decompress").Alias("d")
	source := command.Flag("input", "Input file").Short('i').Required().String()
	output := command.Flag("output", "Output directory.").Short('o').Required().String()

	command.Action(func(context *kingpin.ParseContext) error {
		start := time.Now()
		err := tarutil.Decompress(*source, *output)
		elapsed := time.Since(start)
		log.Printf("Decompress took %s", elapsed)
		return err
	})
}